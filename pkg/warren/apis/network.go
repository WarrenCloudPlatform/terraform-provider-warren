/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package apis is the main package for Warren specific APIs
package apis

import (
	"fmt"
	"net"
	"strings"

	"gitlab.com/warrenio/library/go-client/warren"
)

func GetNetworkErrorFromHttpCallError(err error) error {
	if nil == err {
		return nil
	}

	errString := err.Error()

	if strings.HasPrefix(errString, "[") && strings.Index(errString, "]") == 4 {
		switch errString[1:4] {
		case "404":
			return ErrNetworkNotFound
		default:
			return GetErrorFromHttpCallError(err)
		}
	}

	return err
}

func GetNetworkFromServerUUID(client *warren.Client, uuid string) (*warren.Network, error) {
	server, err := client.VirtualMachine.GetByUuid(uuid)
	if nil != err {
		return nil, GetServerErrorFromHttpCallError(err)
	}

	var serverPrivateIPv4 net.IP

	if server.PrivateIPv4 != "" {
		serverPrivateIPv4 = net.ParseIP(server.PrivateIPv4)
	}

	networks, err := client.Network.ListNetworks()
	if err != nil {
		return nil, GetNetworkErrorFromHttpCallError(err)
	}

	for _, network := range *networks {
		for _, networkServerUUID := range network.VmUuids {
			if networkServerUUID == uuid {
				// Parse subnet only after we found a possible match
				_, subnet, err := net.ParseCIDR(network.Subnet)
				if nil != err {
					return nil, err
				}

				if subnet.Contains(serverPrivateIPv4) {
					return &network, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("%w: Server UUID %s", ErrNetworkNotFound, uuid)
}
