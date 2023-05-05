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

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"gitlab.com/warrenio/library/go-client/warren"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis"
)

func NewOSBaseImage() datasource.DataSource {
	return &OSBaseImage{}
}

func (d *OSBaseImage) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*warren.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"OS base image configure error",
			fmt.Sprintf("Expected *warren.Client, got: %T", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *OSBaseImage) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("display_name"),
			path.MatchRoot("os_name"),
		),
	}
}

func (d *OSBaseImage) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OSBaseImageModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	images, err := d.client.VirtualMachine.ListBaseImages()
	if nil != err {
		resp.Diagnostics.AddError("OS base image read error", apis.GetImageErrorFromHttpCallError(err).Error())
		return
	}

	var isFound bool

	for _, image := range *images {
		if !data.DisplayName.IsNull() && data.DisplayName.Equal(types.StringValue(image.DisplayName)) {
			isFound = true
		}

		if !data.OSName.IsNull() && data.OSName.Equal(types.StringValue(image.OsName)) {
			isFound = true
		}

		if isFound && !data.OSVersion.IsNull() {
			var isOSVersionFound bool
			osVersion := data.OSVersion.ValueString()

			for _, imageVersion := range image.Versions {
				if osVersion == imageVersion.OsVersion {
					isOSVersionFound = true
					break
				}
			}

			if !isOSVersionFound {
				isFound = false
			}
		}

		if !isFound {
			continue
		}

		data.DisplayName = types.StringValue(image.DisplayName)
		data.IsAppCatalog = types.BoolValue(image.IsAppCatalog)
		data.IsDefault = types.BoolValue(image.IsDefault)
		data.OSName = types.StringValue(image.OsName)

		// ID is used for testing only
		if data.ID.IsNull() {
			data.ID = types.StringValue(uuid.NewString())
		}

		for _, imageVersion := range image.Versions {
			data.Versions = append(
				data.Versions,
				OSBaseImageVersionModel{
					DisplayName: types.StringValue(imageVersion.DisplayName),
					OSVersion:   types.StringValue(imageVersion.OsVersion),
					Published:   types.BoolValue(imageVersion.Published),
				},
			)
		}

		if isFound {
			break
		}
	}

	if !isFound {
		var additionalVersionConstraint string

		if !data.OSVersion.IsNull() {
			additionalVersionConstraint = fmt.Sprintf(" (Version %s)", data.OSVersion.ValueString())
		}

		if !data.DisplayName.IsNull() {
			resp.Diagnostics.AddError(
				"OS base image read error",
				fmt.Sprintf("No match found for display name: %s%s", data.DisplayName.ValueString(), additionalVersionConstraint),
			)
		}

		if !data.OSName.IsNull() {
			resp.Diagnostics.AddError(
				"OS base image read error",
				fmt.Sprintf("No match found for OS name: %s%s", data.OSName.ValueString(), additionalVersionConstraint),
			)
		}

		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *OSBaseImage) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_os_base_image"
}

func (d *OSBaseImage) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Warren Platform virtual machine OS base image",

		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				MarkdownDescription: "OS base image display name",
				Computed:            true,
				Optional:            true,
			},
			// ID is used for testing only
			"id": schema.StringAttribute{
				MarkdownDescription: "OS base image ID",
				DeprecationMessage:  "Try to Remove id Attribute Requirement - https://github.com/hashicorp/terraform-plugin-testing/issues/84",
				Computed:            true,
			},
			"is_app_catalog": schema.BoolAttribute{
				MarkdownDescription: "OS base image from app catalog",
				Computed:            true,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "OS base image set as default",
				Computed:            true,
			},
			"os_name": schema.StringAttribute{
				MarkdownDescription: "OS name",
				Computed:            true,
				Optional:            true,
			},
			"os_version": schema.StringAttribute{
				MarkdownDescription: "OS version",
				Optional:            true,
			},
			"versions": schema.ListNestedAttribute{
				MarkdownDescription: "OS versions",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{
							MarkdownDescription: "OS version display name",
							Computed:            true,
						},
						"os_version": schema.StringAttribute{
							MarkdownDescription: "OS version",
							Computed:            true,
						},
						"published": schema.BoolAttribute{
							MarkdownDescription: "OS version published state",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}
