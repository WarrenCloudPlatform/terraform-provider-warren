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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"gitlab.com/warrenio/library/go-client/warren"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis"
)

func NewLocation() datasource.DataSource {
	return &Location{}
}

func (d *Location) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*warren.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Location configure error",
			fmt.Sprintf("Expected *warren.Client, got: %T", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *Location) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("display_name"),
			path.MatchRoot("id"),
		),
	}
}

func (d *Location) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LocationModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	locations, err := d.client.Location.ListLocations()
	if nil != err {
		resp.Diagnostics.AddError("Location read error", apis.GetLocationErrorFromHttpCallError(err).Error())
		return
	}

	var isFound bool
	isParameterOnly := data.DisplayName.IsNull() && data.Slug.IsNull()

	for _, location := range *locations {
		if !data.DisplayName.IsNull() && data.DisplayName.Equal(types.StringValue(location.DisplayName)) {
			isFound = true
		}

		if !data.Slug.IsNull() && data.Slug.Equal(types.StringValue(location.Slug)) {
			isFound = true
		}

		if !data.IsDefault.IsNull() {
			if !data.IsDefault.Equal(types.BoolValue(location.IsDefault)) {
				isFound = false
			} else if isParameterOnly {
				isFound = true
			}
		}

		if !data.IsPreferred.IsNull() {
			if !data.IsPreferred.Equal(types.BoolValue(location.IsPreferred)) {
				isFound = false
			} else if isParameterOnly {
				isFound = true
			}
		}

		if !isFound {
			continue
		}

		data.CountryCode = types.StringValue(location.CountryCode)
		data.DisplayName = types.StringValue(location.DisplayName)
		data.IsDefault = types.BoolValue(location.IsDefault)
		data.IsPreferred = types.BoolValue(location.IsPreferred)
		data.Slug = types.StringValue(location.Slug)

		if isFound {
			break
		}
	}

	if !isFound {
		var parameters string

		if !data.IsDefault.IsNull() {
			parameters = fmt.Sprintf("Default %s", data.IsDefault.String())
		}

		if !data.IsPreferred.IsNull() {
			if "" != parameters {
				parameters += ", "
			}

			parameters += fmt.Sprintf("Preferred %s", data.IsPreferred.String())
		}

		if "" != parameters {
			parameters = fmt.Sprintf(" (%s)", parameters)
		}

		if !data.DisplayName.IsNull() {
			resp.Diagnostics.AddError(
				"Location read error",
				fmt.Sprintf("No match found for display name: %s%s", data.DisplayName.ValueString(), parameters),
			)
		}

		if !data.Slug.IsNull() {
			resp.Diagnostics.AddError(
				"Location read error",
				fmt.Sprintf("No match found for slug: %s%s", data.Slug.ValueString(), parameters),
			)
		}

		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *Location) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

func (d *Location) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Warren Platform location",

		Attributes: map[string]schema.Attribute{
			"country_code": schema.StringAttribute{
				MarkdownDescription: "Location country code",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Location description",
				Computed:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Location display name",
				Computed:            true,
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Location slug",
				Computed:            true,
				Optional:            true,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "Location set as default",
				Computed:            true,
				Optional:            true,
			},
			"is_preferred": schema.BoolAttribute{
				MarkdownDescription: "Location set as preferred",
				Computed:            true,
				Optional:            true,
			},
		},
	}
}
