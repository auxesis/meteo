package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics represents metrics to be exported
type Metrics struct {
	battery       prometheus.Gauge
	temperature   prometheus.Gauge
	humidity      prometheus.Gauge
	windDirection prometheus.Gauge
	windAvg       prometheus.Gauge
	windMax       prometheus.Gauge
	rain          prometheus.Gauge
}

// NewMetrics registers new metrics to export
func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		battery: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_battery",
			Help: "Current battery status of weather station.",
		}),
		temperature: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_temperature_celsius",
			Help: "Current temperature outside of house.",
		}),
		humidity: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_humidity_percentage",
			Help: "Relative humidity outside of house.",
		}),
		windDirection: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_wind_direction_degree",
			Help: "Direction of wind in degrees.",
		}),
		windAvg: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_wind_speed_average_kilometers_per_hour",
			Help: "Average wind speed in kilometers per hour.",
		}),
		windMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_wind_speed_burst_kilometers_per_hour",
			Help: "Max burst wind speed in kilometers per hour.",
		}),
		rain: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_rain_millimetres",
			Help: "Rainfall in millimeters",
		}),
	}
	reg.MustRegister(m.battery)
	reg.MustRegister(m.temperature)
	reg.MustRegister(m.humidity)
	reg.MustRegister(m.windDirection)
	reg.MustRegister(m.windAvg)
	reg.MustRegister(m.windMax)
	reg.MustRegister(m.rain)

	return m
}

func readMeasurementLoop(metrics *Metrics) {
	opts := mqtt.NewClientOptions()
	opts.SetClientID("misol_weather_station_rtl433_exporter")
	opts.AddBroker("tcp://localhost:1883")
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	readMeasurement := measurementReader(metrics)

	topic := "sensors/rtl_433/#"
	token := client.Subscribe(topic, 1, readMeasurement)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

// Measurements tracks the types of measurements we receive
var Measurements = map[string]string{
	"time":          "time",    // "2023-02-23 11:22:24"
	"model":         "string",  // "Fineoffset-WHx080"
	"id":            "int",     // 240
	"battery_ok":    "bool",    // 1
	"temperature_C": "decimal", // 20.800
	"humidity":      "int",     // 68
	"wind_dir_deg":  "int",     // 135
	"wind_avg_km_h": "decimal", // 1.224
	"wind_max_km_h": "decimal", // 2.448
	"rain_mm":       "decimal", // 70.200
}

func measurementReader(metrics *Metrics) func(mqtt.Client, mqtt.Message) {
	return func(c mqtt.Client, msg mqtt.Message) {
		parts := strings.Split(msg.Topic(), "/")
		name := parts[len(parts)-1]
		fmt.Printf("topic: %s, payload: %s\n", msg.Topic(), msg.Payload())
		switch name {
		case "battery_ok":
			float, err := strconv.ParseFloat(string(msg.Payload()), 64)
			if err != nil {
				fmt.Printf("error: unable to parse float for temperature_C: %s", err)
				return
			}
			metrics.battery.Set(float)
		case "temperature_C":
			float, err := strconv.ParseFloat(string(msg.Payload()), 64)
			if err != nil {
				fmt.Printf("error: unable to parse float for temperature_C: %s", err)
				return
			}
			metrics.temperature.Set(float)
		case "humidity":
			float, err := strconv.ParseFloat(string(msg.Payload()), 64)
			if err != nil {
				fmt.Printf("error: unable to parse float for temperature_C: %s", err)
				return
			}
			metrics.humidity.Set(float)
		case "wind_dir_deg":
			float, err := strconv.ParseFloat(string(msg.Payload()), 64)
			if err != nil {
				fmt.Printf("error: unable to parse float for temperature_C: %s", err)
				return
			}
			metrics.windDirection.Set(float)
		case "wind_avg_km_h":
			float, err := strconv.ParseFloat(string(msg.Payload()), 64)
			if err != nil {
				fmt.Printf("error: unable to parse float for wind_avg_km_h: %s", err)
				return
			}
			metrics.windAvg.Set(float)
		case "wind_max_km_h":
			float, err := strconv.ParseFloat(string(msg.Payload()), 64)
			if err != nil {
				fmt.Printf("error: unable to parse float for wind_max_km_h: %s", err)
				return
			}
			metrics.windMax.Set(float)
		case "rain_mm":
			float, err := strconv.ParseFloat(string(msg.Payload()), 64)
			if err != nil {
				fmt.Printf("error: unable to parse float for rain_mm: %s", err)
				return
			}
			metrics.rain.Set(float)
		}
	}
}

func main() {
	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	//mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	// Create a non-global registry.
	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	metrics := NewMetrics(reg)

	// Read measurements via MQTT, update metrics
	go readMeasurementLoop(metrics)

	// Expose metrics and custom registry via an HTTP server
	// using the HandleFor function. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(":10000", nil))
}
