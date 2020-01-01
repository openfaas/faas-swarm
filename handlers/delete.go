package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
)

type deleteFunctionRequest struct {
	FunctionName string `json:"functionName"`
}

// ServiceDeleter is the sub-interface of client.ServiceAPIClient that is required for deleting
// a OpenFaaS Function. This interface is satisfied by *client.Client
type ServiceDeleter interface {
	ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error)
	ServiceRemove(ctx context.Context, serviceID string) error
}

// DeleteHandler delete a function
func DeleteHandler(c ServiceDeleter) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		req := deleteFunctionRequest{}
		defer r.Body.Close()
		reqData, _ := ioutil.ReadAll(r.Body)
		unmarshalErr := json.Unmarshal(reqData, &req)

		if (len(req.FunctionName) == 0) || unmarshalErr != nil {
			log.Printf("Error parsing request to remove service: %s\n", unmarshalErr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Attempting to remove service %s\n", req.FunctionName)

		serviceFilter := filters.NewArgs()
		options := types.ServiceListOptions{
			Filters: serviceFilter,
		}

		services, err := c.ServiceList(context.Background(), options)
		if err != nil {
			log.Printf("Error listing services: %s\n", err)
		}

		// TODO: Filter only "faas" functions (via metadata?)
		var serviceIDs []string
		for _, service := range services {
			isFunction := len(service.Spec.TaskTemplate.ContainerSpec.Labels["function"]) > 0

			if isFunction && req.FunctionName == service.Spec.Name {
				serviceIDs = append(serviceIDs, service.ID)
			}
		}

		if len(serviceIDs) == 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("No such service found: %s.", req.FunctionName)))
			return
		}

		var serviceRemoveErrors []error
		for _, serviceID := range serviceIDs {
			err := c.ServiceRemove(context.Background(), serviceID)
			if err != nil {
				serviceRemoveErrors = append(serviceRemoveErrors, err)
			}
		}

		if len(serviceRemoveErrors) > 0 {
			log.Printf("Error(s) removing service: %s\n", req.FunctionName)
			log.Println(serviceRemoveErrors)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}

	}
}
