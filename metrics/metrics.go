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

package metrics

import (
	"fmt"

	"github.com/rcrowley/go-metrics"
)

var (
	registry metrics.Registry = metrics.DefaultRegistry
)

const (
	metricPrefix = "tenablesc"
)

// Okay, so all of this is currently _massive_ overkill because it's designed for a
// multithreaded solution, which we do not have here.
// The primary purpose of sc-metrics is to emit point-in-time metrics which _may age out_; this registry
// appears to continue to emit metrics at their last value? forever? which is bad and wrong.

// ResetRegistry creates a new registry, discarding all old metrics
func ResetRegistry() {
	registry = metrics.NewRegistry()
}

// SetRegistry updates the default registry to the provided one
func SetRegistry(r metrics.Registry) {
	registry = r
}

// GetRegistry returns the current metrics  registry
func GetRegistry() metrics.Registry {
	return registry
}

// Update sets the provided metric (name) to the specified value
func Update(name string, value int64) {
	metrics.GetOrRegisterGauge(fmt.Sprintf("%s.%s", metricPrefix, name), GetRegistry()).Update(value)
}

// Increment adds the specified value to the provided metric's (name) current value
func Increment(name string, value int64) {
	gauge := metrics.GetOrRegisterGauge(fmt.Sprintf("%s.%s", metricPrefix, name), GetRegistry())

	oldValue := gauge.Value()

	gauge.Update(oldValue + value)
}
