/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package warren is the main provider code package for the Warren Platform
package warren

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.com/warrenio/library/go-client/warren"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis"
)

func NetworkReadData(client *warren.Client, data *NetworkModel) error {
	networks, err := client.Network.ListNetworks()
	if nil != err {
		return apis.GetNetworkErrorFromHttpCallError(err)
	}

	var isFound bool
	isParameterOnly := data.Name.IsNull() && data.UUID.IsNull()

	for _, network := range *networks {
		if !data.Name.IsNull() && data.Name.Equal(types.StringValue(network.Name)) {
			isFound = true
		}

		if !data.UUID.IsNull() && data.UUID.Equal(types.StringValue(network.Uuid)) {
			isFound = true
		}

		if !data.IsDefault.IsNull() {
			if !data.IsDefault.Equal(types.BoolValue(network.IsDefault)) {
				isFound = false
			} else if isParameterOnly {
				isFound = true
			}
		}

		if !isFound {
			continue
		}

		NetworkSetStateData(&network, data)

		if isFound {
			break
		}
	}

	if !isFound {
		var parameter string

		if !data.IsDefault.IsNull() {
			parameter = fmt.Sprintf("Default %s", data.IsDefault.String())
		}

		if isParameterOnly {
			return fmt.Errorf("No match found for parameter: %s", parameter)
		} else {
			if "" != parameter {
				parameter = fmt.Sprintf(" (%s)", parameter)
			}

			if !data.Name.IsNull() {
				return fmt.Errorf("No match found for name: %s%s", data.Name.ValueString(), parameter)
			}

			if !data.UUID.IsNull() {
				return fmt.Errorf("No match found for UUID: %s%s", data.UUID.ValueString(), parameter)
			}
		}
	}

	return nil
}

func NetworkSetStateData(network *warren.Network, data *NetworkModel) {
	data.CreatedAt = types.StringValue(network.CreatedAt)
	data.IsDefault = types.BoolValue(network.IsDefault)
	data.Name = types.StringValue(network.Name)
	data.SubnetIPv4 = types.StringValue(network.Subnet)
	data.SubnetIPv6 = types.StringValue(network.SubnetIPv6)
	data.Type = types.StringValue(network.Type)
	data.UpdatedAt = types.StringValue(network.UpdatedAt)
	data.UUID = types.StringValue(network.Uuid)
	data.VLANID = types.Int64Value(int64(network.VlanId))

	serverUUIDs := []attr.Value{}

	for _, serverUUID := range network.VmUuids {
		serverUUIDs = append(serverUUIDs, types.StringValue(serverUUID))
	}

	data.ServerUUIDs = types.ListValueMust(types.StringType, serverUUIDs)
}
