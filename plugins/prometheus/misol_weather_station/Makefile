all: test

build:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o misol_weather_station_rtl433_exporter-linux-arm5
	env GOOS=darwin GOARCH=amd64 go build -o misol_weather_station_rtl433_exporter-macos-amd64

test:
	go test ./...
