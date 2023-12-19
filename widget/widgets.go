package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

// Samples is a map of the latest Prometheus samples
type Samples map[string]float64

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
	widgets, err := loadWidgets(configPath)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	samples := Samples{}
	go pollForSamples(widgets, &samples)
	http.HandleFunc("/", handleWidgetQuery(widgets, &samples))

	log.Printf("info: starting server on port %d", port)
	for _, w := range widgets {
		log.Printf("info: serving widget for: %s", w.ID)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
