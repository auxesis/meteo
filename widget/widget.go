package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/auxesis/meteo/widget/internal/feedback"
	api "github.com/auxesis/meteo/widget/internal/http"
	"github.com/auxesis/meteo/widget/internal/prometheus"
	"github.com/auxesis/meteo/widget/internal/widget"
)

var (
	configPath string
	port       int
)

func init() {
	flag.StringVar(&configPath, "c", "config.toml", "path/to/widgets/config.toml")
	flag.IntVar(&port, "p", 10002, "port to run server")
}

func main() {
	flag.Parse()
	widgets, err := widget.LoadWidgets(configPath)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	samples := api.Samples{}
	sigs := make(chan feedback.Signal, 1024)
	status := feedback.Status{}
	go prometheus.PollForSamples(widgets, &samples, sigs)
	go feedback.ProcessSignals(sigs, &status)
	http.HandleFunc("/", api.HandleWidgetQuery(widgets, &samples, &status))

	log.Printf("info: starting server on port %d", port)
	for _, w := range widgets {
		log.Printf("info: serving widget for: %s", w.ID)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
