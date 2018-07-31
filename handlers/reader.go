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
func FunctionReader(wildcard bool, c client.ServiceAPIClient) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		functions, err := readServices(c)
		if err != nil {
			log.Printf("Error getting service list: %s\n", err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)
	}
}

func readServices(c client.ServiceAPIClient) ([]requests.Function, error) {
	functions := []requests.Function{}
	serviceFilter := filters.NewArgs()

	options := types.ServiceListOptions{
		Filters: serviceFilter,
	}

	services, err := c.ServiceList(context.Background(), options)
	if err != nil {
		return functions, fmt.Errorf("error getting service list: %s", err.Error())
	}

	for _, service := range services {

		if len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0 {
			envProcess := getEnvProcess(service.Spec.TaskTemplate.ContainerSpec.Env)

			// Required (copy by value)
			labels, annotations := buildLabelsAndAnnotations(service.Spec.Labels)

			f := requests.Function{
				Name:            service.Spec.Name,
				Image:           service.Spec.TaskTemplate.ContainerSpec.Image,
				InvocationCount: 0,
				Replicas:        *service.Spec.Mode.Replicated.Replicas,
				EnvProcess:      envProcess,
				Labels:          &labels,
				Annotations:     &annotations,
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

func buildLabelsAndAnnotations(dockerLabels map[string]string) (labels map[string]string, annotations map[string]string) {
	for k, v := range dockerLabels {
		if strings.HasPrefix(k, annotationLabelPrefix) {
			if annotations == nil {
				annotations = make(map[string]string)
			}

			annotations[k] = v
		} else {
			if labels == nil {
				labels = make(map[string]string)
			}

			labels[k] = v
		}
	}

	return labels, annotations
}
