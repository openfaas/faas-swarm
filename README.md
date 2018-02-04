faas-swarm
==========

[![Build Status](https://travis-ci.org/openfaas/faas-swarm.svg?branch=master)](https://travis-ci.org/openfaas/faas-swarm)

## Summary

This is the Docker Swarm provider for OpenFaaS.

The first and canonical implementation of OpenFaaS was tightly coupled to Docker Swarm. This repository aims to decouple the two so that all providers are isomorphic and symmetrical.

## Status

Status: Released

Features:

* [x] Create
* [x] Proxy
* [x] Update
* [x] Delete
* [x] List
* [x] Scale

Docker image: [`functions/faas-swarm`](https://hub.docker.com/r/functions/faas-swarm/tags/)

## Contributing

The contribution guide applies from the [main OpenFaaS repository](https://github.com/openfaas/faas/blob/master/CONTRIBUTING.md).
