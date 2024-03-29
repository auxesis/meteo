package http

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"

	"github.com/auxesis/meteo/widget/internal/feedback"
	"github.com/auxesis/meteo/widget/internal/widget"
	"github.com/shopspring/decimal"
)

// Samples is a map of the latest metric samples (currently from Prometheus)
type Samples map[string]float64

// HandleWidgetQuery handles rendering a widget in the WCS widget.json format
func HandleWidgetQuery(wdgts []widget.Widget, samples *Samples, status *feedback.Status) func(w http.ResponseWriter, r *http.Request) {
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

		var wdgt widget.Widget
		for _, wi := range wdgts {
			if wi.ID == id && wi.Token == t {
				wdgt = wi
				break
			}
		}
		if len(wdgt.ID) == 0 && len(wdgt.Token) == 0 {
			log.Printf("error: unable to find widget for ID \"%s\" and token \"%s\"", id, t)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if status.Ok {
			wdgt.Layouts = widget.WeatherLayout
			wdgt = addDataFromSamples(wdgt, samples)
			wdgt = adjustColorsFromThresholds(wdgt, samples)
		} else {
			wdgt.Layouts = widget.ErrorLayout
			wdgt = addDataFromFeedback(wdgt, status)
		}
		err := json.NewEncoder(w).Encode(wdgt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// addDataFromSamples populates a widget's data with the latest samples
func addDataFromSamples(w widget.Widget, s *Samples) widget.Widget {
	for k, c := range w.Metrics {
		f := (*s)[k]
		var v decimal.Decimal
		if math.IsNaN(f) {
			log.Printf("warning: %s is NaN, returning -1\n", k)
			v = decimal.New(-1, 0)
		} else {
			v = decimal.NewFromFloat(f)
		}

		var vs string
		if v.Exponent() > -2 {
			vs = v.String()
		} else {
			vs = v.RoundDown(1).String()
		}
		w.Data[k] = fmt.Sprintf("%s%s", vs, c.DisplayUnit)
	}
	return w
}

// addDataFromFeedback populates a widget's data with the latest feedback
func addDataFromFeedback(w widget.Widget, s *feedback.Status) widget.Widget {
	if s.Ok {
		delete(w.Data, "status")
	} else {
		w.Data["status"] = s.Message
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
func adjustColorsFromThresholds(w widget.Widget, s *Samples) widget.Widget {
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
