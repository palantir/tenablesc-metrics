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
	"time"

	"github.com/palantir/tenablesc-client/tenablesc"
	"github.com/rs/zerolog/log"
)

const (
	ceilingTimeInMinutes = 30 * 24 * 60
)

// getScheduledActiveScanAges returns a set of scan names and time since last scan in hours.
func (c *Client) getScheduledActiveScanAges() (map[string]int64, error) {

	// round to hours? round to hours.
	scanAges := make(map[string]int64)

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
			scanAge := minutesSinceEpochString(string(newestScanResult.FinishTime))
			log.Debug().Str("scanAgeEpoch", string(newestScanResult.FinishTime)).Int64("scanAgeDays", scanAge).Msg("got scan age")
			scanAges[scan.Name] = scanAge
		} else {
			log.Debug().Msg("scan had no recent results")
			if createdTime, err := scan.CreatedTime.ToDateTime(); err != nil {
				// if it was created in the last 48h, don't report if no scan results.
				if createdTime.After(time.Now().Add(-48 * time.Hour)) {
					log.Debug().Msg("scan is new, skipping report")
					continue // if it's verifiably new and has no results, don't report it.
				}
			}
			log.Debug().Msg("Giving scan max time due to no results.")
			// Scan's been around, but has no scan results in the scan period, return max.
			scanAges[scan.Name] = ceilingTimeInMinutes
		}
	}

	return scanAges, nil
}

func minutesSinceEpochString(s string) int64 {
	t := epochStringToTime(s)
	if t.IsZero() {
		return ceilingTimeInMinutes
	}

	return int64(time.Since(t).Minutes())
}

func newestResultForScan(scan *tenablesc.Scan, results []*tenablesc.ScanResult) *tenablesc.ScanResult {
	var newestResult *tenablesc.ScanResult
	var newestResultTime time.Time
	for _, result := range results {
		if result.Name != scan.Name {
			continue
		}
		if result.FinishTime == "-1" {
			// scan not complete, move along
			continue
		}
		if result.ImportStatus != "Finished" {
			// scan is currently importing, so can't be considered finished.
			// this can be an issue with job scheduling.
			continue
		}
		if newestResult == nil {
			newestResult = result
			newestResultTime = epochStringToTime(string(result.FinishTime))
		} else {
			finishTime := epochStringToTime(string(result.FinishTime))
			if finishTime.After(newestResultTime) {
				newestResult = result
				newestResultTime = finishTime
			}
		}
	}

	return newestResult
}

func epochStringToTime(s string) time.Time {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(sec, 0)
}

// We need to filter out scans which we don't care about the results of
// which should include:
// - scans created in the last 48h, which _might_ not have scan results yet.
// - scans whose schedule is ondemand or have no repeatrule
func shouldHaveScanResults(scan *tenablesc.Scan) bool {

	createdTime := epochStringToTime(string(scan.CreatedTime))
	if createdTime.After(time.Now().Add(-time.Hour * 48)) {
		return false
	}

	if scan.Schedule.Type != "ical" {
		return false
	}
	if scan.Schedule.RepeatRule == "" {
		return false
	}

	return true
}
