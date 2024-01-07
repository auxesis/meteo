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

func TestPrometheusFeedbackIsSentWhenOk(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"temperature_celsius","instance":"localhost:10000","job":"prometheus"},"value":[1704115202.421,"1.0"]}]}}`)
	}))
	client, err := api.NewClient(api.Config{
		Address: ts.URL,
	})
	assert.NoError(err)
	v1api := v1.NewAPI(client)
	w := widget.Widget{Metrics: map[string]widget.MetricConfig{"temperature": widget.MetricConfig{PrometheusQuery: "outdoor_temperature_celsius"}}}
	feedback := make(chan feedback.Signal, 1)

	fetchPrometheus(v1api, w, feedback)

	assert.NotEmpty(feedback)
	f := <-feedback
	assert.True(f.Ok)
	assert.NoError(f.Error)
	assert.Equal(f.Metric, "temperature")
}

func TestPrometheusFeedbackIsSentWhenPrometheusUnavailable(t *testing.T) {
	assert := assert.New(t)

	client, err := api.NewClient(api.Config{
		Address: "http://prometheus-that-never-resolves.test",
	})
	assert.NoError(err)
	v1api := v1.NewAPI(client)
	w := widget.Widget{Metrics: map[string]widget.MetricConfig{"temperature": widget.MetricConfig{PrometheusQuery: "temperature_celsius"}}}
	feedback := make(chan feedback.Signal, 1)

	fetchPrometheus(v1api, w, feedback)

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
	feedback := make(chan feedback.Signal, 1)

	fetchPrometheus(v1api, w, feedback)

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
	feedback := make(chan feedback.Signal, 1)

	fetchPrometheus(v1api, w, feedback)

	assert.NotEmpty(feedback)
	f := <-feedback
	assert.Error(f.Error)
	assert.Contains(f.Error.Error(), "strconv.ParseFloat")
}

func TestPrometheusDoesNotUpdateWhenDeltaTooLarge(t *testing.T) {
	assert := assert.New(t)

	type test struct {
		name      string
		current   h.Samples
		changes   h.Samples
		different bool
	}
	tests := []test{
		{"initial", h.Samples{"temperature": 0.0}, h.Samples{"temperature": 10.0}, true},
		{"no change", h.Samples{"temperature": 10.0}, h.Samples{"temperature": 10.0}, true}, // not actually true, but we need to trigger the right test path
		{"20% increase", h.Samples{"temperature": 10.0}, h.Samples{"temperature": 12.0}, true},
		{"50% increase", h.Samples{"temperature": 10.0}, h.Samples{"temperature": 15.0}, true},
		{"100% increase", h.Samples{"temperature": 10.0}, h.Samples{"temperature": 20.0}, false},
		{"150% increase", h.Samples{"temperature": 10.0}, h.Samples{"temperature": 25.0}, false},
		{"20% decrease", h.Samples{"temperature": 10.0}, h.Samples{"temperature": 8.0}, true},
		{"50% decrease", h.Samples{"temperature": 10.0}, h.Samples{"temperature": 5.0}, true},
		{"100% decrease", h.Samples{"temperature": 10.0}, h.Samples{"temperature": 0.0}, false},
		{"150% decrease", h.Samples{"temperature": 10.0}, h.Samples{"temperature": -5.0}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updateSamples(&tc.current, tc.changes)
			if tc.different {
				// current should be updated to match changes
				assert.Equal(tc.current, tc.changes)
			} else {
				// current should not be updated
				assert.NotEqual(tc.current, tc.changes)
			}
		})
	}
}
