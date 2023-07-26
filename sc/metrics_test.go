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
	"testing"
)

func Test_buildTaggedMetricString(t *testing.T) {

	tests := []struct {
		name       string
		metricName string
		tagMap     map[string]string
		want       string
	}{
		{
			name: "empty metric",
			want: "",
		},
		{
			name:       "metric with no tags",
			metricName: scanDurationSecondsMetricName,
			want:       scanDurationSecondsMetricName,
		},
		{
			name:       "metric with single tag",
			metricName: scanDurationSecondsMetricName,
			tagMap:     map[string]string{"foo-key": "foo-value"},
			want:       `scanDurationSeconds[foo-key:foo-value]`,
		},
		{
			name:       "metric with single tag(empty value)",
			metricName: scanDurationSecondsMetricName,
			tagMap:     map[string]string{"foo-key": ""},
			want:       scanDurationSecondsMetricName,
		},
		{
			name:       "metric with multi tag",
			metricName: scanDurationSecondsMetricName,
			tagMap: map[string]string{
				"foo-key": "foo-value",
				"bar-key": "bar-value",
			},
			want: `scanDurationSeconds[foo-key:foo-value,bar-key:bar-value]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildTaggedMetricString(tt.metricName, tt.tagMap); got != tt.want {
				t.Errorf("buildTaggedMetricString() = %v, want %v", got, tt.want)
			}
		})
	}
}
