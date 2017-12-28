package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/openfaas/faas/gateway/requests"
)

// FunctionReader reads functions from Swarm metadata
func FunctionReader(wildcard bool, c *client.Client) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		functions, err := readServices(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)

	}
}

func readServices(c *client.Client) ([]requests.Function, error) {

	serviceFilter := filters.NewArgs()

	options := types.ServiceListOptions{
		Filters: serviceFilter,
	}

	services, err := c.ServiceList(context.Background(), options)
	if err != nil {
		log.Println("Error getting service list:", err)
		return nil, fmt.Errorf("error getting service list")
	}

	// TODO: Filter only "faas" functions (via metadata?)
	functions := []requests.Function{}

	for _, service := range services {

		if len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0 {
			envProcess := getEnvProcess(service.Spec.TaskTemplate.ContainerSpec.Env)

			// Required (copy by value)
			labels := service.Spec.Annotations.Labels

			f := requests.Function{
				Name:            service.Spec.Name,
				Image:           service.Spec.TaskTemplate.ContainerSpec.Image,
				InvocationCount: 0,
				Replicas:        *service.Spec.Mode.Replicated.Replicas,
				EnvProcess:      envProcess,
				Labels:          &labels,
			}

			functions = append(functions, f)
		}
	}

	return functions, err
}

func getEnvProcess(envVars []string) string {
	var value string
	for _, env := range envVars {
		if strings.Contains(env, "fprocess=") {
			value = env[len("fprocess="):]
		}
	}

	return value
}
