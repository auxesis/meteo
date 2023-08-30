all: test

build:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o misol_weather_station_rtl433_exporter-linux-arm5
	env GOOS=darwin GOARCH=amd64 go build -o misol_weather_station_rtl433_exporter-macos-amd64

test: gotest goerrcheck gostaticcheck

gotest:
	go test ./... -v -timeout=45s -failfast

goerrcheck:
	errcheck -exclude .errcheck-excludes -ignoretests ./...

gostaticcheck:
	staticcheck ./...
