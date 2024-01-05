package feedback

import "time"

// Signal represents an error that needs to be fed back asynchronously
type Signal struct {
	Time  time.Time
	Error error
}

// NewSignal initialises a Signal for the current moment
func NewSignal(err error) Signal {
	return Signal{time.Now(), err}
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
	for {
		//s :=
		<-sigs
		status.Ok = false
		status.Message = "Unable to fetch latest data"
	}
}
