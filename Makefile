TAG?=latest-dev
.PHONY: build
build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-swarm:$(TAG) .

.PHONY: test-unit
test-unit:
	go test -v $(go list ./... | grep -v /vendor/) -cover

.PHONY: build-armhf
build-armhf:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-swarm:$(TAG)-armhf . -f Dockerfile.armhf

.PHONY: push
push:
	docker push openfaas/faas-swarm:$(TAG)

.PHONY: all
all: build

.PHONY: ci-armhf-build
ci-armhf-build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-swarm:$(TAG)-armhf . -f Dockerfile.armhf

.PHONY: ci-armhf-push
ci-armhf-push:
	docker push openfaas/faas-swarm:$(TAG)-armhf

.PHONY: ci-arm64-build
ci-arm64-build:
	docker build --build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" -t openfaas/faas-swarm:$(TAG)-arm64 . -f Dockerfile.arm64

.PHONY: ci-arm64-push
ci-arm64-push:
	docker push openfaas/faas-swarm:$(TAG)-arm64
