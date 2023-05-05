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
	"strings"

	"gitlab.com/warrenio/library/go-client/warren"
)

func GetFloatingIPErrorFromHttpCallError(err error) error {
	if nil == err {
		return nil
	}

	errString := err.Error()

	if strings.HasPrefix(errString, "[") && strings.Index(errString, "]") == 4 {
		switch errString[1:4] {
		case "404":
			return ErrFloatingIPNotFound
		default:
			return GetErrorFromHttpCallError(err)
		}
	}

	return err
}

func GetFloatingIPByID(client *warren.Client, id int) (*warren.FloatingIp, error) {
	floatingIPs, err := client.Network.ListFloatingIps()
	if err != nil {
		return nil, GetFloatingIPErrorFromHttpCallError(err)
	}

	for _, ip := range *floatingIPs {
		if ip.Enabled && ip.Id == id {
			return &ip, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrFloatingIPNotFound, id)
}

func GetFloatingIPFromAssignedUUID(client *warren.Client, uuid string) (*warren.FloatingIp, error) {
	floatingIPs, err := client.Network.ListFloatingIps()
	if nil != err {
		return nil, GetFloatingIPErrorFromHttpCallError(err)
	}

	for _, ip := range *floatingIPs {
		if ip.Enabled && ip.AssignedTo == uuid {
			return &ip, nil
		}
	}

	return nil, fmt.Errorf("%w: No match for UUID %s", ErrFloatingIPNotFound, uuid)
}
