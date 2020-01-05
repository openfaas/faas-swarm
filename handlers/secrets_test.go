package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/openfaas/faas/gateway/requests"
)

func genFakeSecret(name string, data string, includeOwnerLabel bool) swarm.Secret {
	secret := swarm.Secret{
		ID: name,
		Spec: swarm.SecretSpec{
			Data: []byte(data),
			Annotations: swarm.Annotations{
				Name: name,
			},
		},
	}

	if includeOwnerLabel {
		secret.Spec.Annotations.Labels = map[string]string{
			ownerLabel: ownerLabelValue,
		}
	}

	return secret
}

func getInitialSecrets(includeNonOwnedSecrets bool) map[string]swarm.Secret {
	initialSecrets := map[string]swarm.Secret{
		"foo":        genFakeSecret("foo", "baz", true),
		"foobar":     genFakeSecret("foobar", "bar", true),
		"baz_fuz":    genFakeSecret("baz_fuz", "foo", true),
		"barfoo_baz": genFakeSecret("barfoo_baz", "foobar", true),
	}

	if includeNonOwnedSecrets {
		initialSecrets["without_label"] = genFakeSecret("without_label", "test", false)
	}

	return initialSecrets
}

type fakeDockerSecretAPIClient struct {
	secrets map[string]swarm.Secret
}

func newFakeDockerSecretAPIClient() fakeDockerSecretAPIClient {
	return fakeDockerSecretAPIClient{
		secrets: getInitialSecrets(true),
	}
}

func (c *fakeDockerSecretAPIClient) Reset() {
	c.secrets = getInitialSecrets(true)
}

func (c *fakeDockerSecretAPIClient) SecretList(
	_ context.Context,
	options types.SecretListOptions,
) ([]swarm.Secret, error) {
	secrets := []swarm.Secret{}

	for _, value := range c.secrets {
		secrets = append(secrets, value)
	}

	return secrets, nil
}

func (c *fakeDockerSecretAPIClient) SecretCreate(
	_ context.Context,
	secretDesc swarm.SecretSpec,
) (types.SecretCreateResponse, error) {
	id := secretDesc.Name

	newSecret := swarm.Secret{
		ID: id,
		Spec: swarm.SecretSpec{
			Annotations: swarm.Annotations{
				Name: secretDesc.Name,
				Labels: map[string]string{
					ownerLabel: ownerLabelValue,
				},
			},
			Data: secretDesc.Data,
		},
	}

	c.secrets[secretDesc.Name] = newSecret

	return types.SecretCreateResponse{ID: id}, nil
}

func (c *fakeDockerSecretAPIClient) SecretRemove(_ context.Context, id string) error {
	if _, ok := c.secrets[id]; !ok {
		return fmt.Errorf("secret with id: %s not found", id)
	}

	delete(c.secrets, id)

	return nil
}

func (c *fakeDockerSecretAPIClient) SecretInspectWithRaw(
	_ context.Context,
	name string,
) (swarm.Secret, []byte, error) {
	return swarm.Secret{}, nil, fmt.Errorf("SecretInspectWithRaw Not Implemented")
}

func (c *fakeDockerSecretAPIClient) SecretUpdate(
	_ context.Context,
	id string,
	version swarm.Version,
	secret swarm.SecretSpec,
) error {
	return fmt.Errorf("returning error because it's not possible to update existing secrets with docker")
}

func Test_SecretsHandler(t *testing.T) {
	dockerClient := newFakeDockerSecretAPIClient()
	secretsHandler := MakeSecretsHandler(&dockerClient)

	secretName := "testsecret"

	t.Run("create managed secrets returns error", func(t *testing.T) {
		defer dockerClient.Reset()

		secretValue := "testsecretvalue"
		payload := fmt.Sprintf(`{"name": "%s", "value": "%s"}`, secretName, secretValue)
		req := httptest.NewRequest("POST", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected status code '%d', got '%d'", http.StatusCreated, resp.StatusCode)
		}

		if _, secretExist := dockerClient.secrets[secretName]; !secretExist {
			t.Errorf("secret `%s` was not created as expected", secretName)
		}

		if data := dockerClient.secrets[secretName].Spec.Data; !bytes.Equal(data, []byte(secretValue)) {
			t.Errorf("want secret: `%s` to be equal `%s`, got: `%s`", secretName, secretValue, string(data))
		}
	})

	t.Run("update managed secrets", func(t *testing.T) {
		newSecretValue := "newtestsecretvalue"
		payload := fmt.Sprintf(`{"name": "%s", "value": "%s"}`, "foo", newSecretValue)
		req := httptest.NewRequest("PUT", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status code '%d', got '%d'", http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("list managed secrets only", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		w := httptest.NewRecorder()
		want := getInitialSecrets(false)
		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code '%d', got '%d'", http.StatusOK, resp.StatusCode)
		}

		decoder := json.NewDecoder(resp.Body)

		secretList := []requests.Secret{}
		err := decoder.Decode(&secretList)
		if err != nil {
			t.Error(err)
		}

		for key, secret := range want {
			var exists bool
			var data string

			for _, secret := range secretList {
				if secret.Name == key {
					exists = true
					data = string(secret.Value)

					break
				}
			}

			if !exists {
				t.Errorf("expected secret: `%s` to be listed", key)
			}

			if string(secret.Spec.Data) != data {
				t.Errorf("expected secret: `%s` to have value: `%s`, got: `%s`", key, string(secret.Spec.Data), data)
			}
		}
	})

	t.Run("delete managed secrets", func(t *testing.T) {
		defer dockerClient.Reset()

		secretName := "foobar"
		payload := fmt.Sprintf(`{"name": "%s"}`, secretName)
		req := httptest.NewRequest("DELETE", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code '%d', got '%d'", http.StatusOK, resp.StatusCode)
		}

		if _, secretExist := dockerClient.secrets[secretName]; secretExist {
			t.Errorf("expected secret with name: `%s` to be removed", secretName)
		}
	})
}
