// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"log"
	"time"

	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/proxy"

	"github.com/docker/docker/client"

	bootstrap "github.com/openfaas/faas-provider"
	bootTypes "github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faas-swarm/handlers"
	"github.com/openfaas/faas-swarm/types"
	"github.com/openfaas/faas-swarm/version"
)

func main() {

	var err error
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Error with Docker client: %s.", err.Error())
	}

	dockerVersion, versionErr := dockerClient.ServerVersion(context.Background())
	if versionErr != nil {
		log.Fatalf("Error with Docker server: %s", versionErr.Error())
	}

	log.Printf("Docker API version: %s, %s\n", dockerVersion.APIVersion, dockerVersion.Version)
	// How many times to reschedule a function.
	maxRestarts := uint64(5)

	// Delay between container restarts
	restartDelay := time.Second * 5

	readConfig := types.ReadConfig{}
	osEnv := types.OsEnv{}
	cfg := readConfig.Read(osEnv)

	log.Printf("HTTP Read Timeout: %s\n", cfg.ReadTimeout)
	log.Printf("HTTP Write Timeout: %s\n", cfg.WriteTimeout)

	funcProxyHandler := handlers.NewFunctionLookup(dockerClient, cfg.DNSRoundRobin)

	bootstrapHandlers := bootTypes.FaaSHandlers{
		DeleteHandler:  handlers.DeleteHandler(dockerClient),
		DeployHandler:  handlers.DeployHandler(dockerClient, maxRestarts, restartDelay),
		FunctionReader: handlers.FunctionReader(true, dockerClient),
		FunctionProxy:  proxy.NewHandlerFunc(cfg.ReadTimeout, funcProxyHandler),
		ReplicaReader:  handlers.ReplicaReader(dockerClient),
		ReplicaUpdater: handlers.ReplicaUpdater(dockerClient),
		UpdateHandler:  handlers.UpdateHandler(dockerClient, maxRestarts, restartDelay),
		HealthHandler:  handlers.Health(),
		InfoHandler:    handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommit),
		SecretHandler:  handlers.MakeSecretsHandler(dockerClient),
		LogHandler:     logs.NewLogHandlerFunc(handlers.NewLogRequester(dockerClient)),
	}

	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		TCPPort:         &cfg.TCPPort,
		EnableHealth:    true,
		EnableBasicAuth: cfg.EnableBasicAuth,
		SecretMountPath: "/run/secrets",
	}

	log.Printf("Basic authentication: %v\n", bootstrapConfig.EnableBasicAuth)

	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
}
