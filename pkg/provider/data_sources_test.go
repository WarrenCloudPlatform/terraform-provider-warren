/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package provider is the main Terraform provider code package
package provider

import (
	. "github.com/onsi/ginkgo/v2"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/provider/data_sources"
)

var _ = Describe("Resources", func() {
	data_sources.LocationTest(testProviderV6Factories)
	data_sources.NetworkTest(testProviderV6Factories)
	data_sources.OSBaseImageTest(testProviderV6Factories)
})
