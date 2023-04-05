/*
Copyright 2022 OYE Network OÜ. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package mock provides all methods required to simulate a Warren Platform environment
package mock

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"gitlab.com/warrenio/library/go-client/warren"
)

// MockTestEnv represents the test environment for testing Warren Platform API calls
type MockTestEnv struct {
	Server         *httptest.Server
	Mux            *http.ServeMux
	Client         *warren.Client
	ProviderConfig string
}

const (
	TestNamespace = "test"
	TestProviderNamespace = "test"
	TestPlacementGroupID = "42"
	testPlacementGroupJsonValue = float64(42)
	TestSSHKey = "ssh-rsa invalid"
)

// Teardown shuts down the test environment server
func (env *MockTestEnv) Teardown() {
	env.Server.Close()

	env.Server = nil
	env.Mux = nil
	env.Client = nil
}

// NewMockTestEnv generates a new, unconfigured test environment for testing purposes.
func NewMockTestEnv() MockTestEnv {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	client, err := (&warren.ClientBuilder{}).ApiUrl(server.URL).ApiToken("dummy-token").LocationSlug("cyc01").Build()
	if nil != err {
		panic(err)
	}

	config := fmt.Sprintf(
		`
provider "warren" {
	api_token = "dummy-token"
	api_url   = "%s/v1/cyc01"
}
		`,
		client.BaseURL.String(),
	)

	return MockTestEnv{
		Server:         server,
		Mux:            mux,
		Client:         client,
		ProviderConfig: config,
	}
}
