package handlers

import (
	"net/http"
)

func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
