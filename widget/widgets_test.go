package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindingColorForValue(t *testing.T) {
	assert := assert.New(t)

	levels := map[string]int{"base": 0, "low": 10, "medium": 20, "high": 30}
	var tests = []struct {
		value  float64
		levels map[string]int
		expect string
	}{
		{0, levels, "blue-500"},
		{1, levels, "blue-500"},
		{10, levels, "green-500"},
		{11, levels, "green-500"},
		{20, levels, "yellow-500"},
		{21, levels, "yellow-500"},
		{30, levels, "red-500"},
		{31, levels, "red-500"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%.1f", tc.value), func(t *testing.T) {
			c := findColorForValue(tc.value, tc.levels)
			assert.Equal(tc.expect, c)
		})
	}
}

func TestValueFormatting(t *testing.T) {
	assert := assert.New(t)
	ws, err := loadWidgets("testdata/config.toml")
	assert.NoError(err)
	wdgt := ws[0]

	var tests = []struct {
		metric string
		value  float64
		expect string
	}{
		{"temperature", 10.0, "10°"},
		{"temperature", 10.1, "10.1°"},
		{"temperature", 10.23, "10.2°"},
		{"temperature", 10.44, "10.4°"},
		{"temperature", 10.45, "10.4°"},
		{"temperature", 10.456, "10.4°"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s/%.4f", tc.metric, tc.value), func(t *testing.T) {
			samples := Samples{tc.metric: tc.value}
			w := addDataFromSamples(wdgt, &samples)
			assert.Equal(tc.expect, w.Data[tc.metric])
		})
	}
}
