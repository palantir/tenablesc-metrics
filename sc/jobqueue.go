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

	"github.com/rs/zerolog/log"
)

func (c *Client) getJobMetrics() (global map[string]int64, jobtypes map[string]int64, err error) {

	globalMetrics := make(map[string]int64)
	jobTypeMetrics := make(map[string]int64)

	// all right, the job queue.

	jobs, err := c.GetAllJobs()
	if err != nil {
		return nil, nil, err
	}
	globalMetrics[jobQueueLengthMetricName] = int64(len(jobs))

	// Buffering the 'now'; it's fine if jobs were targetted to start before now and haven't.
	// it's not fine if they were targetted to start _waaaay_ sooner and haven't.
	nowishEpoch := time.Now().Add(-5 * time.Minute).Unix()

	globalMetrics[jobsNotStartedMetricName] = 0
	for _, job := range jobs {

		targetedTime, err := strconv.ParseInt(string(job.TargetedTime), 10, 64)
		if err != nil {
			log.Err(err).Msg("Failed to parse targeted time as epoch time, skipping.")
			continue
		}
		if targetedTime > 0 && targetedTime < nowishEpoch {
			globalMetrics[jobsNotStartedMetricName]++
		}

		jobTypeMetrics[job.Type]++
	}

	return globalMetrics, jobTypeMetrics, nil
}
