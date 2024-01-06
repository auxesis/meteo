package feedback

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeedbackOnlyErrorsOnFullFailure(t *testing.T) {
	assert := assert.New(t)

	status := Status{Ok: true}
	metrics := map[string]Signal{
		"temperature": Signal{Ok: true},
		"humidity":    Signal{Ok: true},
		"rain":        Signal{Ok: true},
		"wind":        Signal{Ok: true},
	}

	type signal struct {
		metric string
		err    error
	}

	testsigs := []signal{
		{"temperature", errors.New("server error: 502")},
		{"humidity", errors.New("server error: 502")},
		{"rain", errors.New("server error: 502")},
		{"wind", errors.New("server error: 502")},
	}

	for i, s := range testsigs {
		s := NewSignalWithError(s.metric, s.err)
		handleSignal(&status, &metrics, s)

		if i < 3 {
			assert.True(status.Ok)
		} else {
			assert.False(status.Ok)
		}
	}

}
