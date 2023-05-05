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

func generateVirtualMachineConfig(mockTestEnv mock.MockTestEnv, serverName string) string {
	return fmt.Sprintf(
		`
%s

resource "warren_virtual_machine" "test" {
	disk_size_in_gb = 20
	memory = 2048
	name = %q
	username = "example"
	os_name = "ubuntu"
	os_version = "16.04"
	vcpu = 1
}
		`,
		mockTestEnv.ProviderConfig,
		serverName,
	)
}

func VirtualMachineTests(providerFactories map[string]func() (tfprotov6.ProviderServer, error)) {
	var mockTestEnv mock.MockTestEnv
	t := GinkgoT()

	var _ = BeforeEach(func() {
		mockTestEnv = mock.NewMockTestEnv()

		apis.SetClientForToken("dummy-token", mockTestEnv.Client)
		mock.SetupNetworkEndpointOnMux(mockTestEnv.Mux, false)
	})

	var _ = AfterEach(func() {
		mockTestEnv.Teardown()
		apis.SetClientForToken("dummy-token", nil)
	})

	var _ = Describe("VirtualMachine", func() {
		It("is correctly imported", func() {
			mock.SetupVMEndpointOnMux(mockTestEnv.Mux, false)

			resource.UnitTest(
				t,
				resource.TestCase{
					ProtoV6ProviderFactories: providerFactories,
					Steps: []resource.TestStep{
						// ImportState testing
						{
							Config:        generateVirtualMachineConfig(mockTestEnv, fmt.Sprintf(mock.TestServerNameTemplate, mock.TestServerUUID)),
							ResourceName:  "warren_virtual_machine.test",
							ImportState:   true,
							ImportStateId: mock.TestServerUUID,
							Check:         resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr("warren_virtual_machine.test", "id", mock.TestServerUUID),
							),
						},
						// Delete testing automatically occurs in TestCase
					},
				},
			)
		})

		It("is correctly handled", func() {
			mock.SetupVMEndpointOnMux(mockTestEnv.Mux, true)

			resource.UnitTest(
				t,
				resource.TestCase{
					ProtoV6ProviderFactories: providerFactories,
					Steps: []resource.TestStep{
						// Create and Read testing
						{
							Config: generateVirtualMachineConfig(mockTestEnv, fmt.Sprintf(mock.TestServerNameTemplate, mock.TestServerUUID)),
							Check: resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr("warren_virtual_machine.test", "id", mock.TestServerUUID),
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
