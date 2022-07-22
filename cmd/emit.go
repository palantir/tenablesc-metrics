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

package cmd

import (
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/palantir/go-baseapp/baseapp/datadog"
	"github.com/palantir/tenablesc-metrics/metrics"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var emitMetricsCommand = &cobra.Command{
	Use:    "emit",
	Short:  "Emit DD Metrics",
	Long:   "Emit DD Metrics",
	PreRun: bindSubCmdFlags,
	RunE:   emitMetrics,
}

func init() {
	RootCmd.AddCommand(emitMetricsCommand)
	emitMetricsCommand.Flags().BoolP("once", "", false, "set to emit metrics once and exit.")
	emitMetricsCommand.Flags().BoolP("dry-run", "n", false, "set to emit only to stdout")
}

const (
	failedRunsMetric = "failedUpdate"
)

func emitMetrics(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	once := viper.GetBool("once")
	dryRun := viper.GetBool("dry-run")

	log.Debug().Bool("dryRun", dryRun).Bool("once", once).Msg("Starting emitMetrics")
	cfg, err := readConfig(viper.GetString("config"))
	if err != nil {
		log.Error().Err(err).Msg("failed to parse config")
		return err
	}

	var emitter *datadog.Emitter
	var ddClient *statsd.Client
	if !dryRun {
		log.Debug().Str("address", cfg.Datadog.Address).Interface("tags", cfg.Datadog.Tags).Msg("setting up datadog config")
		ddClient, err = statsdClient(cfg.Datadog.Address, cfg.Datadog.Tags)
		if err != nil {
			return err
		}
	}

	timer := time.NewTicker(cfg.Interval)
	defer timer.Stop()

	// weird syntax to run immediately
	// https://github.com/golang/go/issues/17601#issuecomment-311955879
	for ; true; <-timer.C {
		log.Debug().Msg("entering emit loop")

		if !dryRun {
			metrics.ResetRegistry()
			emitter = datadog.NewEmitter(ddClient, metrics.GetRegistry())
		}

		err := updateMetricsRegistry(cfg)
		if err != nil {
			log.Error().Err(err).Msg("failed to collect metrics")
		}

		if !dryRun {
			emitter.EmitOnce()
			err := emitter.Flush()
			if err != nil {
				return err
			}
		}

		if once {
			log.Info().Msg("Terminating due to once flag")
			break
		}
		log.Debug().Str("sleepDuration", cfg.Interval.String()).Msg("sleeping between updates")
	}

	return nil
}

func statsdClient(address string, tags []string) (*statsd.Client, error) {

	client, err := statsd.New(address, statsd.WithTags(tags))
	if err != nil {
		return nil, err
	}

	return client, nil

}

func updateMetricsRegistry(cfg *config) error {

	metricData, err := cfg.TenableSCConfig.GenerateMetricData()
	if err != nil {
		metrics.Increment(failedRunsMetric, 1)
		return err
	}

	for k, v := range metricData {
		log.Info().Int64(k, v).Msg("updating metric")
		metrics.Update(k, v)
	}

	metrics.Update(failedRunsMetric, 0)
	return nil
}
