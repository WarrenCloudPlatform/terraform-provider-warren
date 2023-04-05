/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package data_sources contains all Terraform data sources supported
package data_sources

import (
    "github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.com/warrenio/library/go-client/warren"
)

// Location defines the data source implementation.
type Location struct {
	client *warren.Client
}

// LocationModel describes the data source model for an OS base image.
type LocationModel struct {
	CountryCode types.String `tfsdk:"country_code"`
	Description types.String `tfsdk:"description"`
	DisplayName types.String `tfsdk:"display_name"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	IsPreferred types.Bool   `tfsdk:"is_preferred"`
	Slug        types.String `tfsdk:"id"`
}

// Network defines the data source implementation.
type Network struct {
	client *warren.Client
}

// OSBaseImage defines the data source implementation.
type OSBaseImage struct {
	client *warren.Client
}

// OSBaseImageModel describes the data source model for an OS base image.
type OSBaseImageModel struct {
	DisplayName  types.String              `tfsdk:"display_name"`
	ID           types.String              `tfsdk:"id"`
	IsAppCatalog types.Bool                `tfsdk:"is_app_catalog"`
	IsDefault    types.Bool                `tfsdk:"is_default"`
	OSName       types.String              `tfsdk:"os_name"`
	OSVersion    types.String              `tfsdk:"os_version"`
	Versions     []OSBaseImageVersionModel `tfsdk:"versions"`
}

// OSBaseImageModel describes the data source model for versions of an OS base image.
type OSBaseImageVersionModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	OSVersion   types.String `tfsdk:"os_version"`
	Published   types.Bool   `tfsdk:"published"`
}
