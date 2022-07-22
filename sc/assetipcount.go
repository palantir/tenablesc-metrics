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
	"strconv"
)

func (c *Client) getAssetIPCounts() (map[string]int64, error) {

	assetMap := make(map[string]int64)

	assets, err := c.GetAllAssets()
	if err != nil {
		return assetMap, err
	}

	for _, asset := range assets {
		ipCount, err := strconv.ParseInt(string(asset.IPCount), 10, 64)
		if err != nil {
			return assetMap, err
		}
		if ipCount != -1 {
			// ipCountMetricName -1 returns when it's mid-update; we should not update that metric with a -1
			assetMap[asset.Name] = ipCount
		}
	}

	return assetMap, nil
}
