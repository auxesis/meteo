package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWidgetsSetsContentType(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/widgets/hello", nil)
	var ws []Widget
	var s Samples
	handleWidgetQuery(ws, &s)(w, r)
	res := w.Result()
	assert.Equal(res.Header.Get("Content-Type"), "application/json")
}

func TestWidgetsLookupByIDAndReturnsNotFound(t *testing.T) {
	assert := assert.New(t)

	type test struct {
		url    string
		widget Widget
		status int
	}
	tests := []test{
		{"http://a.test/widgets/not_found", Widget{}, http.StatusNotFound},
		{"http://a.test/widgets/unauthorized?token=right", Widget{ID: "unauthorized", Token: "wrong"}, http.StatusNotFound},
		{"http://a.test/widgets/authorized?token=s3cr3t", Widget{ID: "authorized", Token: "s3cr3t"}, http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tc.url, nil)
			ws := []Widget{tc.widget}
			var s Samples
			handleWidgetQuery(ws, &s)(w, r)
			res := w.Result()
			assert.Equal(tc.status, res.StatusCode)
		})
	}
}

func TestWidgetsHasData(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://a.test/widgets/sydney?token=s3cr3t", nil)
	ws, err := loadWidgets("testdata/config.toml")
	assert.NoError(err)
	var s Samples
	handleWidgetQuery(ws, &s)(w, r)
	res := w.Result()

	body, err := io.ReadAll(res.Body)
	assert.NoError(err)

	var widget Widget
	err = json.Unmarshal(body, &widget)
	assert.NoError(err)
	assert.NotEmpty(widget.Name)
	assert.NotEmpty(widget.Description)
	assert.NotEmpty(widget.Data)
	for _, k := range []string{"content_url", "temperature", "humidity", "wind_gust", "rainfall"} {
		t.Run(k, func(t *testing.T) {
			assert.NotEmpty(widget.Data[k])
		})
	}
}

func TestWidgetsHasColours(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://a.test/widgets/sydney?token=s3cr3t", nil)
	ws, err := loadWidgets("testdata/config.toml")
	assert.NoError(err)
	var s Samples
	handleWidgetQuery(ws, &s)(w, r)
	res := w.Result()

	body, err := io.ReadAll(res.Body)
	assert.NoError(err)

	var widget Widget
	err = json.Unmarshal(body, &widget)
	assert.NoError(err)
	assert.NotEmpty(widget.Layouts)

	for _, l := range widget.Layouts {
		assert.NotEmpty(l.Size)
		assert.NotEmpty(l.Styles)
		assert.NotEmpty(l.Styles.Colors)
		for _, v := range l.Styles.Colors {
			assert.NotEmpty(v.Color)
		}
	}
}

func TestWidgetsUsesDataRefs(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://a.test/widgets/sydney?token=s3cr3t", nil)
	ws, err := loadWidgets("testdata/config.toml")
	assert.NoError(err)
	var s Samples
	handleWidgetQuery(ws, &s)(w, r)
	res := w.Result()

	body, err := io.ReadAll(res.Body)
	assert.NoError(err)

	var widget Widget
	err = json.Unmarshal(body, &widget)
	assert.NoError(err)
	assert.NotEmpty(widget.Layouts)

	var dataRefs []string
	for _, lyts := range widget.Layouts {
		for _, lyrs := range lyts.Layers {
			assert.NotEmpty(lyrs)
			for _, r := range lyrs.Rows {
				assert.NotEmpty(r)
				for _, c := range r.Cells {
					if len(c.Text.DataRef) > 0 {
						dataRefs = append(dataRefs, c.Text.DataRef)
					}
				}
			}
		}
	}

	var data []string
	for k := range widget.Data {
		if k != "content_url" {
			data = append(data, k)
		}
	}

	assert.ElementsMatch(dataRefs, data)
}

func TestWidgetsUsesLatestSamples(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://a.test/widgets/sydney?token=s3cr3t", nil)
	ws, err := loadWidgets("testdata/config.toml")
	assert.NoError(err)
	s := Samples{"temperature": 30.2, "humidity": 50, "rainfall": 1.2, "wind_gust": 3.6}
	handleWidgetQuery(ws, &s)(w, r)
	res := w.Result()

	body, err := io.ReadAll(res.Body)
	assert.NoError(err)

	var widget Widget
	err = json.Unmarshal(body, &widget)
	assert.NoError(err)
	assert.NotEmpty(widget.Layouts)
	assert.Greater(len(widget.Data), 1)
	assert.NotEmpty(widget.Data["content_url"])
	for k, v := range widget.Data {
		if k != "content_url" {
			vs := regexp.MustCompile(`\d+.?\d+?`).FindString(v)
			d, err := strconv.ParseFloat(vs, 64)
			assert.NoError(err)
			assert.Equal(s[k], d)
		}
	}
}

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

func TestWidgetsUsesColorsForThresholds(t *testing.T) {
	assert := assert.New(t)
	ws, err := loadWidgets("testdata/config.toml")
	assert.NoError(err)

	var tests = []struct {
		metric string
		value  float64
		expect string
	}{
		{"temperature", float64(ws[0].Metrics["temperature"].Levels["base"] + 1), "blue-500"},
		{"temperature", float64(ws[0].Metrics["temperature"].Levels["low"] + 1), "green-500"},
		{"temperature", float64(ws[0].Metrics["temperature"].Levels["medium"] + 1), "yellow-500"},
		{"temperature", float64(ws[0].Metrics["temperature"].Levels["high"] + 1), "red-500"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s/%f/%s", tc.metric, tc.value, tc.expect), func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://a.test/widgets/sydney?token=s3cr3t", nil)
			s := Samples{tc.metric: tc.value, "humidity": 0, "rainfall": 0, "wind_gust": 0}
			handleWidgetQuery(ws, &s)(w, r)
			res := w.Result()

			body, err := io.ReadAll(res.Body)
			assert.NoError(err)
			var widget Widget
			err = json.Unmarshal(body, &widget)
			assert.NoError(err)
			assert.NotEmpty(widget.Layouts)

			var color string
			for _, lyts := range widget.Layouts {
				for _, lyrs := range lyts.Layers {
					assert.NotEmpty(lyrs)
					for _, r := range lyrs.Rows {
						assert.NotEmpty(r)
						for _, c := range r.Cells {
							if c.Text.DataRef == tc.metric {
								color = c.Text.ColorStyle
							}
						}
					}
				}
			}
			assert.NotEmpty(color)
			assert.Equal(tc.expect, color)
		})
	}
}
