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
	"fmt"
	"os"
	"strings"

	"github.com/palantir/tenablesc-metrics/version"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	versionFlag bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:              "sc-metrics",
	Short:            "tenable SC Metric Reporting Tool",
	Long:             "tenable SC Metric Reporting Tool",
	PersistentPreRun: setupLogging,
	RunE:             rootCmd,
}

func rootCmd(cmd *cobra.Command, args []string) error {
	if versionFlag {
		fmt.Println(version.GetVersion())
		return nil
	}
	return errors.New("subcommand required")
}

func init() {
	RootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "print version and exit")

	RootCmd.PersistentFlags().Bool("verbose", false, "Enable Verbose debugging output")
	RootCmd.PersistentFlags().String("config", "./config.yml", "path to config file")

	cobra.OnInitialize(initConfig)

	err := viper.BindPFlags(RootCmd.PersistentFlags())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to bind root persistent flags")
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName("sc-metrics") // name of config file (without extension)
	viper.AddConfigPath(".")          // optionally look for config in cwd
	if secret, ok := os.LookupEnv("NOMAD_SECRETS_DIR"); ok {
		viper.AddConfigPath(secret)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // check environment variables when accessing variables

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
}

func bindSubCmdFlags(cmd *cobra.Command, args []string) {
	logger := zerolog.Ctx(cmd.Context())
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to bind subcommand's flags")
	}

}

func setupLogging(cmd *cobra.Command, _ []string) {
	if viper.GetBool("verbose") {
		cmd.DebugFlags()
		viper.Debug()
	}

	cfg, err := readConfig(viper.GetString("config"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get config")
	}

	out := cmd.OutOrStdout()
	if cfg.Logging.Pretty {
		out = zerolog.ConsoleWriter{Out: out}
	}

	logger := zerolog.New(out).With().Timestamp().Logger()
	if cfg.Logging.Level != "" {
		level, err := zerolog.ParseLevel(cfg.Logging.Level)
		if err == nil {
			logger = logger.Level(level)
		} else {
			log.Warn().Msgf("Invalid log level %q, using the default level instead", cfg.Logging.Level)
		}
	}

	log.Debug().Msg("Logging set up")

	log.Logger = logger

}
