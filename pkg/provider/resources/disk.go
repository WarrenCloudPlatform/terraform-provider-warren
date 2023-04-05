/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package resources contains all Terraform resources supported
package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.com/warrenio/library/go-client/warren"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis"
)

func NewDisk() resource.Resource {
	return &Disk{}
}

func (r *Disk) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*warren.Client)

	if !ok {
		resp.Diagnostics.AddError("Disk configure error", fmt.Sprintf("Expected *warren.Client, got: %T", req.ProviderData))
		return
	}

	r.client = client
}

func (r *Disk) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DiskModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	extendedCtx := context.WithValue(ctx, ctxWrapDataKey("MethodData"), &diskCreateMethodData{})

	err := r.create(extendedCtx, req, &data)
	if nil != err {
		resp.Diagnostics.AddError("Disk create error", err.Error())
		r.createOnErrorCleanup(extendedCtx, req, err)

		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Disk) create(ctx context.Context, req resource.CreateRequest, data *DiskModel) error {
	resultData := ctx.Value(ctxWrapDataKey("MethodData")).(*diskCreateMethodData)

	createReq := &warren.CreateDiskRequest{ SizeGb: warren.New(int(data.SizeInGB.ValueInt64())) }

	if !data.SourceImageUUID.IsNull() {
		createReq.SourceImage = warren.New(data.SourceImageUUID.ValueString())

		if !data.SourceImageType.IsNull() {
			createReq.SourceImageType = warren.New(warren.SourceImageType(data.SourceImageType.ValueString()))
		}
	}

	disk, err := r.client.BlockStorage.CreateDisk(createReq)
	if nil != err {
		return apis.GetVolumeErrorFromHttpCallError(err)
	}

	resultData.DiskUUID = disk.Uuid

	if !data.ServerUUID.IsNull() {
		_, err := r.client.VirtualMachine.AttachDisk(data.ServerUUID.ValueString(), disk.Uuid)
		if nil != err {
			return apis.GetServerErrorFromHttpCallError(err)
		}
	}

	r.setStateData(ctx, disk, data)

	return nil
}

// createOnErrorCleanup cleans up a failed disk creation request
//
// PARAMETERS
// ctx context.Context        Execution context
// req resource.CreateRequest The create request for disk creation
// err error                  Error encountered
func (r *Disk) createOnErrorCleanup(ctx context.Context, req resource.CreateRequest, err error) {
	resultData := ctx.Value(ctxWrapDataKey("MethodData")).(*diskCreateMethodData)

	if resultData.DiskUUID != "" {
		_ = r.client.BlockStorage.DeleteDiskById(resultData.DiskUUID)
	}
}

func (r *Disk) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DiskModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diskUUID := data.UUID.ValueString()

	_, err := r.client.BlockStorage.GetDiskById(diskUUID)
	if nil != err {
		err = apis.GetVolumeErrorFromHttpCallError(err)

		if errors.Is(err, apis.ErrVolumeNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Disk has already been deleted: %s", diskUUID))
		} else {
			resp.Diagnostics.AddError("Disk delete error", err.Error())
		}

		return
	}

	err = r.client.BlockStorage.DeleteDiskById(diskUUID)
	if nil != err {
		resp.Diagnostics.AddError("Disk delete error", apis.GetVolumeErrorFromHttpCallError(err).Error())
	}
}

func (r *Disk) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	disk, err := r.client.BlockStorage.GetDiskById(req.ID)
	if nil != err {
		resp.Diagnostics.AddError("Disk import error", apis.GetVolumeErrorFromHttpCallError(err).Error())
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	data := DiskModel{}
	r.setStateData(ctx, disk, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Disk) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_disk"
}

func (r *Disk) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DiskModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	disk, err := r.client.BlockStorage.GetDiskById(data.UUID.ValueString())
	if nil != err {
		err = apis.GetVolumeErrorFromHttpCallError(err)

		if errors.Is(err, apis.ErrVolumeNotFound) {
			tflog.Trace(ctx, fmt.Sprintf("Disk has been deleted: %s", data.UUID.ValueString()))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Disk read error", err.Error())
		}

		return
	}

	r.setStateData(ctx, disk, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Disk) setStateData(ctx context.Context, disk *warren.Disk, data *DiskModel) error {
	data.BillingAccount = types.Int64Value(int64(disk.BillingAccountId))
	data.CreatedAt = types.StringValue(disk.CreatedAt)
	data.SizeInGB = types.Int64Value(int64(disk.SizeGb))
	data.SourceImageUUID = types.StringValue(disk.SourceImage)
	data.SourceImageType = types.StringValue(disk.SourceImageType)
	data.Status = types.StringValue(disk.Status)
	data.StatusComment = types.StringValue(disk.StatusComment)
	data.UpdatedAt = types.StringValue(disk.UpdatedAt)
	data.UserID = types.Int64Value(int64(disk.UserId))
	data.UUID = types.StringValue(disk.Uuid)

	snapshots := []attr.Value{}

	for _, diskSnapshot := range disk.Snapshots {
		snapshots = append(
			snapshots,
			types.ObjectValueMust(
				DiskSnapshotType,
				map[string]attr.Value{
					"created_at": types.StringValue(diskSnapshot.CreatedAt),
					"disk_uuid":  types.StringValue(diskSnapshot.DiskUuid),
					"size_in_gb": types.Int64Value(int64(diskSnapshot.SizeGb)),
					"uuid":       types.StringValue(diskSnapshot.Uuid),
				},
			),
		)
	}

	data.Snapshots = types.ListValueMust(types.ObjectType{ AttrTypes: DiskSnapshotType }, snapshots)

	server, _ := apis.GetServerFromVolumeUUID(r.client, disk.Uuid)
	if nil != server {
		data.ServerUUID = types.StringValue(server.Uuid)
	}

	return nil
}

func (r *Disk) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Warren Platform disk",

		Attributes: map[string]schema.Attribute{
			"billing_account": schema.Int64Attribute{
				MarkdownDescription: "Disk billing account ID",
				Computed:            true,
				Optional:            true,
				PlanModifiers:       []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Disk created at date and time",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Disk UUID",
				Computed:            true,
			},
			"server_uuid": schema.StringAttribute{
				MarkdownDescription: "Server UUID disk is attached to",
				Optional:            true,
			},
			"size_in_gb": schema.Int64Attribute{
				MarkdownDescription: "Disk size in GB",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"snapshots": schema.ListNestedAttribute{
				MarkdownDescription: "Disk snapshots",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Disk snapshot created at date and time",
							Computed:            true,
						},
						"disk_uuid": schema.StringAttribute{
							MarkdownDescription: "Source disk UUID",
							Computed:            true,
						},
						"size_in_gb": schema.Int64Attribute{
							MarkdownDescription: "Disk snapshot size",
							Computed:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "Disk snapshot UUID",
							Computed:            true,
						},
					},
				},
			},
			"source_image_type": schema.StringAttribute{
				MarkdownDescription: "Disk source image type",
				Computed:            true,
				Optional:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_image_uuid": schema.StringAttribute{
				MarkdownDescription: "Disk source image UUID",
				Computed:            true,
				Optional:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Disk status",
				Computed:            true,
			},
			"status_comment": schema.StringAttribute{
				MarkdownDescription: "Disk status comment",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Disk updated at date and time",
				Computed:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "Disk owner's user ID",
				Computed:            true,
			},
		},
	}
}

func (r *Disk) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		oldData DiskModel
		newData DiskModel
	)

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &newData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !oldData.ServerUUID.Equal(newData.ServerUUID) {
		diskUUID := oldData.UUID.ValueString()
		newServerUUID := newData.ServerUUID.ValueString()

		if !oldData.ServerUUID.IsNull() {
			oldServerUUID := oldData.ServerUUID.ValueString()
			tflog.Trace(ctx, fmt.Sprintf("Disk will be detached from server UUID: %s", oldServerUUID))

			err := r.client.VirtualMachine.DetachDisk(oldServerUUID, diskUUID)
			if nil != err {
				resp.Diagnostics.AddError("Disk update error", apis.GetServerErrorFromHttpCallError(err).Error())
				return
			}
		}

		if "" != newServerUUID {
			tflog.Trace(ctx, fmt.Sprintf("Disk will be attached to server UUID: %s", newServerUUID))

			_, err := r.client.VirtualMachine.AttachDisk(newServerUUID, diskUUID)
			if nil != err {
				resp.Diagnostics.AddError("Disk update error", apis.GetServerErrorFromHttpCallError(err).Error())
				return
			}
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
}
