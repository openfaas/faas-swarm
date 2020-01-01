package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

type fakeServiceDeleter struct {
	services  []swarm.Service
	listErr   error
	removeErr error
}

func (s fakeServiceDeleter) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	return s.services, s.listErr
}

func (s fakeServiceDeleter) ServiceRemove(ctx context.Context, serviceID string) error {
	return s.removeErr
}

func Test_DeleteHandler(t *testing.T) {

	cases := []struct {
		name         string
		funcName     string
		services     []swarm.Service
		listErr      error
		removeErr    error
		expectedCode int
	}{
		{
			name:         "parsing error returns StatusBadRequest",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "listing error returns StatusNotFound",
			funcName:     "test-func",
			listErr:      errors.New("failed to list functions"),
			expectedCode: http.StatusNotFound,
		},
		{
			name:     "remove error returns StatusInternalServerError",
			funcName: "test-func",
			services: []swarm.Service{
				{
					ID: "test-func-id",
					Spec: swarm.ServiceSpec{
						Annotations: swarm.Annotations{Name: "test-func"},
						TaskTemplate: swarm.TaskSpec{
							ContainerSpec: &swarm.ContainerSpec{
								Labels: map[string]string{"function": "true"},
							},
						},
					},
				},
			},
			removeErr:    errors.New("failed to delete function"),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:     "returns Accepted when no errors",
			funcName: "test-func",
			services: []swarm.Service{
				{
					ID: "test-func-id",
					Spec: swarm.ServiceSpec{
						Annotations: swarm.Annotations{Name: "test-func"},
						TaskTemplate: swarm.TaskSpec{
							ContainerSpec: &swarm.ContainerSpec{
								Labels: map[string]string{"function": "true"},
							},
						},
					},
				},
			},
			expectedCode: http.StatusAccepted,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := fakeServiceDeleter{
				listErr:   tc.listErr,
				removeErr: tc.removeErr,
				services:  tc.services,
			}
			handler := DeleteHandler(client)

			payload := fmt.Sprintf(`{"functionName": %q}`, tc.funcName)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/", strings.NewReader(payload))
			handler(w, r)

			if w.Code != tc.expectedCode {
				t.Fatalf("expected status code %d, got %d", tc.expectedCode, w.Code)
			}

		})
	}

}
