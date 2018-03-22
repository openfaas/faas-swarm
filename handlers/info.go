package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/openfaas/faas-provider/types"
)

const (
	//SwarmIdentifier identifier string for swarm provider
	SwarmIdentifier = "swarm"
	//SwarmProvider provider string for swarm provider
	SwarmProvider = "faas-swarm"
)

//MakeInfoHandler creates handler for /system/info endpoint
func MakeInfoHandler(version, sha string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		infoRequest := types.InfoRequest{
			Orchestration: SwarmIdentifier,
			Provider:      SwarmProvider,
			Version: types.ProviderVersion{
				Release: version,
				SHA:     sha,
			},
		}

		jsonOut, marshalErr := json.Marshal(infoRequest)
		if marshalErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonOut)
	}
}
