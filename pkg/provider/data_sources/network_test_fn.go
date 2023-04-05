/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package data_sources contains all Terraform data sources supported
package data_sources

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren/apis/mock"
)

func generateNetworkConfig(mockTestEnv mock.MockTestEnv, networkName string) string {
	return fmt.Sprintf(
		`
%s

data "warren_network" "test" {
	name = %q
}
		`,
		mockTestEnv.ProviderConfig,
		networkName,
	)
}

func NetworkTest(providerFactories map[string]func() (tfprotov6.ProviderServer, error)) {
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

	var _ = Describe("Network", func() {
		It("is correctly read", func() {
			resource.UnitTest(
				t,
				resource.TestCase{
					ProtoV6ProviderFactories: providerFactories,
					Steps: []resource.TestStep{
						// Read testing
						{
							Config: generateNetworkConfig(mockTestEnv, "test"),
							Check:  resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr("data.warren_network.test", "id", mock.TestNetworkUUID),
							),
						},
					},
				},
			)
		})

		Expect(t.Failed()).To(BeFalse())
	})
}
