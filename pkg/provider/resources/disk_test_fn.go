/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package resources contains all Terraform resources supported
package resources

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis/mock"
)

func generateDiskConfig(mockTestEnv mock.MockTestEnv) string {
	return fmt.Sprintf(
		`
%s

resource "warren_disk" "test" {
	server_uuid = %q
	size_in_gb = 20
}
		`,
		mockTestEnv.ProviderConfig,
		mock.TestServerUUID,
	)
}

func DiskTests(providerFactories map[string]func() (tfprotov6.ProviderServer, error)) {
	var mockTestEnv mock.MockTestEnv
	t := GinkgoT()

	var _ = BeforeEach(func() {
		mockTestEnv = mock.NewMockTestEnv()

		apis.SetClientForToken("dummy-token", mockTestEnv.Client)
		mock.SetupVMEndpointOnMux(mockTestEnv.Mux, false)
	})

	var _ = AfterEach(func() {
		mockTestEnv.Teardown()
		apis.SetClientForToken("dummy-token", nil)
	})

	var _ = Describe("Disk", func() {
		It("is correctly imported", func() {
			mock.SetupDiskEndpointOnMux(mockTestEnv.Mux, false)

			resource.UnitTest(
				t,
				resource.TestCase{
					ProtoV6ProviderFactories: providerFactories,
					Steps: []resource.TestStep{
						// ImportState testing
						{
							Config:        generateDiskConfig(mockTestEnv),
							ResourceName:  "warren_disk.test",
							ImportState:   true,
							ImportStateId: mock.TestDiskUUID,
							Check:         resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr("warren_disk.test", "id", mock.TestDiskUUID),
							),
						},
						// Delete testing automatically occurs in TestCase
					},
				},
			)
		})

		It("is correctly handled", func() {
			mock.SetupDiskEndpointOnMux(mockTestEnv.Mux, true)

			resource.UnitTest(
				t,
				resource.TestCase{
					ProtoV6ProviderFactories: providerFactories,
					Steps: []resource.TestStep{
						// Create and Read testing
						{
							Config: generateDiskConfig(mockTestEnv),
							Check:  resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr("warren_disk.test", "id", mock.TestDiskUUID),
							),
						},
						// Delete testing automatically occurs in TestCase
					},
				},
			)
		})

		Expect(t.Failed()).To(BeFalse())
	})
}
