/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package provider is the main Terraform provider code package
package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provider", func() {
	t := GinkgoT()

	It("validates correctly", func() {
		resource.UnitTest(
			t,
			resource.TestCase{
				IsUnitTest: true,
				ProtoV6ProviderFactories: testProviderV6Factories,
				Steps: []resource.TestStep{
					// Provider testing
					{ Config: `provider "warren" {}` },
				},
			},
		)

		Expect(t.Failed()).To(BeFalse())
	})
})
