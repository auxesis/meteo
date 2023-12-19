package main

import (
	"github.com/BurntSushi/toml"
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
