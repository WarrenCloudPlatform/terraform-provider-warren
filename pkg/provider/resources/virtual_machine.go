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
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

    "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.com/warrenio/library/go-client/warren"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis"
)

func NewVirtualMachine() resource.Resource {
	return &VirtualMachine{}
}

func (r *VirtualMachine) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*warren.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Virtual machine configure error",
			fmt.Sprintf("Expected *warren.Client, got: %T", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *VirtualMachine) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtualMachineModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password := data.Password.ValueString()

	if password == "" {
		var passwordBuilder strings.Builder
		seededRand := rand.New(rand.NewSource(time.Now().Unix()))

		for i := 0; i < (warrenPasswordGeneratedLength - 3); i++ {
			passwordBuilder.WriteRune(rune(32 + seededRand.Intn(94)))
		}

		// Lower-case letter
		passwordBuilder.WriteRune(rune(97 + seededRand.Intn(26)))
		// Upper-case letter
		passwordBuilder.WriteRune(rune(65 + seededRand.Intn(26)))
		// number
		passwordBuilder.WriteRune(rune(48 + seededRand.Intn(10)))

		password = passwordBuilder.String()
	}

	username := data.Username.ValueString()

	if username == "" {
		username = warrenDefaultVMUsername
	}

	createReq := &warren.CreateVirtualMachineRequest{
		Name:            warren.New(data.Name.ValueString()),
		OsName:          warren.New(data.OSName.ValueString()),
		OsVersion:       warren.New(data.OSVersion.ValueString()),
		Disks:           warren.New(int(data.DiskSizeInGB.ValueInt64())),
		VCpu:            warren.New(int(data.VCPU.ValueInt64())),
		Ram:             warren.New(int(data.Memory.ValueInt64())),
		ReservePublicIp: warren.New(data.ReservePublicIP.ValueBool()),
		Backup:          warren.New(data.Backup.ValueBool()),
		Username:        warren.New(username),
		Password:        warren.New(password),
	}

	if !data.CloudInit.IsNull() {
		cloudInit := data.CloudInit.ValueString()
		cloudInit = strings.ReplaceAll(cloudInit, "\r\n", "\n")
		cloudInit = strings.ReplaceAll(cloudInit, "\r", "\n")

		if !strings.HasPrefix(cloudInit, "#cloud-config\n") {
			jsonCloudInit, err := json.Marshal(map[string]interface{}{ "runcmd": strings.Split(cloudInit, "\n\n") })
			if nil != err {
				resp.Diagnostics.AddError("Virtual machine create error", err.Error())
				return
			}

			cloudInit = string(jsonCloudInit)
		}

		createReq.CloudInit = warren.New(cloudInit)
	}

	if !data.PublicKey.IsNull() {
		createReq.PublicKey = warren.New(data.PublicKey.ValueString())
	}

	if !(data.NetworkUUID.IsNull() || data.NetworkUUID.IsUnknown()) {
		createReq.NetworkUuid = warren.New(data.NetworkUUID.ValueString())
	}

	if !data.SourceUUID.IsNull() {
		createReq.SourceUuid = warren.New(data.SourceUUID.ValueString())

		if !data.SourceReplica.IsNull() {
			createReq.SourceReplica = warren.New(data.SourceReplica.ValueString())
		}
	}

	server, err := r.client.VirtualMachine.CreateVirtualMachine(createReq)
	if nil != err {
		resp.Diagnostics.AddError("Virtual machine create error", apis.GetServerErrorFromHttpCallError(err).Error())
		return
	}

	r.setStateData(ctx, server, &data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualMachine) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtualMachineModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverUUID := data.UUID.ValueString()

	_, err := r.client.VirtualMachine.GetByUuid(serverUUID)
	if nil != err {
		err = apis.GetServerErrorFromHttpCallError(err)

		if errors.Is(err, apis.ErrServerNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Virtual machine has already been deleted: %s", serverUUID))
		} else {
			resp.Diagnostics.AddError("Virtual machine delete error", err.Error())
		}

		return
	}

	err = r.client.VirtualMachine.DeleteVm(serverUUID)
	if nil != err {
		resp.Diagnostics.AddError("Virtual machine delete error", apis.GetServerErrorFromHttpCallError(err).Error())
	}
}

func (r *VirtualMachine) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	server, err := r.client.VirtualMachine.GetByUuid(req.ID)
	if nil != err {
		resp.Diagnostics.AddError("Virtual machine import error", apis.GetServerErrorFromHttpCallError(err).Error())
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	data := VirtualMachineModel{}
	r.setStateData(ctx, server, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualMachine) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (r *VirtualMachine) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtualMachineModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.VirtualMachine.GetByUuid(data.UUID.ValueString())
	if nil != err {
		err = apis.GetServerErrorFromHttpCallError(err)

		if errors.Is(err, apis.ErrServerNotFound) {
			tflog.Trace(ctx, fmt.Sprintf("Virtual machine has been deleted: %s", data.UUID.ValueString()))
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Virtual machine read error", err.Error())
		}

		return
	}

	r.setStateData(ctx, server, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualMachine) setStateData(ctx context.Context, server *warren.VirtualMachine, data *VirtualMachineModel) error {
	data.Backup = types.BoolValue(server.Backup)
	data.BillingAccount = types.Int64Value(int64(server.BillingAccount))
	data.CreatedAt = types.StringValue(server.CreatedAt)
	data.Description = types.StringValue(server.Description)
	data.Hostname = types.StringValue(server.Hostname)
	data.Name = types.StringValue(server.Name)
	data.MAC = types.StringValue(server.Mac)
	data.Memory = types.Int64Value(int64(server.Memory))
	data.OSName = types.StringValue(server.OsName)
	data.OSVersion = types.StringValue(server.OsVersion)
	data.PrivateIPv4 = types.StringValue(server.PrivateIPv4)
	data.PublicIPv6 = types.StringValue(server.PublicIPv6)
	data.PublicIPv6 = types.StringValue(server.PublicIPv6)
	data.Status = types.StringValue(server.Status)
	data.UpdatedAt = types.StringValue(server.UpdatedAt)
	data.UserID = types.Int64Value(int64(server.UserId))
	data.Username = types.StringValue(server.Username)
	data.UUID = types.StringValue(server.Uuid)
	data.VCPU = types.Int64Value(int64(server.VCpu))

	storage := []attr.Value{}

	for _, serverStorage := range server.Storage {
		if serverStorage.Primary && data.DiskSizeInGB.IsNull() {
			data.DiskSizeInGB = types.Int64Value(int64(serverStorage.Size))
		}

		storageReplica := []attr.Value{}

		for _, serverStorageReplica := range serverStorage.Replica {
			storageReplica = append(
				storageReplica,
				types.ObjectValueMust(
					VirtualMachineStorageReplicaType,
					map[string]attr.Value{
						"created_at":  types.StringValue(serverStorageReplica.CreatedAt),
						"master_uuid": types.StringValue(serverStorageReplica.MasterUuid),
						"size_in_gb":  types.Int64Value(int64(serverStorageReplica.Size)),
						"type":        types.StringValue(serverStorageReplica.Type),
						"uuid":          types.StringValue(serverStorageReplica.Uuid),
					},
				),
			)
		}

		storage = append(
			storage,
			types.ObjectValueMust(
				VirtualMachineStorageType,
				map[string]attr.Value{
					"created_at": types.StringValue(serverStorage.CreatedAt),
					"name":       types.StringValue(serverStorage.Name),
					"primary":    types.BoolValue(serverStorage.Primary),
					"replica":    types.ListValueMust(
						types.ObjectType{ AttrTypes: VirtualMachineStorageReplicaType },
						storageReplica,
					),
					"size_in_gb": types.Int64Value(int64(serverStorage.Size)),
					"user_id":    types.Int64Value(int64(serverStorage.UserId)),
					"uuid":         types.StringValue(serverStorage.Uuid),
				},
			),
		)
	}

	data.Storage = types.ListValueMust(types.ObjectType{ AttrTypes: VirtualMachineStorageType }, storage)

	network, _ := apis.GetNetworkFromServerUUID(r.client, server.Uuid)
	if nil != network {
		data.NetworkUUID = types.StringValue(network.Uuid)
	}

	return nil
}

func (r *VirtualMachine) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Warren Platform virtual machine",

		Attributes: map[string]schema.Attribute{
			"backup": schema.BoolAttribute{
				MarkdownDescription: "Virtual machine backup value",
				Computed:            true,
				Optional:            true,
				PlanModifiers:       []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"billing_account": schema.Int64Attribute{
				MarkdownDescription: "Virtual machine billing account ID",
				Computed:            true,
				Optional:            true,
				PlanModifiers:       []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"cloud_init": schema.StringAttribute{
				MarkdownDescription: "Virtual machine cloud init configuration",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Virtual machine created at date and time",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Virtual machine description",
				Computed:            true,
			},
			"disk_size_in_gb": schema.Int64Attribute{
				MarkdownDescription: "Virtual machine boot disk size in GB",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Virtual machine hostname",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Virtual machine UUID",
				Computed:            true,
			},
			"mac": schema.StringAttribute{
				MarkdownDescription: "Virtual machine MAC",
				Computed:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: "Virtual machine memory value in MB",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{
					// @TODO add support for update
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Virtual machine name",
				Required:            true,
				PlanModifiers:       []planmodifier.String{
					// @TODO add support for update
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_uuid": schema.StringAttribute{
				MarkdownDescription: "Virtual machine network UUID attached",
				Computed:            true,
				Optional:            true,
			},
			"os_name": schema.StringAttribute{
				MarkdownDescription: "Virtual machine OS image name",
				Required:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"os_version": schema.StringAttribute{
				MarkdownDescription: "Virtual machine OS image version",
				Required:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Virtual machine password for SSH access",
				Sensitive:           true,
				Optional:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_ipv4": schema.StringAttribute{
				MarkdownDescription: "Virtual machine private IPv4",
				Computed:            true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Virtual machine public key for SSH access",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_ipv6": schema.StringAttribute{
				MarkdownDescription: "Virtual machine public public IPv6",
				Computed:            true,
			},
			"reserve_public_ip": schema.BoolAttribute{
				MarkdownDescription: "Virtual machine public IP should be reserved at creation if set",
				Optional:            true,
				PlanModifiers:       []planmodifier.Bool{
					// @TODO add support for update
					boolplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Virtual machine status",
				Computed:            true,
			},
			"storage": schema.ListNestedAttribute{
				MarkdownDescription: "Virtual machine storages",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Virtual machine storage created at date and time",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Virtual machine storage name",
							Computed:            true,
						},
						"primary": schema.BoolAttribute{
							MarkdownDescription: "Virtual machine storage is primary if set",
							Computed:            true,
						},
						"replica": schema.ListNestedAttribute{
							MarkdownDescription: "Virtual machine storage replicas",
							Computed:            true,
							NestedObject:        schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"created_at": schema.StringAttribute{
										MarkdownDescription: "Virtual machine storage replica created at date and time",
										Computed:            true,
									},
									"master_uuid": schema.StringAttribute{
										MarkdownDescription: "Virtual machine storage replica master UUID",
										Computed:            true,
									},
									"size_in_gb": schema.Int64Attribute{
										MarkdownDescription: "Virtual machine storage replica size",
										Computed:            true,
									},
									"type": schema.StringAttribute{
										MarkdownDescription: "Virtual machine storage replica type",
										Computed:            true,
									},
									"uuid": schema.StringAttribute{
										MarkdownDescription: "Virtual machine storage replica UUID",
										Computed:            true,
									},
								},
							},
						},
						"size_in_gb": schema.Int64Attribute{
							MarkdownDescription: "Virtual machine storage size",
							Computed:            true,
						},
						"user_id": schema.Int64Attribute{
							MarkdownDescription: "Virtual machine storage owner's user ID",
							Computed:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "Virtual machine storage UUID",
							Computed:            true,
						},
					},
				},
			},
			"source_uuid": schema.StringAttribute{
				MarkdownDescription: "Virtual machine boot disk source UUID",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_replica": schema.StringAttribute{
				MarkdownDescription: "Virtual machine boot disk source replica",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Virtual machine updated at date and time",
				Computed:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "Virtual machine owner's user ID",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Virtual machine user name for SSH access",
				Computed:            true,
				Optional:            true,
				PlanModifiers:       []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vcpu": schema.Int64Attribute{
				MarkdownDescription: "Virtual machine VCPU value",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{
					// @TODO add support for update
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *VirtualMachine) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Virtual machine update error", "Updating an existing machine is currently not implemented")
}
