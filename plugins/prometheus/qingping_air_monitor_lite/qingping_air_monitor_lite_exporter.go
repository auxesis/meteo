package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics represents metrics to be exported
type Metrics struct {
	temperature prometheus.Gauge
	humidity    prometheus.Gauge
	co2         prometheus.Gauge
	pm25        prometheus.Gauge
	pm10        prometheus.Gauge
}

// NewMetrics registers new metrics to export
func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		temperature: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "upstairs_temperature_celsius",
			Help: "Current room temperature.",
		}),
		humidity: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "upstairs_humidity_percentage",
			Help: "Relative humidity in room.",
		}),
		co2: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "upstairs_co2_parts_per_million",
			Help: "Carbon dioxide in parts per million.",
		}),
		pm25: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "upstairs_pm25_micrograms_per_meter_cubed",
			Help: "PM2.5 µg/m3 averaged over 1 hour.",
		}),
		pm10: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "upstairs_pm10_micrograms_per_meter_cubed",
			Help: "PM10 µg/m3 averaged over 1 hour.",
		}),
	}
	reg.MustRegister(m.temperature)
	reg.MustRegister(m.humidity)
	reg.MustRegister(m.co2)
	reg.MustRegister(m.pm25)
	reg.MustRegister(m.pm10)

	return m
}

func readMeasurementLoop(metrics *Metrics, host string, port int, mac string, ttl time.Duration) {
	opts := mqtt.NewClientOptions()
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	pid := os.Getpid()

	clientID := fmt.Sprintf("qingping_air_monitor_lite_exporter-%s-%d", hostname, pid)
	log.Printf("Connecting with ClientID: %s\n", clientID)
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

	// Set up the client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic := fmt.Sprintf("/+/%s/#", mac)
	if token := client.Subscribe(topic, 1, readMeasurement); token.Wait() && token.Error() != nil {
		fmt.Printf("error: %s", token.Error())
		os.Exit(1)
	}
	log.Printf("Subscribed to topic: %s\n", topic)
}

// QingpingMQTTMsg represents a MQTT message from the Qingping Air Monitor Lite sensor
type QingpingMQTTMsg struct {
	Type       string `json:"type"`
	MAC        string `json:"mac"`
	Timestamp  int    `json:"timestamp"`
	SensorData []QingpingSensorData
}

// QingpingSensorData represents sensor data in an MQTT message (type 17) from the Qingping Air Monitor Lite sensor
type QingpingSensorData struct {
	Timestamp   QingpingIntValue   `json:"timestamp"`
	Temperature QingpingFloatValue `json:"temperature"`
	Humidity    QingpingFloatValue `json:"humidity"`
	CO2         QingpingFloatValue `json:"co2"`
	PM25        QingpingFloatValue `json:"pm25"`
	PM10        QingpingFloatValue `json:"pm10"`
}

// QingpingIntValue represents an integer value, found in sensor data in an MQTT message (type 17) from the Qingping Air Monitor Lite sensor
type QingpingIntValue struct {
	Value int `json:"value"`
}

// QingpingFloatValue represents an float value, found in sensor data in an MQTT message (type 17) from the Qingping Air Monitor Lite sensor
type QingpingFloatValue struct {
	Value float64 `json:"value"`
}

func measurementReader(metrics *Metrics, refresh chan time.Time) func(mqtt.Client, mqtt.Message) {
	return func(c mqtt.Client, msg mqtt.Message) {
		if debug {
			fmt.Printf("topic: %s, payload: %s\n", msg.Topic(), msg.Payload())
		}

		var qmsg QingpingMQTTMsg
		err := json.Unmarshal(msg.Payload(), &qmsg)
		if err != nil {
			log.Fatalf("error: unable to decode JSON: %s", err)
		}

		// Ignore everything but messages with sensorData
		if qmsg.Type != "17" {
			log.Printf("ignoring message (type %s) without sensorData\n", qmsg.Type)
			return
		}
		log.Printf("got sensorData")
		refresh <- time.Now()

		if len(qmsg.SensorData) > 1 {
			log.Printf("warning: multiple sensorData received: expected 1, got %d", len(qmsg.SensorData))
		}

		metrics.temperature.Set(qmsg.SensorData[0].Temperature.Value)
		metrics.humidity.Set(qmsg.SensorData[0].Humidity.Value)
		metrics.co2.Set(qmsg.SensorData[0].CO2.Value)
		metrics.pm25.Set(qmsg.SensorData[0].PM25.Value)
		metrics.pm10.Set(qmsg.SensorData[0].PM10.Value)
	}
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
			metrics.temperature.Set(NaN)
			metrics.humidity.Set(NaN)
			metrics.co2.Set(NaN)
			metrics.pm25.Set(NaN)
			metrics.pm10.Set(NaN)
		}
		if now.Sub(last) > exit {
			fmt.Printf("error: no updates for %s - exiting\n", exit)
			os.Exit(2)
		}
	}
}

var (
	host                    string
	port                    int
	mac                     string
	debug                   bool
	ttl                     time.Duration
	rateLimitedPrintlnTable map[string]time.Time
)

func init() {
	flag.StringVar(&host, "h", "[::1]", "hostname/address of MQTT broker")
	flag.IntVar(&port, "p", 1883, "tcp port of MQTT broker")
	flag.StringVar(&mac, "m", "", "MAC address of Qingping device")
	flag.BoolVar(&debug, "d", false, "turn on debug output")
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
	go readMeasurementLoop(metrics, host, port, mac, ttl)

	// Expose metrics and custom registry via an HTTP server
	// using the HandleFor function. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(":10001", nil))
}
