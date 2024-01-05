package widget

import (
	"github.com/BurntSushi/toml"
)

// Widget is a container for a widget.json-formatted response, suitable for WCS
type Widget struct {
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	Data          map[string]string       `json:"data"`
	Layouts       map[string]Layout       `json:"layouts"`
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

// Layout is a layout for a widget.json widget
type Layout struct {
	Size   string  `json:"size"`
	Styles Styles  `json:"styles"`
	Layers []Layer `json:"layers"`
}

// Styles is a collection of styles for a widget.json widget
type Styles struct {
	Colors map[string]Color `json:"colors"`
}

// Color is a color for a widget.json widget
type Color struct {
	Color string `json:"color"`
}

// Layer is a layer for a widget.json widget
type Layer struct {
	Rows []Row `json:"rows"`
}

// Row is a layer row for a widget.json widget
type Row struct {
	Height float64 `json:"height"`
	Cells  []Cell  `json:"cells,omitempty"`
}

// Cell is a cell for layer row for a widget.json widget
type Cell struct {
	Width                int     `json:"width"`
	BackgroundColorStyle string  `json:"background_color_style,omitempty" toml:"background_color_style"`
	Padding              float64 `json:"padding,omitempty"`
	Text                 Text    `json:"text,omitempty"`
	LinkURL              string  `json:"link_url,omitempty"`
}

// Text is a text object, for a cell, for layer row, for a widget.json widget
type Text struct {
	String         string  `json:"string,omitempty"`
	DataRef        string  `json:"data_ref,omitempty"`
	Size           float64 `json:"size,omitempty"`
	ColorStyle     string  `json:"color_style,omitempty"`
	FontStyle      string  `json:"font_style,omitempty"`
	Weight         string  `json:"weight,omitempty"`
	Justification  string  `json:"justification,omitempty"`
	MinScaleFactor float64 `json:"min_scale_factor,omitempty"`
}

// ErrorLayout is the layout used to render error messages
var ErrorLayout = map[string]Layout{
	"error": Layout{
		Size: "small",
		Styles: Styles{
			Colors: map[string]Color{
				"black":      Color{Color: "#000000"},
				"stone-100":  Color{Color: "#f5f5f4"},
				"stone-950":  Color{Color: "#0c0a09"},
				"blue-500":   Color{Color: "#3b82f6"},
				"green-500":  Color{Color: "#84cc16"},
				"yellow-500": Color{Color: "#facc15"},
				"red-500":    Color{Color: "#ef4444"},
			},
		},
		Layers: []Layer{
			Layer{
				Rows: []Row{{
					Height: 12,
					Cells: []Cell{{
						Width:                12,
						BackgroundColorStyle: "stone-950",
					}},
				}},
			},
			Layer{
				Rows: []Row{
					{Height: 1},
					{
						Height: 2.75,
						Cells: []Cell{
							{
								Width:   12,
								Padding: 1.15,
								Text: Text{
									DataRef:       "status",
									Size:          8,
									ColorStyle:    "yellow-500",
									Weight:        "bold",
									Justification: "left",
								},
							},
						},
					},
				},
			},
		},
	},
}

// WeatherLayout is the layout used to render the weather widget
var WeatherLayout = map[string]Layout{
	"weather_small": Layout{
		Size: "small",
		Styles: Styles{
			Colors: map[string]Color{
				"black":      Color{Color: "#000000"},
				"stone-100":  Color{Color: "#f5f5f4"},
				"stone-950":  Color{Color: "#0c0a09"},
				"blue-500":   Color{Color: "#3b82f6"},
				"green-500":  Color{Color: "#84cc16"},
				"yellow-500": Color{Color: "#facc15"},
				"red-500":    Color{Color: "#ef4444"},
			},
		},
		Layers: []Layer{
			Layer{
				Rows: []Row{{
					Height: 12,
					Cells: []Cell{{
						Width:                12,
						BackgroundColorStyle: "stone-950",
					}},
				}},
			},
			Layer{
				Rows: []Row{
					{Height: 1},
					{
						Height: 2.75,
						Cells: []Cell{
							{
								Width:   12,
								Padding: 1.15,
								Text: Text{
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
						Cells: []Cell{
							{
								Width:   12,
								Padding: 1.15,
								Text: Text{
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
						Cells: []Cell{
							{
								Width:   12,
								Padding: 1.15,
								Text: Text{
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
						Cells: []Cell{
							{
								Width:   12,
								Padding: 1.15,
								Text: Text{
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

// LoadWidgets loads a widget from a file path
func LoadWidgets(configPath string) (widgets []Widget, err error) {
	var widget Widget
	_, err = toml.DecodeFile(configPath, &widget)
	if err != nil {
		return widgets, err
	}
	widget.Data = map[string]string{"content_url": widget.WidgetURL}
	return []Widget{widget}, err
}
