
build:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o qingping_air_monitor_lite_exporter-linux-arm5
	env GOOS=darwin GOARCH=amd64 go build -o qingping_air_monitor_lite_exporter-macos-amd64

all: test

test: gotest goerrcheck gostaticcheck

gotest:
	go test ./... -v -timeout=45s -failfast

goerrcheck:
	errcheck -exclude .errcheck-excludes -ignoretests ./...

gostaticcheck:
	staticcheck ./...

