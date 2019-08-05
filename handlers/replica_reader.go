package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"

	typesv1 "github.com/openfaas/faas-provider/types"
)

// ReplicaReader reads replica and image status data from a function
func ReplicaReader(c *client.Client) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		functionName := vars["name"]

		log.Printf("ReplicaReader - reading function: %s\n", functionName)

		functions, err := readServices(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		var found *typesv1.FunctionStatus
		for _, function := range functions {
			if function.Name == functionName {
				found = &function
				break
			}
		}

		if found == nil {
			w.WriteHeader(404)
			return
		}

		replicas, replicaErr := getAvailableReplicas(c, found.Name)
		if replicaErr != nil {
			log.Printf("%s\n", replicaErr.Error())

			// Fail-over as 0
		}

		found.AvailableReplicas = replicas

		functionBytes, _ := json.Marshal(found)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)
	}
}

func getAvailableReplicas(c *client.Client, service string) (uint64, error) {

	taskFilter := filters.NewArgs()
	taskFilter.Add("_up-to-date", "true")
	taskFilter.Add("service", service)
	taskFilter.Add("desired-state", "running")

	tasks, err := c.TaskList(context.Background(), types.TaskListOptions{Filters: taskFilter})
	if err != nil {
		return 0, fmt.Errorf("getAvailableReplicas for: %s failed %s", service, err.Error())
	}

	replicas := uint64(0)
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning {
			replicas++
		}
	}

	return replicas, nil
}
