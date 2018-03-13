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
	"github.com/docker/docker/client"
	"github.com/openfaas/faas/gateway/requests"
)

// DeleteHandler delete a function
func DeleteHandler(c *client.Client) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		req := requests.DeleteFunctionRequest{}
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
			w.WriteHeader(http.StatusOK)
		}

	}
}
