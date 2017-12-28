package handlers

import (
	"encoding/json"
	"log"
	"net/http"

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

		functionBytes, _ := json.Marshal(found)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)

	}
}
