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
	"net"
	"strconv"

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

func NewFloatingIP() resource.Resource {
	return &FloatingIP{}
}

func (r *FloatingIP) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*warren.Client)

	if !ok {
		resp.Diagnostics.AddError("Floating IP configure error", fmt.Sprintf("Expected *warren.Client, got: %T", req.ProviderData))
		return
	}

	r.client = client
}

func (r *FloatingIP) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FloatingIPModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	extendedCtx := context.WithValue(ctx, ctxWrapDataKey("MethodData"), &floatingIPCreateMethodData{})

	err := r.create(extendedCtx, req, &data)
	if nil != err {
		resp.Diagnostics.AddError("Floating IP create error", err.Error())
		r.createOnErrorCleanup(extendedCtx, req, err)

		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FloatingIP) create(ctx context.Context, req resource.CreateRequest, data *FloatingIPModel) error {
	var (
		err        error
		floatingIP *warren.FloatingIp
	)

	assignedToUUID := data.AssignedTo.ValueString()
	resultData := ctx.Value(ctxWrapDataKey("MethodData")).(*floatingIPCreateMethodData)

	if "" != assignedToUUID {
		floatingIP, err = apis.GetFloatingIPFromAssignedUUID(r.client, assignedToUUID)
		if nil != err && !errors.Is(err, apis.ErrFloatingIPNotFound) {
			return apis.GetFloatingIPErrorFromHttpCallError(err)
		}
	}

	if nil == floatingIP {
		floatingIP, err = r.client.Network.CreateFloatingIp(
			&warren.CreateFloatingIpRequest{ Name: warren.New(data.Name.ValueString()) },
		)
		if nil != err {
			return apis.GetFloatingIPErrorFromHttpCallError(err)
		}

		resultData.FloatingIPID = floatingIP.Id
	}

	floatingIPName := data.Name.ValueString()

	if "" != floatingIP.Name && floatingIPName != floatingIP.Name {
		return fmt.Errorf(
			"Floating IP configuration mismatch between existing instance name and given one: %s - %s",
			floatingIP.Name,
			floatingIPName,
		)
	}

	if "" != floatingIP.AssignedTo && assignedToUUID != floatingIP.AssignedTo {
		if floatingIP.AssignedToResourceType != NetworkAssignedToResourceVM {
			return fmt.Errorf(
				"Floating IP configuration mismatch between existing instance attached resource type: %s = %s",
				floatingIP.AssignedTo,
				floatingIP.AssignedToResourceType,
			)
		}

		return fmt.Errorf(
			"Floating IP configuration mismatch between existing instance attached server UUID and given one: %s - %s",
			floatingIP.AssignedTo,
			assignedToUUID,
		)
	}

	if "" != assignedToUUID && "" == floatingIP.AssignedTo {
		floatingIP, err = r.client.Network.AssignFloatingIp(net.ParseIP(floatingIP.Address), assignedToUUID)
		if nil != err {
			return apis.GetFloatingIPErrorFromHttpCallError(err)
		}
	}

	r.setStateData(ctx, floatingIP, data)

	return nil
}

// createOnErrorCleanup cleans up a failed floating IP creation request
//
// PARAMETERS
// ctx context.Context        Execution context
// req resource.CreateRequest The create request for floating IP creation
// err error                  Error encountered
func (r *FloatingIP) createOnErrorCleanup(ctx context.Context, req resource.CreateRequest, err error) {
	resultData := ctx.Value(ctxWrapDataKey("MethodData")).(*floatingIPCreateMethodData)

	if resultData.FloatingIPID != 0 {
		floatingIP, _ := apis.GetFloatingIPByID(r.client, resultData.FloatingIPID)
		if nil != floatingIP {
			_ = r.client.Network.DeleteFloatingIp(net.ParseIP(floatingIP.Address))
		}
	}
}

func (r *FloatingIP) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FloatingIPModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	floatingIPID, err := strconv.Atoi(data.ID.ValueString())
	if nil != err {
		resp.Diagnostics.AddError("Floating IP delete error", err.Error())
		return
	}

	floatingIP, err := apis.GetFloatingIPByID(r.client, floatingIPID)
	if nil != err {
		if errors.Is(err, apis.ErrFloatingIPNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Floating IP has already been deleted: %s", floatingIPID))
		} else {
			resp.Diagnostics.AddError("Floating IP delete error", err.Error())
		}

		return
	}

	err = r.client.Network.DeleteFloatingIp(net.ParseIP(floatingIP.Address))
	if nil != err {
		resp.Diagnostics.AddError("Floating IP delete error", apis.GetFloatingIPErrorFromHttpCallError(err).Error())
	}
}

func (r *FloatingIP) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	floatingIPID, err := strconv.Atoi(req.ID)
	if nil != err {
		resp.Diagnostics.AddError("Floating IP import error", err.Error())
	}

	floatingIP, err := apis.GetFloatingIPByID(r.client, floatingIPID)
	if nil != err {
		resp.Diagnostics.AddError("Floating IP import error", err.Error())
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	data := FloatingIPModel{}
	r.setStateData(ctx, floatingIP, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FloatingIP) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip"
}

func (r *FloatingIP) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var data FloatingIPModel

	if req.Plan.Raw.IsNull() {
		// Nothing to do for resource instance destruction
		return
	}

	// Read Terraform prior plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AssignedTo.IsNull() {
		data.NetworkUUID = types.StringValue("")
	}

	// Save updated data into Terraform plan
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &data)...)
}

func (r *FloatingIP) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FloatingIPModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	floatingIPID, err := strconv.Atoi(data.ID.ValueString())
	if nil != err {
		resp.Diagnostics.AddError("Floating IP read error", err.Error())
		return
	}

	floatingIP, err := apis.GetFloatingIPByID(r.client, floatingIPID)
	if nil != err {
		if errors.Is(err, apis.ErrFloatingIPNotFound) {
			tflog.Trace(ctx, fmt.Sprintf("Floating IP has been deleted: %s", floatingIPID))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Floating IP read error", err.Error())
		}

		return
	}

	r.setStateData(ctx, floatingIP, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FloatingIP) setStateData(ctx context.Context, floatingIP *warren.FloatingIp, data *FloatingIPModel) error {
	data.AssignedTo = types.StringValue(floatingIP.AssignedTo)
	data.AssignedToPrivateIP = types.StringValue(floatingIP.AssignedToPrivateIp)
	data.AssignedToResourceType = types.StringValue(floatingIP.AssignedToResourceType)
	data.Address = types.StringValue(floatingIP.Address)
	data.BillingAccount = types.Int64Value(int64(floatingIP.BillingAccountId))
	data.CreatedAt = types.StringValue(floatingIP.CreatedAt)
	data.Enabled = types.BoolValue(floatingIP.Enabled)
	data.ID = types.StringValue(strconv.Itoa(floatingIP.Id))
	data.IsIPv6 = types.BoolValue(floatingIP.IsIPv6)
	data.Name = types.StringValue(floatingIP.Name)
	data.Type = types.StringValue(floatingIP.Type)
	data.UpdatedAt = types.StringValue(floatingIP.UpdatedAt)
	data.UserID = types.Int64Value(int64(floatingIP.UserId))
	data.UUID = types.StringValue(floatingIP.Uuid)

	if !data.AssignedTo.IsNull() {
		network, err := apis.GetNetworkFromServerUUID(r.client, data.AssignedTo.ValueString())
		if nil != err {
			return apis.GetNetworkErrorFromHttpCallError(err)
		}

		data.NetworkUUID = types.StringValue(network.Uuid)
	}

	return nil
}

func (r *FloatingIP) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Warren Platform floating IP",

		Attributes: map[string]schema.Attribute{
			"assigned_to": schema.StringAttribute{
				MarkdownDescription: "UUID of the resource the floating IP is assigned to",
				Computed:            true,
				Optional:            true,
			},
			"assigned_to_private_ip": schema.StringAttribute{
				MarkdownDescription: "Private IP of the resource the floating IP is assigned to",
				Computed:             true,
			},
			"assigned_to_resource_type": schema.StringAttribute{
				MarkdownDescription: "Type of the resource the floating IP is assigned to",
				Computed:            true,
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "Floating IP address",
				Computed:            true,
			},
			"billing_account": schema.Int64Attribute{
				MarkdownDescription: "Floating IP billing account ID",
				Computed:            true,
				Optional:            true,
				PlanModifiers:       []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Floating IP created at date and time",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Value if the floating IP is enabled",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Floating IP ID",
				Computed:            true,
			},
			"is_ipv6": schema.BoolAttribute{
				MarkdownDescription: "True if the floating IP is an IPv6 address",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Floating IP name",
				Computed:            true,
				Optional:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Floating IP type",
				Computed:            true,
			},
			"network_uuid": schema.StringAttribute{
				MarkdownDescription: "Network UUID the floating IP is routed to",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Floating IP updated at date and time",
				Computed:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "Floating IP owner's user ID",
				Computed:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "Floating IP UUID",
				Computed:            true,
			},
		},
	}
}

func (r *FloatingIP) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		oldData FloatingIPModel
		newData FloatingIPModel
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

	if !oldData.AssignedTo.Equal(newData.AssignedTo) {
		floatingIPID, err := strconv.Atoi(oldData.ID.ValueString())
		if nil != err {
			resp.Diagnostics.AddError("Floating IP update error", err.Error())
			return
		}

		floatingIP, err := apis.GetFloatingIPByID(r.client, floatingIPID)
		if nil != err {
			resp.Diagnostics.AddError("Floating IP update error", err.Error())
			return
		}

		floatingIPAddress := net.ParseIP(floatingIP.Address)
		newAssignedTo := newData.AssignedTo.ValueString()

		if !oldData.AssignedTo.IsNull() {
			oldAssignedTo := oldData.AssignedTo.ValueString()
			tflog.Trace(ctx, fmt.Sprintf("Floating IP will be detached from resource UUID: %s", oldAssignedTo))

			floatingIP, err = r.client.Network.UnAssignFloatingIp(floatingIPAddress)
			if nil != err {
				resp.Diagnostics.AddError("Floating IP update error", apis.GetFloatingIPErrorFromHttpCallError(err).Error())
				return
			}
		}

		if "" != newAssignedTo {
			tflog.Trace(ctx, fmt.Sprintf("Floating IP will be attached to server UUID: %s", newAssignedTo))

			floatingIP, err = r.client.Network.AssignFloatingIp(floatingIPAddress, newAssignedTo)
			if nil != err {
				resp.Diagnostics.AddError("Floating IP update error", apis.GetFloatingIPErrorFromHttpCallError(err).Error())
				return
			}
		}

		r.setStateData(ctx, floatingIP, &newData)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
}
