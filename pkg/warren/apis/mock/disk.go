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
	jsonDiskDataTemplate = `
{
	"uuid": %q,
	"status": "Active",
	"user_id": 8,
	"billing_account_id": 6,
	"size_gb": 20,
	"source_image_type": "EMPTY",
	"created_at": "2022-09-01T12:03:14.355+0000",
	"updated_at": "2022-09-01T12:03:14.355+0000"
}
	`
	TestDiskUUID = "12345678-9abc-def0-1234-56789abcdef0"
)

// SetupDiskEndpointOnMux configures a "/networks" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupDiskEndpointOnMux(mux *http.ServeMux, emptyUntilCreated bool) {
	baseURL := "/v1/cyc01/storage"
	isDiskCreated := !emptyUntilCreated

	mux.HandleFunc(fmt.Sprintf("%s/disks", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if strings.ToLower(req.Method) == "get" {
			queryParams := req.URL.Query()

			if queryParams.Get("network_uuid") == TestDiskUUID && isDiskCreated {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte(fmt.Sprintf(jsonDiskDataTemplate, TestDiskUUID)))
			} else {
				res.WriteHeader(http.StatusNotFound)
				res.Write([]byte(`{ "errors": { "Error": "[404] Server not found" } }`))
			}
		} else if strings.ToLower(req.Method) == "post" {
			isDiskCreated = true
			res.WriteHeader(http.StatusCreated)

			res.Write([]byte(fmt.Sprintf(jsonDiskDataTemplate, TestDiskUUID)))
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/disk/%s", baseURL, TestDiskUUID), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if strings.ToLower(req.Method) == "delete" {
			isDiskCreated = false
			res.WriteHeader(http.StatusAccepted)
		} else if strings.ToLower(req.Method) == "get" {
			if isDiskCreated {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte(fmt.Sprintf(jsonDiskDataTemplate, TestDiskUUID)))
			} else {
				res.WriteHeader(http.StatusNotFound)
				res.Write([]byte(`{ "errors": { "Error": "[404] Server not found" } }`))
			}
		} else if strings.ToLower(req.Method) == "patch" {
			if isDiskCreated {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte(fmt.Sprintf(jsonDiskDataTemplate, TestDiskUUID)))
			} else {
				res.WriteHeader(http.StatusNotFound)
				res.Write([]byte(`{ "errors": { "Error": "[404] Server not found" } }`))
			}
		} else {
			panic("Unsupported HTTP method call")
		}
	})
}
