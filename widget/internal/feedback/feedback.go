package feedback

import (
	"time"
)

// Signal represents an status that needs to be fed back asynchronously
type Signal struct {
	Time   time.Time
	Ok     bool
	Metric string
	Error  error
}

// NewSignalWithError initialises a Signal for the current moment with an error
func NewSignalWithError(metric string, err error) Signal {
	return Signal{time.Now(), false, metric, err}
}

// NewSignal initialises a Signal for the current moment with no error
func NewSignal(metric string) Signal {
	return Signal{time.Now(), true, metric, nil}
}

// Status represents the current status of polling data sources
type Status struct {
	Ok      bool
	Message string
}

// ProcessSignals looks at signals from data collectors and updates the status.
//
// The status is used by the HTTP endpoint when rendering responses.
//
// The only data collector right now is Prometheus.
func ProcessSignals(sigs chan Signal, status *Status) {
	metrics := map[string]Signal{}
	for {
		s := <-sigs
		handleSignal(status, &metrics, s)
	}
}

func handleSignal(status *Status, metrics *map[string]Signal, s Signal) {
	(*metrics)[s.Metric] = s

	failCount := 0
	for _, v := range *metrics {
		if !v.Ok {
			failCount++
		}
	}

	if failCount == len(*metrics) {
		status.Ok = false
		status.Message = "Unable to fetch latest data"
	} else {
		status.Ok = true
		status.Message = ""
	}
}
