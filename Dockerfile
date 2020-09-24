FROM teamserverless/license-check:0.3.6 as license-check

FROM golang:1.13 as build
ARG GO111MODULE=off
ARG CGO_ENABLED=0
COPY --from=license-check /license-check /usr/bin/

RUN mkdir -p /go/src/github.com/openfaas/faas-swarm/

WORKDIR /go/src/github.com/openfaas/faas-swarm

COPY . .

RUN license-check -path /go/src/github.com/openfaas/faas-swarm/ --verbose=false "Alex Ellis" "OpenFaaS Author(s)"

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*")

RUN CGO_ENABLED=$CGO_ENABLED go test $(go list ./... | grep -v /vendor/) -cover

RUN VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
    && GIT_COMMIT=$(git rev-list -1 HEAD) \
    && CGO_ENABLED=$CGO_ENABLED GOOS=linux go build --ldflags "-s -w \
    -X github.com/openfaas/faas-swarm/version.GitCommit=${GIT_COMMIT}\
    -X github.com/openfaas/faas-swarm/version.Version=${VERSION}" \
    -a -installsuffix cgo -o faas-swarm .

FROM alpine:3.12 as ship

LABEL org.label-schema.license="MIT" \
      org.label-schema.vcs-url="https://github.com/openfaas/faas-swarm" \
      org.label-schema.vcs-type="Git" \
      org.label-schema.name="openfaas/faas-swarm" \
      org.label-schema.vendor="openfaas" \
      org.label-schema.docker.schema-version="1.0"

RUN apk --no-cache add ca-certificates

WORKDIR /root/

EXPOSE 8080

ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/openfaas/faas-swarm/faas-swarm    .

CMD ["./faas-swarm"]
