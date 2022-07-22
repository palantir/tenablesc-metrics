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
	"io/ioutil"
	"time"

	"github.com/palantir/tenablesc-metrics/sc"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

type config struct {
	Datadog struct {
		Address string   `yaml:"address"`
		Tags    []string `yaml:"tags"`
	} `yaml:"datadog"`
	Interval        time.Duration `yaml:"interval"`
	TenableSCConfig sc.Config     `yaml:"tenablesc"`
	Logging         struct {
		Level  string `yaml:"level"`
		Pretty bool   `yaml:"pretty"`
	} `yaml:"logging"`
}

func parseConfig(bytes []byte) (*config, error) {
	var c config
	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		return nil, errors.Wrapf(err, "failed unmarshalling yaml")
	}

	if c.TenableSCConfig.URL == "" {
		return nil, errors.New("No tenable URL set")
	}

	if c.Interval == 0 {
		c.Interval = 5 * time.Minute
	}

	if c.Datadog.Address == "" {
		c.Datadog.Address = "localhost:8125"
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}

	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return nil, errors.Wrapf(err, "error validating config")
	}

	return &c, nil
}

func readConfig(cfgFile string) (*config, error) {
	bytes, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading config file: %s", cfgFile)
	}

	cfg, err := parseConfig(bytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed parsing config")
	}

	return cfg, nil
}
