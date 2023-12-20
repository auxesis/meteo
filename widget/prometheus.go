package main

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func pollForSamples(wdgts []Widget, samples *Samples) {
	w := wdgts[0]
	client, err := api.NewClient(api.Config{
		Address: w.PrometheusURL,
	})
	if err != nil {
		log.Fatalf("error: unable to create Prometheus client: %s", err)
	}
	v1api := v1.NewAPI(client)

	fetchPrometheus(v1api, w, samples) // first tick

	c := time.Tick(90 * time.Second)
	for range c {
		fetchPrometheus(v1api, w, samples)
	}
}

func fetchPrometheus(v1api v1.API, w Widget, samples *Samples) {
	log.Printf("debug: polling Prometheus\n")
	for k, v := range w.Metrics {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, warnings, err := v1api.Query(ctx, v.PrometheusQuery, time.Now(), v1.WithTimeout(10*time.Second))
		if err != nil {
			log.Printf("error: unable to query Prometheus: %s\n", err)
			continue
		}
		if len(warnings) > 0 {
			log.Printf("warning: when querying Prometheus: %v\n", warnings)
		}
		results := strings.Split(result.String(), " ")
		v, err := strconv.ParseFloat(results[len(results)-2], 64)
		if err != nil {
			log.Printf("error: unable to query Prometheus: %s\n", err)
			continue
		}
		(*samples)[k] = v
	}
}
