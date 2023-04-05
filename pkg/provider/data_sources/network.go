/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package data_sources contains all Terraform data sources supported
package data_sources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	warrenClient "gitlab.com/warrenio/library/go-client/warren"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren"
)

func NewNetwork() datasource.DataSource {
	return &Network{}
}

func (d *Network) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *Network) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data warren.NetworkModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := warren.NetworkReadData(d.client, &data)
	if nil != err {
		resp.Diagnostics.AddError("Network read error", err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *Network) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *Network) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
				Optional:            true,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "Network set as default",
				Computed:            true,
				Optional:            true,
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
