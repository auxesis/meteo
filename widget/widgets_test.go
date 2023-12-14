package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWidgetsSetsContentType(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/widgets/hello", nil)
	var ws []Widget
	handleWidgetQuery(ws)(w, r)
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
			handleWidgetQuery(ws)(w, r)
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
	handleWidgetQuery(ws)(w, r)
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
	handleWidgetQuery(ws)(w, r)
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
	handleWidgetQuery(ws)(w, r)
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
