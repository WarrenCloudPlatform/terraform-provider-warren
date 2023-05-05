/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package provider is the main Terraform provider code package
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/provider/data_sources"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/provider/resources"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis"
)

// WarrenProvider defines the provider implementation.
type WarrenProvider struct {
	// version is set to the provider version on build
	version string
}

// WarrenProviderModel describes the provider data model.
type WarrenProviderModel struct {
	APIToken types.String `tfsdk:"api_token"`
	APIURL   types.String `tfsdk:"api_url"`
}

func (p *WarrenProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "warren"
	resp.Version = p.version
}

func (p *WarrenProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Optional:    true,
				Description: "Token for the Warren platform API",
			},
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the Warren platform API",
			},
		},
	}
}

func (p *WarrenProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data WarrenProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiToken := data.APIToken.ValueString()

	// Configuration values are now available.
	if "" == apiToken {
		apiToken = os.Getenv("WARREN_API_TOKEN")
	}

	client := apis.GetClientForTokenAndEndpoint(ctx, apiToken, data.APIURL.ValueString())
	if client == nil {
		tflog.Error(ctx, "Failed to create Warren platform client")
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *WarrenProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewDisk,
		resources.NewFloatingIP,
		resources.NewNetwork,
		resources.NewVirtualMachine,
	}
}

func (p *WarrenProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		data_sources.NewLocation,
		data_sources.NewNetwork,
		data_sources.NewOSBaseImage,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WarrenProvider{
			version: version,
		}
	}
}
