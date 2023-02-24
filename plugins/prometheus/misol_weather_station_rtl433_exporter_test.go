package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

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
func (tm TestMsg) Ack() {
	return
}

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
		{TestMsg{TopicString: "battery_ok", PayloadBytes: []byte(strconv.Itoa(1))}, "outdoor_battery 1"},
		{TestMsg{TopicString: "temperature_C", PayloadBytes: []byte(strconv.FormatFloat(18.3, 'f', -1, 64))}, "outdoor_temperature_celsius 18.3"},
		{TestMsg{TopicString: "humidity", PayloadBytes: []byte(strconv.FormatFloat(47, 'f', -1, 64))}, "outdoor_humidity_percentage 47"},
		{TestMsg{TopicString: "wind_dir_deg", PayloadBytes: []byte(strconv.FormatFloat(170, 'f', -1, 64))}, "outdoor_wind_direction_degree 170"},
		{TestMsg{TopicString: "wind_avg_km_h", PayloadBytes: []byte(strconv.FormatFloat(20, 'f', -1, 64))}, "outdoor_wind_speed_average_kilometers_per_hour 20"},
		{TestMsg{TopicString: "wind_max_km_h", PayloadBytes: []byte(strconv.FormatFloat(45, 'f', -1, 64))}, "outdoor_wind_speed_average_max_kilometers_per_hour 45"},
		{TestMsg{TopicString: "rain_mm", PayloadBytes: []byte(strconv.FormatFloat(3.5, 'f', -1, 64))}, "outdoor_rain_millimetres 3.5"},
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
