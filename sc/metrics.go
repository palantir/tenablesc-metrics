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

func taggedMetric(name string, tagMap map[string]string) string {

	var tags []string

	for k, v := range tagMap {
		if len(v) == 0 {
			tags = append(tags, fmt.Sprintf("\"%s\"", k))
		} else {
			tags = append(tags, fmt.Sprintf("\"%s:%s\"", k, v))
		}
	}

	return fmt.Sprintf("%s[%s]", name, strings.Join(tags, ","))
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

	metrics[taggedMetric(healthyScannerCountMetricName, globalTagMap)] = scannerStatus.Unhealthy
	metrics[taggedMetric(unhealthyScannerCountMetricName, globalTagMap)] = scannerStatus.Unhealthy
	metrics[taggedMetric(totalScannerCountMetricName, globalTagMap)] = scannerStatus.Total

	for zone, status := range scannerStatus.ByZoneName {
		metrics[taggedMetric(healthyScannerCountMetricName, map[string]string{scanZoneTagName: zone})] = status.Healthy
		metrics[taggedMetric(unhealthyScannerCountMetricName, map[string]string{scanZoneTagName: zone})] = status.Unhealthy
		metrics[taggedMetric(totalScannerCountMetricName, map[string]string{scanZoneTagName: zone})] = status.Total
	}

	globalJobMetrics, jobTypeMetrics, err := adminClient.getJobMetrics()
	if err != nil {
		return nil, err
	}
	for metricName, metric := range globalJobMetrics {
		metrics[taggedMetric(metricName, globalTagMap)] = metric
	}
	for metricName, metric := range jobTypeMetrics {
		metrics[taggedMetric(jobQueueLengthMetricName, map[string]string{jobTypeTagName: metricName})] = metric
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
			metrics[taggedMetric(minutesSinceLastScanMetricName, map[string]string{orgTagName: orgName, scanNameTagName: scanName})] = metric
		}

		scanDurations, err := orgClient.getScheduledActiveScanDurations()
		if err != nil {
			return nil, err
		}
		for scanName, metric := range scanDurations {
			metrics[taggedMetric(scanDurationSecondsMetricName, map[string]string{orgTagName: orgName, scanNameTagName: scanName})] = metric
		}

		assetIPCounts, err := orgClient.getAssetIPCounts()
		if err != nil {
			return nil, err
		}
		for assetName, metric := range assetIPCounts {
			metrics[taggedMetric(ipCountMetricName, map[string]string{orgTagName: orgName, assetNameTagName: assetName})] = metric
		}

	}

	return metrics, nil
}
