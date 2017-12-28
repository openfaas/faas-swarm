TAG?=latest

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t functions/faas-swarm:$(TAG) .

push:
	docker push functions/faas-swarm:$(TAG)

all: build
