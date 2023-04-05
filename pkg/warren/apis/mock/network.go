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

const (
	jsonFloatingIPDataTemplate = `
{
	"id": %d,
	"address": %q,
	"user_id": 8,
	"billing_account_id": 6,
	"type": "public",
	"name": "test",
	"enabled": true,
	"created_at": "2019-10-31 10:52:19",
	"updated_at": "2019-11-01 10:22:19",
	"uuid": "3456789a-bcde-4012-3f56-789abcdef012",
	"is_deleted": false,
	"is_ipv6": false,
	"assigned_to": "01234567-89ab-4def-0123-c56789abcdef",
	"assigned_to_private_ip": "10.42.0.1",
	"assigned_to_resource_type": "virtual_machine"
}
	`
	jsonNetworkDataTemplate = `
{
    "vlan_id": 42,
    "subnet": "10.42.0.0/24",
    "name": "test",
    "created_at": "2021-06-29 08:22:52",
    "updated_at": "2021-06-29 08:22:52",
    "uuid": %q,
    "type": "private",
    "is_default": true,
    "vm_uuids": [ "01234567-89ab-4def-0123-c56789abcdef" ],
    "resources_count": 0
}
	`
	TestFloatingIP = "42.42.42.42"
	TestFloatingIPID = 42
	TestNetworkUUID = "23456789-abcd-4f01-23e5-6789abcdef01"
)

// SetupIPAddressesEndpointOnMux configures a "/networks" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupIPAddressesEndpointOnMux(mux *http.ServeMux, emptyUntilCreated bool) {
	baseURL := "/v1/cyc01/network/ip_addresses"
	isFloatingIPCreated := !emptyUntilCreated

	mux.HandleFunc(baseURL, func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if strings.ToLower(req.Method) == "get" {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte("["))

			if isFloatingIPCreated {
				res.Write([]byte(fmt.Sprintf(jsonFloatingIPDataTemplate, TestFloatingIPID, TestFloatingIP)))
			}

			res.Write([]byte("]"))
		} else if strings.ToLower(req.Method) == "post" {
			isFloatingIPCreated = true
			res.WriteHeader(http.StatusCreated)

			res.Write([]byte(fmt.Sprintf(jsonFloatingIPDataTemplate, TestFloatingIPID, TestFloatingIP)))
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/%s/", baseURL, TestFloatingIP), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if strings.ToLower(req.Method) == "delete" {
			isFloatingIPCreated = false
			res.WriteHeader(http.StatusAccepted)
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/networks", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		res.WriteHeader(http.StatusOK)
		res.Write([]byte("["))

		if isFloatingIPCreated {
			res.Write([]byte(fmt.Sprintf(jsonNetworkDataTemplate, TestNetworkUUID)))
		}

		res.Write([]byte("]"))
	})
}

// SetupNetworkEndpointOnMux configures a "/networks" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupNetworkEndpointOnMux(mux *http.ServeMux, emptyUntilCreated bool) {
	baseURL := "/v1/cyc01/network"
	isNetworkCreated := !emptyUntilCreated

	mux.HandleFunc(fmt.Sprintf("%s/network", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if strings.ToLower(req.Method) == "post" {
			isNetworkCreated = true
			res.WriteHeader(http.StatusCreated)

			res.Write([]byte(fmt.Sprintf(jsonNetworkDataTemplate, TestNetworkUUID)))
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/network/%s/", baseURL, TestNetworkUUID), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if strings.ToLower(req.Method) == "delete" {
			isNetworkCreated = false
			res.WriteHeader(http.StatusAccepted)
		} else if strings.ToLower(req.Method) == "get" {
			if isNetworkCreated {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte(fmt.Sprintf(jsonNetworkDataTemplate, TestNetworkUUID)))
			} else {
				res.WriteHeader(http.StatusNotFound)
				res.Write([]byte(`{ "errors": { "Error": "[404] Server not found" } }`))
			}
		} else if strings.ToLower(req.Method) == "patch" {
			if isNetworkCreated {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte(fmt.Sprintf(jsonNetworkDataTemplate, TestNetworkUUID)))
			} else {
				res.WriteHeader(http.StatusNotFound)
				res.Write([]byte(`{ "errors": { "Error": "[404] Server not found" } }`))
			}
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/networks", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		res.WriteHeader(http.StatusOK)
		res.Write([]byte("["))

		if isNetworkCreated {
			res.Write([]byte(fmt.Sprintf(jsonNetworkDataTemplate, TestNetworkUUID)))
		}

		res.Write([]byte("]"))
	})
}
