package prometheus

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/auxesis/meteo/widget/internal/feedback"
	"github.com/auxesis/meteo/widget/internal/http"
	"github.com/auxesis/meteo/widget/internal/widget"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// PollForSamples polls a Prometheus endpoint, and updates the cache of samples
func PollForSamples(wdgts []widget.Widget, samples *http.Samples, errs chan feedback.Signal) {
	w := wdgts[0]
	client, err := api.NewClient(api.Config{
		Address: w.PrometheusURL,
	})
	if err != nil {
		log.Fatalf("error: unable to create Prometheus client: %s", err)
	}
	v1api := v1.NewAPI(client)

	fetchPrometheus(v1api, w, samples, errs) // first tick

	ticker := time.NewTicker(90 * time.Second)
	for range ticker.C {
		fetchPrometheus(v1api, w, samples, errs)
	}
}

func fetchPrometheus(v1api v1.API, w widget.Widget, samples *http.Samples, errs chan feedback.Signal) {
	log.Printf("debug: polling Prometheus\n")
	for k, v := range w.Metrics {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, warnings, err := v1api.Query(ctx, v.PrometheusQuery, time.Now(), v1.WithTimeout(10*time.Second))
		if err != nil {
			log.Printf("error: unable to query Prometheus: %s\n", err)
			errs <- feedback.NewSignal(err)
			continue
		}
		if len(warnings) > 0 {
			log.Printf("warning: when querying Prometheus: %v\n", warnings)
		}
		if len(result.String()) == 0 {
			err := fmt.Errorf("no data from Prometheus when scraping %s (%s)", k, v.PrometheusQuery)
			log.Printf("warning: %s\n", err)
			errs <- feedback.NewSignal(err)
			continue
		}
		results := strings.Split(result.String(), " ")
		v, err := strconv.ParseFloat(results[len(results)-2], 64)
		if err != nil {
			log.Printf("error: unable to parse value from Prometheus: %s\n", err)
			errs <- feedback.NewSignal(err)
			continue
		}
		(*samples)[k] = v
	}
}
