TAG?=latest-dev
.PHONY: build

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t functions/faas-swarm:$(TAG) .

test-unit:
	go test -v $(go list ./... | grep -v /vendor/) -cover

build-armhf:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t functions/faas-swarm:$(TAG)-armhf . -f Dockerfile.armhf

push:
	docker push functions/faas-swarm:$(TAG)

all: build
