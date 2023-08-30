package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

// readMeasurementLoop sets up a mqtt client, reads measurements from a topic, and updates exported metrics
func readMeasurementLoop(metrics *Metrics, host string, port int, ttl time.Duration) {
	opts := mqtt.NewClientOptions()
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	pid := os.Getpid()

	clientID := fmt.Sprintf("misol_weather_station_rtl433_exporter-%s-%d", hostname, pid)
	fmt.Printf("Connecting with ClientID: %s\n", clientID)
	opts.SetClientID(clientID)
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	opts.SetPingTimeout(1 * time.Second)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetOrderMatters(false)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	// Set up the TTL checker early, in case MQTT is unavailable
	refresh := make(chan time.Time)
	readMeasurement := measurementReader(metrics, refresh)
	exit := ttl * 10
	go nilIfTTLExpired(metrics, refresh, ttl, exit)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic := "sensors/rtl_433/#"
	if token := client.Subscribe(topic, 1, readMeasurement); token.Wait() && token.Error() != nil {
		fmt.Printf("error: %s", token.Error())
		os.Exit(1)
	}
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

func rateLimitedPrintln(s string, d time.Duration) {
	last := rateLimitedPrintlnTable[s]
	now := time.Now()
	if now.Sub(last) > d {
		fmt.Println(s)
		rateLimitedPrintlnTable[s] = now
	}
}

// nilIfTTLExpired nils out a metric if an update isn't received within a timeout
func nilIfTTLExpired(metrics *Metrics, refresh chan time.Time, ttl time.Duration, exit time.Duration) {
	NaN := math.Log(-1.0)
	last := time.Now()

	// read for updates
	go func() {
		for t := range refresh {
			last = t
		}
	}()

	// check if the TTL has expired
	ticker := time.NewTicker(time.Second)
	for {
		now := <-ticker.C
		if now.Sub(last) > ttl {
			rateLimitedPrintln("error: TTL expired on last measurement - setting all measurements to NaN", 30*time.Second)
			metrics.battery.Set(NaN)
			metrics.temperature.Set(NaN)
			metrics.humidity.Set(NaN)
			metrics.windDirection.Set(NaN)
			metrics.windAvg.Set(NaN)
			metrics.windMax.Set(NaN)
			metrics.rain.Set(NaN)
		}
		if now.Sub(last) > exit {
			fmt.Printf("error: no updates for %s - exiting\n", exit)
			os.Exit(2)
		}
	}
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

// measurementReader returns a function to be used as a callback when messages are received in client.Subscribe
func measurementReader(metrics *Metrics, refresh chan time.Time) func(mqtt.Client, mqtt.Message) {
	return func(c mqtt.Client, msg mqtt.Message) {
		parts := strings.Split(msg.Topic(), "/")
		name := parts[len(parts)-1]
		fmt.Printf("topic: %s, payload: %s\n", msg.Topic(), msg.Payload())
		refresh <- time.Now()
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

var (
	host                    string
	port                    int
	debug                   bool
	ttl                     time.Duration
	rateLimitedPrintlnTable map[string]time.Time
)

func init() {
	flag.StringVar(&host, "h", "[::1]", "hostname/address of MQTT broker")
	flag.IntVar(&port, "p", 1883, "tcp port of MQTT broker")
	flag.BoolVar(&debug, "d", false, "turn on debug output")
	flag.DurationVar(&ttl, "t", 10*time.Minute, "how long to wait for updates before returning NaNs")
	rateLimitedPrintlnTable = make(map[string]time.Time)
}

func main() {
	flag.Parse()

	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	//mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	// Create a non-global registry.
	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	metrics := NewMetrics(reg)

	// Read measurements via MQTT, update metrics
	go readMeasurementLoop(metrics, host, port, ttl)

	// Expose metrics and custom registry via an HTTP server
	// using the HandleFor function. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(":10000", nil))
}
