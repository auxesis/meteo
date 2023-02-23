package main

import (
	//"math/big"
	"io"
	"net/http"
	"net/http/httptest"
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
		{TestMsg{TopicString: "", PayloadBytes: []byte{}}, "outdoor_temperature_celsius 32"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			readMeasurement(c, tc.msg)

			resp, err := http.Get(ts.URL)
			assert.NoError(err)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(err)

			//t.Logf("body: %+v\n", string(body))
			assert.Contains(string(body), tc.expected)
		})
	}

}
