/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package apis is the main package for Warren specific APIs
package apis

import "errors"

//
const warrenDefaultURL = "https://api.equinix.warren.io/v1"

//
var (
	ErrFloatingIPNotFound   = errors.New("Floating IP not found")
	ErrImageNotFound        = errors.New("Image not found")
	ErrLoadBalancerNotFound = errors.New("Load balancer not found")
	ErrLocationNotFound     = errors.New("Location not found")
	ErrRateLimitExceeded    = errors.New("API rate limit exceeded error")
	ErrNetworkNotFound      = errors.New("Network not found")
	ErrServerIsLocked       = errors.New("Server is locked")
	ErrServerNotFound       = errors.New("Server not found")
	ErrUnknownInternal      = errors.New("Internal API error")
	ErrVolumeNotFound       = errors.New("Volume not found")
)
