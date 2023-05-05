/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package main provides the application's entry point
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/go-hclog"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/provider"
	"gitlab.com/warrenio/library/terraform-provider-warren/pkg/warren"
)

// @TODO: Move me to cmd/terraform-provider-warren after Terraform does handle non-root entrypoints correctly

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: warren.PublishedName,
		Debug:   debug,
	}

	if debug {
		logger := hclog.Default()

		logger.Info(fmt.Sprintf("Compiled executable version: %s", warren.ProviderVersion))
		logger.Debug(fmt.Sprintf("Sys info: NumCPU: %v", runtime.NumCPU()))
	}

	err := providerserver.Serve(context.Background(), provider.New(warren.ProviderVersion), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
