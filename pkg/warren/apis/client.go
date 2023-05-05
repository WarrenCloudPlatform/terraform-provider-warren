/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package apis is the main package for Warren specific APIs
package apis

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.com/warrenio/library/go-client/warren"
)

var singletons = make(map[string]*warren.Client)

// getClient returns a configured Warren client.
//
// PARAMETERS
// url      string Warren client base URL
// token    string Warren client token
// location string Warren client location
func getClient(ctx context.Context, url, token, location string) *warren.Client {
	clientBuilder := (&warren.ClientBuilder{}).ApiUrl(url).ApiToken(token)
	if location != "" {
		clientBuilder = clientBuilder.LocationSlug(location)
	}

	client, err := clientBuilder.Build()
	if nil != err {
		tflog.Error(ctx, fmt.Sprintf("Warren platform client initialization failed: %w", err))
	}

	return client
}

// GetClientForToken returns an underlying Warren client for the given token.
//
// PARAMETERS
// token string Token to look up client instance for
func GetClientForTokenAndEndpoint(ctx context.Context, token, url string) *warren.Client {
	var client *warren.Client

	if url == "" {
		client, _ = singletons[token]

		if nil == client {
			url = os.Getenv("WARREN_API_URL")

			if "" == url {
				url = warrenDefaultURL
			}
		}
	}

	if nil == client {
		if url[len(url) - 1:] == "/" {
			url = url[:len(url) - 1]
		}

		var location string

		if strings.Count(url, "/") == 3 {
			location = os.Getenv("WARREN_API_LOCATION")
		} else {
			urlData := strings.Split(url, "/")

			location = urlData[len(urlData) - 1]
			url = strings.Join(urlData[0:len(urlData) - 1], "/")
		}

		client = getClient(ctx, url, token, location)
	}

    return client
}

func GetErrorFromHttpCallError(err error) error {
	if nil == err {
		return nil
	}

	errString := err.Error()

	if strings.HasPrefix(errString, "[") && strings.Index(errString, "]") == 4 {
		switch errString[1:4] {
		case "400":
		case "500":
			return fmt.Errorf("%w: %w", ErrUnknownInternal, err)
		case "429":
			return ErrRateLimitExceeded
		}
	}

	return err
}

// GetReconfiguredClientForLocation returns an underlying Warren client for
// the given location.
//
// PARAMETERS
// location string Warren client location
func GetReconfiguredClientForLocation(ctx context.Context, client *warren.Client, location string) *warren.Client {
	if location == "" || location == client.LocationSlug {
		return client
	}

	return getClient(ctx, client.BaseURL.String(), client.ApiToken, location)
}

// SetClientForToken sets a preconfigured Warren client for the given token.
//
// PARAMETERS
// token  string         Token to look up client instance for
// client *warren.Client Preconfigured Warren client
func SetClientForToken(token string, client *warren.Client) {
	if client == nil {
		delete(singletons, token)
	} else {
		singletons[token] = client
	}
}
