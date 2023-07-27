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
	"strings"
)

func buildTaggedMetricString(name string, tagMap map[string]string) string {

	var tagStrings []string

	for k, v := range tagMap {

		var tagValue string
		if v == "" {
			tagValue = noneTagValue
		} else {
			tagValue = v
		}
		tagStrings = append(tagStrings, fmt.Sprintf("%s:%s", k, tagValue))
	}

	if len(tagStrings) == 0 {
		return name
	}

	return name + fmt.Sprintf("[%s]", strings.Join(tagStrings, ","))
}

// GenerateMetricData generates a variety of metrics about SC and returns a map from metric name to value
func (c Config) GenerateMetricData() (map[string]int64, error) {
	metrics := make(map[string]int64)

	adminClient, err := c.TenableAdminClient()
	if err != nil {
		return nil, err
	}

	scannerStatus, err := adminClient.getScannerStatus()
	if err != nil {
		return nil, err
	}

	metrics[buildTaggedMetricString(healthyScannerCountMetricName, nil)] = scannerStatus.Unhealthy
	metrics[buildTaggedMetricString(unhealthyScannerCountMetricName, nil)] = scannerStatus.Unhealthy
	metrics[buildTaggedMetricString(totalScannerCountMetricName, nil)] = scannerStatus.Total

	for zone, status := range scannerStatus.ByZoneName {
		metrics[buildTaggedMetricString(healthyScannerCountMetricName, map[string]string{scanZoneTagName: zone})] = status.Healthy
		metrics[buildTaggedMetricString(unhealthyScannerCountMetricName, map[string]string{scanZoneTagName: zone})] = status.Unhealthy
		metrics[buildTaggedMetricString(totalScannerCountMetricName, map[string]string{scanZoneTagName: zone})] = status.Total
	}

	globalJobMetrics, jobTypeMetrics, err := adminClient.getJobMetrics()
	if err != nil {
		return nil, err
	}
	for metricName, metric := range globalJobMetrics {
		metrics[buildTaggedMetricString(metricName, nil)] = metric
	}
	for metricName, metric := range jobTypeMetrics {
		metrics[buildTaggedMetricString(jobQueueLengthMetricName, map[string]string{jobTypeTagName: metricName})] = metric
	}

	for _, cfgOrgName := range c.TenableOrgNames() {
		orgClient, err := c.TenableOrgClient(cfgOrgName)
		if err != nil {
			return nil, err
		}

		user, err := orgClient.GetCurrentUser()
		if err != nil {
			return nil, err
		}
		orgName := user.OrgName

		scanAges, err := orgClient.getScheduledActiveScanAges()
		if err != nil {
			return nil, err
		}
		for scanName, metric := range scanAges {
			metrics[buildTaggedMetricString(minutesSinceLastScanMetricName, map[string]string{orgTagName: orgName, scanNameTagName: scanName})] = metric
		}

		scanDurations, err := orgClient.getScheduledActiveScanDurations()
		if err != nil {
			return nil, err
		}
		for scanName, metric := range scanDurations {
			metrics[buildTaggedMetricString(scanDurationSecondsMetricName, map[string]string{orgTagName: orgName, scanNameTagName: scanName})] = metric
		}

		assetIPCounts, err := orgClient.getAssetIPCounts()
		if err != nil {
			return nil, err
		}
		for assetName, metric := range assetIPCounts {
			metrics[buildTaggedMetricString(ipCountMetricName, map[string]string{orgTagName: orgName, assetNameTagName: assetName})] = metric
		}

	}

	return metrics, nil
}
