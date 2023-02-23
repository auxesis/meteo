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

type TestMsg struct{}

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
	return ""
}
func (tm TestMsg) MessageID() uint16 {
	return uint16(0)
}
func (tm TestMsg) Payload() []byte {
	return []byte{}
}
func (tm TestMsg) Ack() {
	return
}

func TestMeasurementReader(t *testing.T) {
	assert := assert.New(t)

	// setup
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)

	readMeasurement := measurementReader(metrics)
	c := mqtt.NewClient(mqtt.NewClientOptions())
	msg := TestMsg{}
	readMeasurement(c, msg)

	ts := httptest.NewServer(promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.NoError(err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(err)

	//t.Logf("body: %+v\n", string(body))
	assert.Contains(string(body), "outdoor_temperature_celsius 32")
}
