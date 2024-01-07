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

	latest := fetchPrometheus(v1api, w, errs) // first tick
	updateSamples(samples, latest, w)

	ticker := time.NewTicker(w.FetchInterval)
	for range ticker.C {
		latest = fetchPrometheus(v1api, w, errs)
		updateSamples(samples, latest, w)
	}
}

func fetchPrometheus(v1api v1.API, w widget.Widget, sigs chan feedback.Signal) http.Samples {
	samples := make(http.Samples)
	log.Printf("debug: polling Prometheus\n")
	for k, v := range w.Metrics {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, warnings, err := v1api.Query(ctx, v.PrometheusQuery, time.Now(), v1.WithTimeout(10*time.Second))
		if err != nil {
			log.Printf("error: unable to query Prometheus: %s\n", err)
			sigs <- feedback.NewSignalWithError(k, err)
			continue
		}
		if len(warnings) > 0 {
			log.Printf("warning: when querying Prometheus: %v\n", warnings)
		}
		if len(result.String()) == 0 {
			err := fmt.Errorf("no data from Prometheus when scraping %s (%s)", k, v.PrometheusQuery)
			log.Printf("warning: %s\n", err)
			sigs <- feedback.NewSignalWithError(k, err)
			continue
		}
		results := strings.Split(result.String(), " ")
		v, err := strconv.ParseFloat(results[len(results)-2], 64)
		if err != nil {
			log.Printf("error: unable to parse value from Prometheus: %s\n", err)
			sigs <- feedback.NewSignalWithError(k, err)
			continue
		}
		samples[k] = v
		sigs <- feedback.NewSignal(k)
	}
	return samples
}

// updateSamples takes a new http.Samples and updates an existing http.Samples
// updateSamples doesn't update if there's a > 50% variation in the value.
// This is done to handle weird outlier measurements returned by the weather station.
func updateSamples(old *http.Samples, latest http.Samples, w widget.Widget) {
	for k, l := range latest {
		var d float64
		o := (*old)[k]
		switch {
		case o == 0.0: // just booted
			// doesn't handle case where actual value is 0
			(*old)[k] = l
			continue
		case l == o: // no change
			continue
		case l > o:
			d = l - o
		case l < o:
			d = o - l
		}
		if w.Metrics[k].DampenOutliers {
			if d/o <= 0.5 {
				(*old)[k] = l
			} else {
				log.Printf("debug: ignoring update: > 50%% change on %s (%f, %f)", k, o, l)
			}
		} else {
			(*old)[k] = l
		}
	}
}
