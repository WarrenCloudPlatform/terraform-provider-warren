/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package apis is the main package for Warren specific APIs
package apis

import "strings"

func GetVolumeErrorFromHttpCallError(err error) error {
	if nil == err {
		return nil
	}

	errString := err.Error()

	if strings.HasPrefix(errString, "[") && strings.Index(errString, "]") == 4 {
		switch errString[1:4] {
		case "404":
			return ErrVolumeNotFound
		default:
			return GetErrorFromHttpCallError(err)
		}
	}

	return err
}
