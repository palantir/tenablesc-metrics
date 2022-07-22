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

	"github.com/rs/zerolog/log"
)

// getScheduledActiveScanDurations returns a set of scan names and how long the last one took to complete.
func (c *Client) getScheduledActiveScanDurations() (map[string]int64, error) {

	// round to hours? round to hours.
	scanDurations := make(map[string]int64)

	scans, err := c.GetAllScans()
	if err != nil {
		return nil, err
	}

	scanResults, err := c.GetAllScanResults()
	if err != nil {
		return nil, err
	}

	for _, scan := range scans {
		log := log.With().Str("scan name", scan.Name).Logger()

		if !shouldHaveScanResults(scan) {
			log.Debug().Msg("scan not expected to have results")
			continue
		}

		if newestScanResult := newestResultForScan(scan, scanResults); newestScanResult != nil {

			scanDuration, err := strconv.ParseInt(string(newestScanResult.ScanDuration), 10, 64)
			if err != nil {
				return scanDurations, err
			}

			log.Debug().Str(scanDurationSecondsMetricName, string(newestScanResult.ScanDuration)).Int64("scanAgeDays", scanDuration).Msg("got scan duration")
			scanDurations[scan.Name] = scanDuration
		} else {
			log.Debug().Msg("scan had no recent results")
		}
	}

	return scanDurations, nil
}
