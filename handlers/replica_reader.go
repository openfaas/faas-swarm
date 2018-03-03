package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas/gateway/requests"
)

// ReplicaReader reads replica and image status data from a function
func ReplicaReader(c *client.Client) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Update replicas")

		functions, err := readServices(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		vars := mux.Vars(r)
		functionName := vars["name"]

		var found *requests.Function
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

		found.AvailableReplicas = getAvailableReplicas(c, found.Name)

		functionBytes, _ := json.Marshal(found)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)

	}
}

func getAvailableReplicas(c *client.Client, service string) uint64 {

	taskFilter := filters.NewArgs()
	taskFilter.Add("_up-to-date", "true")
	taskFilter.Add("service", service)
	taskFilter.Add("desired-state", "running")
	tasks, err := c.TaskList(context.Background(), types.TaskListOptions{Filters: taskFilter})
	if err != nil {
		log.Printf("getAvailableReplicas for %s failed %v", service, err)
		return 0
	}
	replicas := uint64(0)
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning {
			replicas++
		}
	}

	return replicas
}
