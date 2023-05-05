/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package warren is the main provider code package for the Warren Platform
package warren

import "github.com/hashicorp/terraform-plugin-framework/types"

// NetworkModel describes the data source and resource model for an network.
type NetworkModel struct {
	CreatedAt   types.String `tfsdk:"created_at"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Name        types.String `tfsdk:"name"`
	ServerUUIDs types.List   `tfsdk:"server_uuids"`
	SubnetIPv4  types.String `tfsdk:"subnet_ipv4"`
	SubnetIPv6  types.String `tfsdk:"subnet_ipv6"`
	Type        types.String `tfsdk:"type"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	UUID        types.String `tfsdk:"id"`
	VLANID      types.Int64  `tfsdk:"vlan_id"`
}

// TODO: Update this string with the published name of your provider.
const PublishedName = "registry.terraform.io/warrenio/warren"

var ProviderVersion = "v0.0.0"
