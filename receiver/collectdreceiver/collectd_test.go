// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package collectdreceiver

import (
	"encoding/json"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDecodeEvent(t *testing.T) {
	metrics := pmetric.NewMetrics()
	sm := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty()
	jsonData, err := os.ReadFile(filepath.Join("testdata", "event.json"))
	require.NoError(t, err)

	var records []collectDRecord
	err = json.Unmarshal(jsonData, &records)
	require.NoError(t, err)

	for _, r := range records {
		err := r.appendToMetrics(sm, map[string]string{})
		assert.NoError(t, err)
		assert.Equal(t, metrics.MetricCount(), 0)
	}
}

func TestDecodeMetrics(t *testing.T) {
	metrics := pmetric.NewMetrics()
	scopeMemtrics := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty()
	jsonData, err := os.ReadFile(filepath.Join("testdata", "collectd.json"))
	require.NoError(t, err)

	var records []collectDRecord
	err = json.Unmarshal(jsonData, &records)
	require.NoError(t, err)

	for _, r := range records {
		err = r.appendToMetrics(scopeMemtrics, map[string]string{})
		assert.NoError(t, err)
	}
	assert.Equal(t, 10, metrics.MetricCount())

	assertMetricsEqual(t, metrics)
}

func TestLabelsFromName(t *testing.T) {
	tests := []struct {
		name           string
		wantMetricName string
		wantLabels     map[string]string
	}{
		{
			name:           "simple",
			wantMetricName: "simple",
		},
		{
			name:           "single[k=v]",
			wantMetricName: "single",
			wantLabels: map[string]string{
				"k": "v",
			},
		},
		{
			name:           "a.b.c.[k=v].d",
			wantMetricName: "a.b.c..d",
			wantLabels: map[string]string{
				"k": "v",
			},
		},
		{
			name:           "a.b[k0=v0,k1=v1,k2=v2].c",
			wantMetricName: "a.b.c",
			wantLabels: map[string]string{
				"k0": "v0", "k1": "v1", "k2": "v2",
			},
		},
		{
			name:           "empty[]",
			wantMetricName: "empty[]",
		},
		{
			name:           "mal.formed[k_no_sep]",
			wantMetricName: "mal.formed[k_no_sep]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetricName, gotLabels := LabelsFromName(&tt.name)
			assert.Equal(t, tt.wantMetricName, gotMetricName)
			assert.Equal(t, tt.wantLabels, gotLabels)
		})
	}
}

func createPtrJsonNumber(v json.Number) *json.Number {
	return &v
}

func TestStartTimestamp(t *testing.T) {
	tests := []struct {
		name                 string
		record               collectDRecord
		metricDescriptorType TargetMetricType
		wantStartTimestamp   pcommon.Timestamp
	}{
		{
			name: "metric type cumulative distribution",
			record: collectDRecord{
				Time:     createPtrJsonNumber(json.Number("10")),
				Interval: createPtrJsonNumber(json.Number("5")),
			},
			metricDescriptorType: CumulativeMetricType,
			wantStartTimestamp:   pcommon.NewTimestampFromTime(time.Unix(5, 0)),
		},
		{
			name: "metric type cumulative double",
			record: collectDRecord{
				Time:     createPtrJsonNumber(json.Number("10")),
				Interval: createPtrJsonNumber(json.Number("5")),
			},
			metricDescriptorType: CumulativeMetricType,
			wantStartTimestamp:   pcommon.NewTimestampFromTime(time.Unix(5, 0)),
		},
		{
			name: "metric type cumulative int64",
			record: collectDRecord{
				Time:     createPtrJsonNumber(json.Number("10")),
				Interval: createPtrJsonNumber(json.Number("5")),
			},
			metricDescriptorType: CumulativeMetricType,
			wantStartTimestamp:   pcommon.NewTimestampFromTime(time.Unix(5, 0)),
		},
		{
			name: "metric type non-cumulative gauge distribution",
			record: collectDRecord{
				Time:     createPtrJsonNumber(json.Number("0")),
				Interval: createPtrJsonNumber(json.Number("0")),
			},
			metricDescriptorType: GaugeMetricType,
			wantStartTimestamp:   pcommon.NewTimestampFromTime(time.Time{}),
		},
		{
			name: "metric type non-cumulative gauge int64",
			record: collectDRecord{
				Time:     createPtrJsonNumber(json.Number("0")),
				Interval: createPtrJsonNumber(json.Number("0")),
			},
			metricDescriptorType: GaugeMetricType,
			wantStartTimestamp:   pcommon.NewTimestampFromTime(time.Time{}),
		},
		{
			name: "metric type non-cumulativegauge double",
			record: collectDRecord{
				Time:     createPtrJsonNumber(json.Number("0")),
				Interval: createPtrJsonNumber(json.Number("0")),
			},
			metricDescriptorType: GaugeMetricType,
			wantStartTimestamp:   pcommon.NewTimestampFromTime(time.Time{}),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStartTimestamp := tc.record.startTimestamp(tc.metricDescriptorType)
			assert.Equal(t, tc.wantStartTimestamp.AsTime(), gotStartTimestamp.AsTime())
		})
	}
}

func assertMetricsEqual(t *testing.T, actual pmetric.Metrics) {
	goldenPath := filepath.Join("testdata", "expected.yaml")
	expectedMetrics, err := golden.ReadMetrics(goldenPath)
	require.NoError(t, err)

	err = pmetrictest.CompareMetrics(expectedMetrics, actual)
	require.NoError(t, err)
}
