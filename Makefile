TAG?=latest-dev

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t functions/faas-swarm:$(TAG) .

build-armhf:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t functions/faas-swarm:$(TAG) . -f Dockerfile.armhf

push:
	docker push functions/faas-swarm:$(TAG)

all: build
