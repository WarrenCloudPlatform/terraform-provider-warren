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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	warrenClient "gitlab.com/warrenio/library/go-client/warren"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis"
)

func NewNetwork() resource.Resource {
	return &Network{}
}

func (r *Network) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*warrenClient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Network configure error",
			fmt.Sprintf("Expected *warren.Client, got: %T", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *Network) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data warren.NetworkModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	network, err := r.client.Network.CreateNetwork(data.Name.ValueString())
	if nil != err {
		resp.Diagnostics.AddError("Network create error", apis.GetNetworkErrorFromHttpCallError(err).Error())
		return
	}

	warren.NetworkSetStateData(network, &data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Network) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data warren.NetworkModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	networkUUID := data.UUID.ValueString()

	_, err := r.client.Network.GetNetworkByUUID(networkUUID)
	if nil != err {
		err = apis.GetNetworkErrorFromHttpCallError(err)

		if errors.Is(err, apis.ErrNetworkNotFound) {
			tflog.Debug(ctx, fmt.Sprintf("Network has already been deleted: %s", networkUUID))
		} else {
			resp.Diagnostics.AddError("Network delete error", err.Error())
		}

		return
	}

	err = r.client.Network.DeleteNetworkByUUID(networkUUID)
	if nil != err {
		resp.Diagnostics.AddError("Network delete error", apis.GetNetworkErrorFromHttpCallError(err).Error())
	}
}

func (r *Network) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	network, err := r.client.Network.GetNetworkByUUID(req.ID)
	if nil != err {
		resp.Diagnostics.AddError("Network import error", apis.GetNetworkErrorFromHttpCallError(err).Error())
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	data := warren.NetworkModel{}
	warren.NetworkSetStateData(network, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Network) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *Network) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data warren.NetworkModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := warren.NetworkReadData(r.client, &data)
	if nil != err {
		resp.Diagnostics.AddError("Network read error", err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Network) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Warren Platform network",

		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Network created at date and time",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Network UUID",
				Computed:            true,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "Network set as default",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Network name",
				Computed:            true,
				Optional:            true,
			},
			"server_uuids": schema.ListAttribute{
				MarkdownDescription: "Network server UUIDs",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"subnet_ipv4": schema.StringAttribute{
				MarkdownDescription: "Network IPv4 subnet",
				Computed:            true,
			},
			"subnet_ipv6": schema.StringAttribute{
				MarkdownDescription: "Network IPv6 subnet",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Network type",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Network updated at date and time",
				Computed:            true,
			},
			"vlan_id": schema.Int64Attribute{
				MarkdownDescription: "Network VLAN ID",
				Computed:            true,
			},
		},
	}
}

func (r *Network) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		oldData warren.NetworkModel
		newData warren.NetworkModel
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

	if !oldData.Name.Equal(newData.Name) {
		network, err := r.client.Network.ChangeNetworkName(oldData.UUID.ValueString(), warrenClient.New(newData.Name.ValueString()))
		if nil != err {
			resp.Diagnostics.AddError("Network update error", apis.GetNetworkErrorFromHttpCallError(err).Error())
			return
		}

		warren.NetworkSetStateData(network, &newData)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
}
