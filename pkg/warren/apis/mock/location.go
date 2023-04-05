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
	"strings"
)

const TestLocationDisplayName = "Cycletown"

// SetupLocationEndpointOnMux configures a "/networks" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupLocationEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc("/v1/cyc01/config/locations", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if strings.ToLower(req.Method) == "get" {
			res.WriteHeader(http.StatusOK)

			res.Write([]byte(fmt.Sprintf(`
[
	{
		"display_name": %q,
		"is_default": true,
		"is_preferred": false,
		"description": "The original location",
		"order_nr": 1,
		"slug": "cyc01",
		"country_code": "est"
	}
]
				`,
				TestLocationDisplayName,
			)))
		} else {
			panic("Unsupported HTTP method call")
		}
	})
}
