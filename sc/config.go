// Copyright 2022 Palantir Technologies, Inc.
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

package sc

import (
	"fmt"
	"sort"

	"github.com/palantir/tenablesc-client/tenablesc"
)

// Config contains the required information to gather SC metrics.
// It requires both credentials of an SC admin, as well as organization admin credentials.
type Config struct {
	URL              string                 `yaml:"url"`
	AdminCredentials Credentials            `yaml:"adminCredentials,omitempty"`
	OrgCredentials   map[string]Credentials `yaml:"orgCredentials,omitempty"`
}

// Credentials containe the API credentials for SC
type Credentials struct {
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
}

// TenableAdminClient returns an SC client using the admin credentials from the config
func (c Config) TenableAdminClient() (*Client, error) {
	return c.AdminCredentials.client(c.URL)
}

func (c Credentials) client(url string) (*Client, error) {
	client := tenablesc.NewClient(url).SetAPIKey(c.AccessKey, c.SecretKey)

	_, err := client.GetCurrentUser()
	if err != nil {
		return nil, err
	}

	return &Client{Client: *client}, nil
}

// TenableOrgNames returns a slice of the org names for which credentials were provided
func (c Config) TenableOrgNames() []string {
	var names []string
	for name := range c.OrgCredentials {
		names = append(names, name)
	}

	//sort for consistency.  Maps are intentionally randomly ordered
	return sort.StringSlice(names)
}

// TenableOrgClient returns an SC client using the credentials for the provided org name
func (c Config) TenableOrgClient(orgName string) (*Client, error) {
	if cred, exists := c.OrgCredentials[orgName]; exists {
		return cred.client(c.URL)
	}
	return nil, fmt.Errorf("no org with name %s", orgName)
}
