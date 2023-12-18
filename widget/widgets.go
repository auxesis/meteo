package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/shopspring/decimal"
)

// Widget is a container for a widget.json-formatted response, suitable for WCS
type Widget struct {
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	Data          map[string]string       `json:"data"`
	Layouts       map[string]WidgetLayout `json:"layouts"`
	ID            string                  `json:"-"`
	Token         string                  `json:"-"`
	Metrics       map[string]MetricConfig `json:"-"`
	WidgetURL     string                  `json:"-" toml:"widget_url"`
	PrometheusURL string                  `json:"-" toml:"prometheus_url"`
}

// MetricConfig defines how to gather and display a metric as data
type MetricConfig struct {
	DisplayUnit     string `toml:"display_unit"`
	PrometheusQuery string `toml:"prometheus_query"`
	Levels          map[string]int
}

// WidgetLayout is a layout for a widget.json widget
type WidgetLayout struct {
	Size   string        `json:"size"`
	Styles WidgetStyles  `json:"styles"`
	Layers []WidgetLayer `json:"layers"`
}

// WidgetStyles is a collection of styles for a widget.json widget
type WidgetStyles struct {
	Colors map[string]WidgetColor `json:"colors"`
}

// WidgetColor is a color for a widget.json widget
type WidgetColor struct {
	Color string `json:"color"`
}

// WidgetLayer is a layer for a widget.json widget
type WidgetLayer struct {
	Rows []WidgetRow `json:"rows"`
}

// WidgetRow is a layer row for a widget.json widget
type WidgetRow struct {
	Height float64      `json:"height"`
	Cells  []WidgetCell `json:"cells,omitempty"`
}

// WidgetCell is a cell for layer row for a widget.json widget
type WidgetCell struct {
	Width                int        `json:"width"`
	BackgroundColorStyle string     `json:"background_color_style,omitempty" toml:"background_color_style"`
	Padding              float64    `json:"padding,omitempty"`
	Text                 WidgetText `json:"text,omitempty"`
	LinkURL              string     `json:"link_url,omitempty"`
}

// WidgetText is a text object, for a cell, for layer row, for a widget.json widget
type WidgetText struct {
	String         string  `json:"string,omitempty"`
	DataRef        string  `json:"data_ref,omitempty"`
	Size           float64 `json:"size,omitempty"`
	ColorStyle     string  `json:"color_style,omitempty"`
	FontStyle      string  `json:"font_style,omitempty"`
	Weight         string  `json:"weight,omitempty"`
	Justification  string  `json:"justification,omitempty"`
	MinScaleFactor float64 `json:"min_scale_factor,omitempty"`
}

// Samples is a map of the latest Prometheus samples
type Samples map[string]float64

// handleWidgetQuery handles rendering a widget in the WCS widget.json format
func handleWidgetQuery(wdgts []Widget, samples *Samples) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request: %s", r.URL)
		w.Header().Add("Content-Type", "application/json")

		re := regexp.MustCompile(`/widgets/(.*)`)
		if !re.MatchString(r.URL.Path) {
			log.Printf("error: bad URL: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		matches := re.FindStringSubmatch(r.URL.Path)
		id := matches[len(matches)-1]
		q := r.URL.Query()
		t := q.Get("token")

		var widget Widget
		for _, wi := range wdgts {
			if wi.ID == id && wi.Token == t {
				widget = wi
				break
			}
		}
		if len(widget.ID) == 0 && len(widget.Token) == 0 {
			log.Printf("error: unable to find widget for ID \"%s\" and token \"%s\"", id, t)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		widget = addDataFromSamples(widget, samples)
		widget = adjustColorsFromThresholds(widget, samples)
		err := json.NewEncoder(w).Encode(widget)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// addDataFromSamples populates a widget's data with the latest samples
func addDataFromSamples(w Widget, s *Samples) Widget {
	for k, c := range w.Metrics {
		f := (*s)[k]
		v := decimal.NewFromFloat(f)
		var vs string
		if v.Exponent() < -3 {
			vs = v.String()
		} else {
			vs = v.RoundDown(1).String()
		}
		w.Data[k] = fmt.Sprintf("%s%s", vs, c.DisplayUnit)
	}
	return w
}

// findColorForValue returns a color for a value, given a map of thresholds
func findColorForValue(v float64, levels map[string]int) string {
	switch true {
	case v >= float64(levels["high"]):
		return "red-500"
	case v >= float64(levels["medium"]):
		return "yellow-500"
	case v >= float64(levels["low"]):
		return "green-500"
	default:
		return "blue-500"
	}
}

// adjustColorsFromThresholds changes a widget's cell colors based on thresholds and samples
func adjustColorsFromThresholds(w Widget, s *Samples) Widget {
	targets := map[string]map[string]int{}
	for k, v := range w.Metrics {
		if v.Levels != nil {
			targets[k] = v.Levels
			//w.Data[k] = fmt.Sprintf("%f%s", (*s)[k], v.DisplayUnit)
		}
	}
	for n, l := range targets {
		for _, lyts := range w.Layouts {
			for _, lyrs := range lyts.Layers {
				for _, r := range lyrs.Rows {
					for i, c := range r.Cells {
						if c.Text.DataRef == n {
							v := (*s)[n]
							r.Cells[i].Text.ColorStyle = findColorForValue(v, l)
						}
					}
				}
			}
		}
	}
	return w
}

var layouts = map[string]WidgetLayout{
	"weather_small": WidgetLayout{
		Size: "small",
		Styles: WidgetStyles{
			Colors: map[string]WidgetColor{
				"black":      WidgetColor{Color: "#000000"},
				"stone-100":  WidgetColor{Color: "#f5f5f4"},
				"stone-950":  WidgetColor{Color: "#0c0a09"},
				"blue-500":   WidgetColor{Color: "#3b82f6"},
				"green-500":  WidgetColor{Color: "#84cc16"},
				"yellow-500": WidgetColor{Color: "#facc15"},
				"red-500":    WidgetColor{Color: "#ef4444"},
			},
		},
		Layers: []WidgetLayer{
			WidgetLayer{
				Rows: []WidgetRow{{
					Height: 12,
					Cells: []WidgetCell{{
						Width:                12,
						BackgroundColorStyle: "stone-950",
					}},
				}},
			},
			WidgetLayer{
				Rows: []WidgetRow{
					{Height: 1},
					{
						Height: 2.75,
						Cells: []WidgetCell{
							{
								Width:   12,
								Padding: 1.15,
								Text: WidgetText{
									DataRef:       "temperature",
									Size:          40,
									ColorStyle:    "green-500",
									Weight:        "bold",
									Justification: "left",
								},
							},
						},
					},
					{Height: 1},
					{
						Height: 2.25,
						Cells: []WidgetCell{
							{
								Width:   12,
								Padding: 1.15,
								Text: WidgetText{
									DataRef:       "humidity",
									Size:          18,
									ColorStyle:    "yellow-500",
									Justification: "left",
								},
							},
						},
					},
					{
						Height: 1.75,
						Cells: []WidgetCell{
							{
								Width:   12,
								Padding: 1.15,
								Text: WidgetText{
									DataRef:       "wind_gust",
									Size:          14,
									ColorStyle:    "stone-100",
									Justification: "left",
								},
							},
						},
					},
					{
						Height: 1.75,
						Cells: []WidgetCell{
							{
								Width:   12,
								Padding: 1.15,
								Text: WidgetText{
									DataRef:       "rainfall",
									Size:          12,
									ColorStyle:    "blue-500",
									Justification: "left",
								},
							},
						},
					},
					{Height: 0.75},
				},
			},
		},
	},
}

var (
	configPath string
	port       int
)

func init() {
	flag.StringVar(&configPath, "c", "config.toml", "path/to/widgets/config.toml")
	flag.IntVar(&port, "p", 10002, "port to run server")
}

func loadWidgets(configPath string) (widgets []Widget, err error) {
	var widget Widget
	_, err = toml.DecodeFile(configPath, &widget)
	if err != nil {
		return widgets, err
	}
	widget.Layouts = layouts
	widget.Data = map[string]string{"content_url": widget.WidgetURL}
	return []Widget{widget}, err
}

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

	c := time.Tick(30 * time.Second)
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
