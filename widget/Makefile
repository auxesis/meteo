all: test

setup:
	cargo install taplo-cli
	npm install prettier prettier-plugin-toml --save-dev --save-exact

build:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o weather_widget-linux-arm5
	env GOOS=linux GOARCH=amd64 GOARM=5 go build -o weather_widget-linux-amd64
	env GOOS=darwin GOARCH=amd64 go build -o weather_widget-macos-amd64

test: gotest goerrcheck gostaticcheck

gotest:
	go test ./... -v -timeout=45s -failfast

goerrcheck:
	errcheck -exclude .errcheck-excludes -ignoretests ./...

gostaticcheck:
	staticcheck ./...
