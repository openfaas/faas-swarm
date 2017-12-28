// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/client"

	"github.com/openfaas/faas-provider"
	bootTypes "github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faas-swarm/handlers"
)

func main() {

	var err error
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		log.Fatal("Error with Docker client.")
	}
	fmt.Println(dockerClient)
	dockerVersion, versionErr := dockerClient.ServerVersion(context.Background())
	if versionErr != nil {
		log.Fatal("Error with Docker server.\n", err)
	}
	log.Printf("Docker API version: %s, %s\n", dockerVersion.APIVersion, dockerVersion.Version)
	// How many times to reschedule a function.
	maxRestarts := uint64(5)

	// Delay between container restarts
	restartDelay := time.Second * 5

	bootstrapHandlers := bootTypes.FaaSHandlers{
		FunctionProxy:  handlers.FunctionProxy(true, dockerClient),
		DeleteHandler:  handlers.DeleteHandler(dockerClient),
		DeployHandler:  handlers.DeployHandler(dockerClient, maxRestarts, restartDelay),
		FunctionReader: handlers.FunctionReader(true, dockerClient),
		ReplicaReader:  handlers.ReplicaReader(dockerClient),
		ReplicaUpdater: handlers.ReplicaUpdater(dockerClient),
		UpdateHandler:  handlers.UpdateHandler(dockerClient, maxRestarts, restartDelay),
		// Health:        handlers.Health(),
	}

	var port int
	port = 8080
	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		TCPPort:      &port,
	}

	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
}
