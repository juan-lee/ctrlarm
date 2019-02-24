// Copyright 2019 The ctrlarm Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/gobuffalo/envy"
)

// ParseEnvironment loads a sibling `.env` file then looks through all environment
// variables to set global configuration.
func ParseEnvironment() error {
	envy.Load()
	azureEnv, _ := azure.EnvironmentFromName("AzurePublicCloud") // shouldn't fail
	authorizationServerURL = azureEnv.ActiveDirectoryEndpoint

	var err error
	// these must be provided by environment
	// clientID
	clientID, err = envy.MustGet("AZURE_CLIENT_ID")
	if err != nil {
		return fmt.Errorf("expected env vars not provided: %s", err)
	}

	// clientSecret
	clientSecret, err = envy.MustGet("AZURE_CLIENT_SECRET")
	if err != nil {
		return fmt.Errorf("expected env vars not provided: %s", err)
	}

	// tenantID (AAD)
	tenantID, err = envy.MustGet("AZURE_TENANT_ID")
	if err != nil {
		return fmt.Errorf("expected env vars not provided: %s", err)
	}

	// subscriptionID (ARM)
	subscriptionID, err = envy.MustGet("AZURE_SUBSCRIPTION_ID")
	if err != nil {
		return fmt.Errorf("expected env vars not provided: %s", err)
	}

	return nil
}
