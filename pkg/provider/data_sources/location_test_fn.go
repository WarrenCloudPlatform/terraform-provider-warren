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

func generateLocationConfig(mockTestEnv mock.MockTestEnv, slug string) string {
	return fmt.Sprintf(
		`
%s

data "warren_location" "test" {
	id = %q
}
		`,
		mockTestEnv.ProviderConfig,
		slug,
	)
}

func LocationTest(providerFactories map[string]func() (tfprotov6.ProviderServer, error)) {
	var mockTestEnv mock.MockTestEnv
	t := GinkgoT()

	var _ = BeforeEach(func() {
		mockTestEnv = mock.NewMockTestEnv()

		apis.SetClientForToken("dummy-token", mockTestEnv.Client)
		mock.SetupLocationEndpointOnMux(mockTestEnv.Mux)
	})

	var _ = AfterEach(func() {
		mockTestEnv.Teardown()
		apis.SetClientForToken("dummy-token", nil)
	})

	var _ = Describe("Location", func() {
		It("is correctly read", func() {
			resource.UnitTest(
				t,
				resource.TestCase{
					ProtoV6ProviderFactories: providerFactories,
					Steps: []resource.TestStep{
						// Read testing
						{
							Config: generateLocationConfig(mockTestEnv, mockTestEnv.Client.LocationSlug),
							Check:  resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr("data.warren_location.test", "display_name", mock.TestLocationDisplayName),
							),
						},
					},
				},
			)
		})

		Expect(t.Failed()).To(BeFalse())
	})
}
