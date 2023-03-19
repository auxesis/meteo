.PHONY: all plan apply destroy

all: build

build:
	ansible-playbook --inventory=hosts ansible/bootstrap-graphs.yaml

#
# For local development
#

export PATH := $(shell pwd)/vendor/python/bin:$(PATH)

build_droplet_from_latest_image:
	ansible-playbook test.yml

dev:
	rsync -auv --delete -e ssh plugins $(shell cat hosts):src/github.com/auxesis/meteo/
#	ssh $(shell cat hosts) "cd src/github.com/auxesis/meteo/plugins && make build-linux-arm"
