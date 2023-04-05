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

func GetServerErrorFromHttpCallError(err error) error {
	if nil == err {
		return nil
	}

	errString := err.Error()

	if strings.HasPrefix(errString, "[") && strings.Index(errString, "]") == 4 {
		switch errString[1:4] {
		case "400":
			if strings.Index(errString, "No such virtual machine exists") > -1 {
				return ErrServerNotFound
			}

			return GetErrorFromHttpCallError(err)
		case "404":
			return ErrServerNotFound
		case "409":
			return ErrServerIsLocked
		default:
			return GetErrorFromHttpCallError(err)
		}
	}

	return err
}

func GetServerFromVolumeUUID(client *warren.Client, uuid string) (*warren.VirtualMachine, error) {
	servers, err := client.VirtualMachine.ListVms()
	if nil != err {
		return nil, GetServerErrorFromHttpCallError(err)
	}

	for _, server := range *servers {
		for _, storage := range server.Storage {
			if storage.Uuid == uuid {
				return &server, nil
			}
		}
	}

	return nil, fmt.Errorf("%w: No match found for UUID %s", ErrServerNotFound, uuid)
}
