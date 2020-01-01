package handlers

import (
	"encoding/json"
	"net/http"

	typesv1 "github.com/openfaas/faas-provider/types"
)

const (
	//OrchestrationIdentifier identifier string for swarm provider
	OrchestrationIdentifier = "swarm"
	//ProviderName provider string for swarm provider
	ProviderName = "faas-swarm"
)

//MakeInfoHandler creates handler for /system/info endpoint
func MakeInfoHandler(version, sha string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		infoResponse := typesv1.InfoResponse{
			Orchestration: OrchestrationIdentifier,
			Provider:      ProviderName,
			Version: typesv1.ProviderVersion{
				Release: version,
				SHA:     sha,
			},
		}

		jsonOut, marshalErr := json.Marshal(infoResponse)
		if marshalErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonOut)
	}
}
