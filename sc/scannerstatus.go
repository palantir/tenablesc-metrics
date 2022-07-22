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
	"github.com/palantir/tenablesc-client/tenablesc"
	"github.com/rs/zerolog/log"
)

type scannerStatus struct {
	healthCount

	ByZoneName map[string]*healthCount
}

type healthCount struct {
	Total, Healthy, Unhealthy int64
}

func (c *Client) getScannerStatus() (scannerStatus, error) {

	status := scannerStatus{}
	status.ByZoneName = make(map[string]*healthCount)

	scanZones, err := c.GetAllScanZones()
	if err != nil {
		return scannerStatus{}, err
	}
	scannerIDToZoneNameMap := scannerIDToZoneNameMap(scanZones)

	log.Debug().Interface("scannerzones", scannerIDToZoneNameMap).Msg("got scan zones map")

	scanners, err := c.GetAllScanners()
	if err != nil {
		return scannerStatus{}, err
	}

	for _, scanner := range scanners {
		zoneName := scannerIDToZoneNameMap[string(scanner.ID)]
		if zoneName == "" {
			zoneName = "no-associated-zone"
		}
		log.Debug().Str("zone", zoneName).Str("scannerName", scanner.Name).Msg("checking scanner")
		zoneStatus, ok := status.ByZoneName[zoneName]
		if !ok {
			zoneStatus = &healthCount{}
			status.ByZoneName[zoneName] = zoneStatus
		}

		status.Total++
		zoneStatus.Total++

		if scannerIsHealthy(scanner) {
			status.Healthy++
			zoneStatus.Healthy++
		} else {
			status.Unhealthy++
			zoneStatus.Unhealthy++
		}
	}

	return status, nil
}

func scannerIDToZoneNameMap(scanzones []*tenablesc.ScanZone) map[string]string {
	scannerIDToZoneNameMap := make(map[string]string)

	for _, zone := range scanzones {
		for _, scanner := range zone.Scanners {
			scannerIDToZoneNameMap[string(scanner.ID)] = zone.Name
		}
	}

	return scannerIDToZoneNameMap
}

func scannerIsHealthy(scanner *tenablesc.Scanner) bool {
	return scanner.Status == "1"
}
