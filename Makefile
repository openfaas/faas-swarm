TAG?=latest-dev
.PHONY: build test-unit build-armhf push all ci-armhf-build ci-armhf-push ci-arm64-build ci-arm64-push

build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-swarm:$(TAG) .

test-unit:
	go test -v $(go list ./... | grep -v /vendor/) -cover

build-armhf:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-swarm:$(TAG)-armhf . -f Dockerfile.armhf

push:
	docker push openfaas/faas-swarm:$(TAG)

all: build

ci-armhf-build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-swarm:$(TAG)-armhf . -f Dockerfile.armhf

ci-armhf-push:
	docker push openfaas/faas-swarm:$(TAG)-armhf

ci-arm64-build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-swarm:$(TAG)-arm64 . -f Dockerfile.arm64

ci-arm64-push:
	docker push openfaas/faas-swarm:$(TAG)-arm64
