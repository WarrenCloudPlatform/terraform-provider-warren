/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package provider is the main Terraform provider code package
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren"
)

var testProviderV6Factories = map[string]func() (tfprotov6.ProviderServer, error){
	"warren": providerserver.NewProtocol6WithError(New(warren.ProviderVersion)()),
}

func TestWarren(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Terraform Provider Platform Suite")
}
