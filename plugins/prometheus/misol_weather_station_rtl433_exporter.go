package main

import (
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics represents metrics to be exported
type Metrics struct {
	temperature   prometheus.Gauge
	humidity      prometheus.Gauge
	windDirection prometheus.Gauge
	windAvg       prometheus.Gauge
	windMax       prometheus.Gauge
	rain          prometheus.Counter
	uvi           prometheus.Gauge
	lightLux      prometheus.Gauge
}

// NewMetrics registers new metrics to export
func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
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
			Name: "outdoor_wind_speed_average_meters_per_second",
			Help: "Average wind speed in meters per second.",
		}),
		windMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_wind_speed_average_max_per_second",
			Help: "Max wind speed in meters per second.",
		}),
		rain: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outdoor_rain_millimetres",
			Help: "Rainfall",
		}),
		uvi: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_uv_index_level",
			Help: "Ultraviolet radiaton. Each point is 25 milliWatts/square metre of UV radiation.",
		}),
		lightLux: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outdoor_light_lux",
			Help: "Light level intensity.",
		}),
	}
	reg.MustRegister(m.temperature)
	reg.MustRegister(m.humidity)
	reg.MustRegister(m.windDirection)
	reg.MustRegister(m.windAvg)
	reg.MustRegister(m.windMax)
	reg.MustRegister(m.rain)
	reg.MustRegister(m.uvi)
	reg.MustRegister(m.lightLux)

	return m
}

// Measurement represents a reading from the weather station
type Measurement struct {
	time          time.Time `json:"time"`
	model         string    `json:"model"`
	id            int       `json:"id"`
	batteryOK     bool      `json:"battery_ok"`
	temperatureC  big.Float `json:"temperature_C"`
	humidity      int       `json:"humidity"`
	windDirection int       `json:"wind_dir_deg"`
	windAvg       big.Float `json:"wind_avg_m_s"`
	windMax       big.Float `json:"wind_avg_m_s"`
	rain          big.Float `json:"rain_mm"`
	uvi           int       `json:"uvi"`
	lightlux      big.Float `json:"light_lux"`
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
	"time":          "time",
	"model":         "string",
	"id":            "int",
	"battery_ok":    "bool",
	"temperature_C": "decimal",
	"humidity":      "int",
	"wind_dir_deg":  "int",
	"wind_avg_km_h": "decimal",
	"wind_max_km_h": "decimal",
	"rain_mm":       "decimal",
}

func measurementReader(metrics *Metrics) func(mqtt.Client, mqtt.Message) {
	return func(c mqtt.Client, msg mqtt.Message) {
		fmt.Printf("message: %+v\n", msg)
		fmt.Printf("topic: %s\npayload: %s\n", msg.Topic(), msg.Payload())

		parts := strings.Split(msg.Topic(), "/")
		name := parts[len(parts)-1]
		fmt.Printf("name: %+v\n", name)
		fmt.Printf("type: %+v\n", Measurements[name])

		//		metrics.temperature.Set(32.0)
		metrics.temperature.Set(32.0)
		metrics.rain.Inc()
	}
}

func main() {
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
