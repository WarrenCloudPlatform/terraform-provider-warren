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
	"encoding/json"
	"net/http"
	"strings"
)

const (
	jsonImageData = `
{
	"os_name": "ubuntu",
	"display_name": "Ubuntu",
	"ui_position": 1,
	"is_default": true,
	"is_app_catalog": false,
	"icon": "...",
	"versions": [
		{
			"os_version": "21.04",
			"display_name": "21.04",
			"published": true
		},
		{
			"os_version": "20.04",
			"display_name": "20.04",
			"published": true
		}
	]
}
	`
	jsonServerDataTemplate = `
{
	"uuid": %q,
	"name": %q,
	"hostname": %q,
	"status": %q,
	"backup": false,
	"billing_account": 6,
	"created_at": "2018-02-22 14:24:30",
	"description": "Proudly copied from the Warren Platform Cloud API documentation",
	"id": 42,
	"mac": "52:54:00:59:44:d1",
	"memory": 2048,
	"os_name": "ubuntu",
	"os_version": "16.04",
	"private_ipv4": "10.42.0.1",
	"public_ipv6": "",
	"storage": [ %s ],
	"tags": null,
	"updated_at": "2018-02-22 14:24:30",
	"user_id": 8,
	"username": "example",
	"vcpu": 1
}
	`
	jsonServerStorageDataTemplate = `
	{
		"created_at": "2018-02-22 14:24:30.312877",
		"id": 42,
		"name": "sda",
		"pool": "default2",
		"primary": true,
		"replica": [],
		"shared": false,
		"size": 20,
		"type": "block",
		"updated_at": null,
		"user_id": 8,
		"uuid": %q
	}
	`
	TestServerNameTemplate = "machine-%s"
	TestServerUUID = "01234567-89ab-4def-0123-c56789abcdef"
)

// newJsonServerData generates a JSON server data object for testing purposes.
//
// PARAMETERS
// serverUUID  string Server ID to use
// serverState string Server state to use
func newJsonServerData(serverUUID string, serverState string) string {
	testServerName := fmt.Sprintf(TestServerNameTemplate, serverUUID)

	return fmt.Sprintf(
		jsonServerDataTemplate,
		serverUUID,
		testServerName,
		testServerName,
		serverState,
		newJsonServerStorageData(TestDiskUUID),
	)
}

// newJsonServerData generates a JSON server data object for testing purposes.
//
// PARAMETERS
// serverUUID  string Server ID to use
// serverState string Server state to use
func newJsonServerStorageData(diskUUID string) string {
	return fmt.Sprintf(jsonServerStorageDataTemplate, diskUUID)
}

// SetupVMEndpointOnMux configures a "/v1/user-resource/vm" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupVMEndpointOnMux(mux *http.ServeMux, emptyUntilCreated bool) {
	baseURL := "/v1/cyc01/user-resource/vm"
	isVMCreated := !emptyUntilCreated

	mux.HandleFunc(baseURL, func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if strings.ToLower(req.Method) == "delete" {
			isVMCreated = false
			res.WriteHeader(http.StatusOK)
		} else if strings.ToLower(req.Method) == "get" {
			queryParams := req.URL.Query()

			if queryParams.Get("uuid") == TestServerUUID && isVMCreated {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte(newJsonServerData(TestServerUUID, "started")))
			} else {
				res.WriteHeader(http.StatusNotFound)
				res.Write([]byte(`{ "errors": { "Error": "[404] Server not found" } }`))
			}
		} else if strings.ToLower(req.Method) == "post" {
			jsonData := make([]byte, req.ContentLength)
			req.Body.Read(jsonData)

			var data map[string]interface{}

			jsonErr := json.Unmarshal(jsonData, &data)
			if jsonErr != nil {
				panic(jsonErr)
			}

			isVMCreated = true
			res.WriteHeader(http.StatusCreated)

			res.Write([]byte(newJsonServerData(TestServerUUID, "stopped")))
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/list", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if strings.ToLower(req.Method) == "get" {
			res.WriteHeader(http.StatusOK)

			res.Write([]byte("["))

			if isVMCreated {
				res.Write([]byte(newJsonServerData(TestServerUUID, "running")))
			}

			res.Write([]byte("]"))
		} else {
			panic("Unsupported HTTP method call")
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/start", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		if isVMCreated {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(newJsonServerData(TestServerUUID, "started")))
		} else {
			res.WriteHeader(http.StatusNotFound)
			res.Write([]byte(`{ "errors": { "Error": "[404] Server not found" } }`))
		}
	})

	mux.HandleFunc(fmt.Sprintf("%s/storage/attach", baseURL), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		err := req.ParseForm()
		if nil != err {
			panic(err)
		}

		formParams := req.PostForm

		if formParams.Get("uuid") == TestServerUUID && formParams.Get("storage_uuid") == TestDiskUUID && isVMCreated {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(newJsonServerStorageData(TestDiskUUID)))
		} else {
			res.WriteHeader(http.StatusNotFound)
			res.Write([]byte(`{ "errors": { "Error": "[404] Server not found" } }`))
		}
	})
}

// SetupVMImagesEndpointOnMux configures a "/v1/config/vm_images" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupVMImagesEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc("/v1/cyc01/config/vm_images", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		res.WriteHeader(http.StatusOK)

		res.Write([]byte(fmt.Sprintf("[ %s ]", jsonImageData)))
	})
}
