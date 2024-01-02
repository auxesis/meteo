package prometheus

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auxesis/meteo/widget/internal/feedback"
	h "github.com/auxesis/meteo/widget/internal/http"
	"github.com/auxesis/meteo/widget/internal/widget"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusFeedbackIsSentWhenPrometheusUnavailable(t *testing.T) {
	assert := assert.New(t)

	client, err := api.NewClient(api.Config{
		Address: "http://prometheus-that-never-resolves.test",
	})
	assert.NoError(err)
	v1api := v1.NewAPI(client)
	w := widget.Widget{Metrics: map[string]widget.MetricConfig{"temperature": widget.MetricConfig{PrometheusQuery: "temperature_celsius"}}}
	samples := h.Samples{}
	feedback := make(chan feedback.Signal, 1)

	fetchPrometheus(v1api, w, &samples, feedback)

	assert.NotEmpty(feedback)
	f := <-feedback
	assert.Contains(f.Error.Error(), "no such host")
}

func TestPrometheusFeedbackIsSentWhenNoValue(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":"success","data":{"resultType":"vector","result":[]}}`)
	}))
	client, err := api.NewClient(api.Config{
		Address: ts.URL,
	})
	assert.NoError(err)
	v1api := v1.NewAPI(client)
	w := widget.Widget{Metrics: map[string]widget.MetricConfig{"temperature": widget.MetricConfig{PrometheusQuery: "outdoor_temperature_celsius"}}}
	samples := h.Samples{}
	feedback := make(chan feedback.Signal, 1)

	fetchPrometheus(v1api, w, &samples, feedback)

	assert.NotEmpty(feedback)
	f := <-feedback
	assert.Contains(f.Error.Error(), "no data from Prometheus when scraping temperature")
}

func TestPrometheusFeedbackIsSentWhenValueWeird(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"temperature_celsius","instance":"localhost:10000","job":"prometheus"},"value":[1704115202.421,"1.0.0"]}]}}`)
	}))
	client, err := api.NewClient(api.Config{
		Address: ts.URL,
	})
	assert.NoError(err)
	v1api := v1.NewAPI(client)
	w := widget.Widget{Metrics: map[string]widget.MetricConfig{"temperature": widget.MetricConfig{PrometheusQuery: "outdoor_temperature_celsius"}}}
	samples := h.Samples{}
	feedback := make(chan feedback.Signal, 1)

	fetchPrometheus(v1api, w, &samples, feedback)

	assert.NotEmpty(feedback)
	f := <-feedback
	assert.Contains(f.Error.Error(), "strconv.ParseFloat")
}
