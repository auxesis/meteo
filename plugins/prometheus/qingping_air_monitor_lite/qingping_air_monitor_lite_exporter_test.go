package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

type TestMsg struct {
	TopicString  string
	PayloadBytes []byte
}

func (tm TestMsg) Duplicate() bool {
	return false
}
func (tm TestMsg) Qos() byte {
	return byte('a')
}
func (tm TestMsg) Retained() bool {
	return false
}
func (tm TestMsg) Topic() string {
	return tm.TopicString
}
func (tm TestMsg) MessageID() uint16 {
	return uint16(0)
}
func (tm TestMsg) Payload() []byte {
	return tm.PayloadBytes
}
func (tm TestMsg) Ack() {}

func TestMeasurementReader(t *testing.T) {
	assert := assert.New(t)

	// setup
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)
	ts := httptest.NewServer(promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	defer ts.Close()

	readMeasurement := measurementReader(metrics)
	c := mqtt.NewClient(mqtt.NewClientOptions())

	testCases := []struct {
		msg      TestMsg
		expected string
	}{
		{TestMsg{TopicString: "/topic/MACADDR/user/update", PayloadBytes: helperLoadBytes(t, "metric_sample.json")}, "upstairs_temperature_celsius 22.29"},
		{TestMsg{TopicString: "/topic/MACADDR/user/update", PayloadBytes: helperLoadBytes(t, "metric_sample.json")}, "upstairs_humidity_percentage 58.18"},
		{TestMsg{TopicString: "/topic/MACADDR/user/update", PayloadBytes: helperLoadBytes(t, "metric_sample.json")}, "upstairs_co2_parts_per_million 538"},
		{TestMsg{TopicString: "/topic/MACADDR/user/update", PayloadBytes: helperLoadBytes(t, "metric_sample.json")}, "upstairs_pm25_micrograms_per_meter_cubed 2"},
		{TestMsg{TopicString: "/topic/MACADDR/user/update", PayloadBytes: helperLoadBytes(t, "metric_sample.json")}, "upstairs_pm10_micrograms_per_meter_cubed 2"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			readMeasurement(c, tc.msg)

			resp, err := http.Get(ts.URL)
			assert.NoError(err)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(err)

			assert.Contains(string(body), tc.expected)
		})
	}

}
