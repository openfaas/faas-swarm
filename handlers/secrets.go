package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/filters"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"

	"github.com/openfaas/faas/gateway/requests"
)

var (
	ownerLabel      = "com.openfaas.owner"
	ownerLabelValue = "openfaas"
)

func MakeSecretsHandler(c client.SecretAPIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		body, readBodyErr := ioutil.ReadAll(r.Body)
		if readBodyErr != nil {
			log.Printf("couldn't read body of a request: %s", readBodyErr)

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		var (
			responseStatus int
			responseBody   []byte
			responseErr    error
		)

		switch r.Method {
		case http.MethodGet:
			responseStatus, responseBody, responseErr = getSecrets(c, body)
			break
		case http.MethodPost:
			responseStatus, responseBody, responseErr = createNewSecret(c, body)
			break
		case http.MethodPut:
			responseStatus, responseBody, responseErr = updateSecret(c, body)
			break
		case http.MethodDelete:
			responseStatus, responseBody, responseErr = deleteSecret(c, body)
			break
		}

		if responseErr != nil {
			log.Println(responseErr)

			w.WriteHeader(responseStatus)

			return
		}

		if responseBody != nil {
			_, writeErr := w.Write(responseBody)

			if writeErr != nil {
				log.Println("cannot write body of a response")

				w.WriteHeader(http.StatusInternalServerError)

				return
			}
		}

		w.WriteHeader(responseStatus)
	}
}

func getSecretsWithLabel(c client.SecretAPIClient, labelName string, labelValue string) ([]swarm.Secret, error) {
	secrets, secretListErr := c.SecretList(context.Background(), types.SecretListOptions{})
	if secretListErr != nil {
		return nil, secretListErr
	}

	var filteredSecrets []swarm.Secret

	for _, secret := range secrets {
		if secret.Spec.Labels[labelName] == labelValue {
			filteredSecrets = append(filteredSecrets, secret)
		}
	}

	return filteredSecrets, nil
}

func getSecretWithName(c client.SecretAPIClient, name string) (secret *swarm.Secret, err error, status int) {
	secrets, secretListErr := c.SecretList(context.Background(), types.SecretListOptions{})
	if secretListErr != nil {
		return nil, secretListErr, http.StatusInternalServerError
	}

	for _, secret := range secrets {
		if secret.Spec.Name == name {
			if secret.Spec.Labels[ownerLabel] == ownerLabelValue {
				return &secret, nil, http.StatusOK
			}

			return nil, fmt.Errorf(
				"found secret with name: %s, but it doesn't have label: %s == %s",
				name,
				ownerLabel,
				ownerLabelValue,
			), http.StatusInternalServerError
		}
	}

	return nil, fmt.Errorf("not found secret with name: %s", name), http.StatusNotFound
}

func getSecrets(c client.SecretAPIClient, _ []byte) (responseStatus int, responseBody []byte, err error) {
	secrets, err := getSecretsWithLabel(c, ownerLabel, ownerLabelValue)
	if err != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf(
			"cannot get secrets with label: %s == %s in secretGetHandler: %s",
			ownerLabel,
			ownerLabelValue,
			err,
		)
	}

	results := []requests.Secret{}

	for _, s := range secrets {
		results = append(results, requests.Secret{Name: s.Spec.Name, Value: string(s.Spec.Data)})
	}

	resultsJson, marshalErr := json.Marshal(results)
	if marshalErr != nil {
		return http.StatusInternalServerError,
			nil,
			fmt.Errorf("error marshalling secrets to json: %s", marshalErr)

	}

	return http.StatusOK, resultsJson, nil
}

func createNewSecret(c client.SecretAPIClient, body []byte) (responseStatus int, responseBody []byte, err error) {
	var secret requests.Secret

	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return http.StatusBadRequest, nil, fmt.Errorf(
			"error unmarshalling body to json in secretPostHandler: %s",
			unmarshalErr,
		)
	}

	_, createSecretErr := c.SecretCreate(context.Background(), swarm.SecretSpec{
		Annotations: swarm.Annotations{
			Name: secret.Name,
			Labels: map[string]string{
				ownerLabel: ownerLabelValue,
			},
		},
		Data: []byte(secret.Value),
	})
	if createSecretErr != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf(
			"error creating secret in secretPostHandler: %s",
			createSecretErr,
		)
	}

	return http.StatusCreated, nil, nil
}

func updateSecret(c client.SecretAPIClient, body []byte) (responseStatus int, responseBody []byte, err error) {
	var secret requests.Secret

	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return http.StatusBadRequest, nil, fmt.Errorf(
			"error unmarshaling secret in secretPutHandler: %s",
			unmarshalErr,
		)
	}

	foundSecret, getSecretErr, status := getSecretWithName(c, secret.Name)
	if getSecretErr != nil {
		return status, nil, fmt.Errorf(
			"cannot get secret with name: %s. Error: %s",
			secret.Name,
			getSecretErr.Error(),
		)
	}

	updateSecretErr := c.SecretUpdate(context.Background(), foundSecret.ID, foundSecret.Version, swarm.SecretSpec{
		Annotations: swarm.Annotations{
			Name: secret.Name,
			Labels: map[string]string{
				ownerLabel: ownerLabelValue,
			},
		},
		Data: []byte(secret.Value),
	})

	if updateSecretErr != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf(
			"couldn't update secret (name: %s, ID: %s) because of error: %s",
			secret.Name,
			foundSecret.ID,
			updateSecretErr.Error(),
		)
	}

	return http.StatusOK, nil,nil
}

func deleteSecret(c client.SecretAPIClient, body []byte) (responseStatus int, responseBody []byte, err error) {
	var secret requests.Secret

	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return http.StatusBadRequest, nil, fmt.Errorf(
			"error unmarshaling secret in secretDeleteHandler: %s",
			unmarshalErr,
		)
	}

	foundSecret, getSecretErr, status := getSecretWithName(c, secret.Name)
	if getSecretErr != nil {
		return status, nil, fmt.Errorf(
			"cannot get secret with name: %s, which you want to remove. Error: %s",
			secret.Name,
			getSecretErr,
		)
	}

	removeSecretErr := c.SecretRemove(context.Background(), foundSecret.ID)
	if removeSecretErr != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf(
			"error trying to remove secret (name: `%s`, ID: `%s`): %s",
			secret.Name,
			foundSecret.ID,
			removeSecretErr,
		)
	}

	return http.StatusOK, nil, nil
}


func makeSecretsArray(c *client.Client, secretNames []string) ([]*swarm.SecretReference, error) {
	values := []*swarm.SecretReference{}

	if len(secretNames) == 0 {
		return values, nil
	}

	secretOpts := new(opts.SecretOpt)
	for _, secret := range secretNames {
		secretSpec := fmt.Sprintf("source=%s,target=/var/openfaas/secrets/%s", secret, secret)
		if err := secretOpts.Set(secretSpec); err != nil {
			return nil, err
		}
	}

	requestedSecrets := make(map[string]bool)
	ctx := context.Background()

	// query the Swarm for the requested secret ids, these are required to complete
	// the spec
	args := filters.NewArgs()
	for _, opt := range secretOpts.Value() {
		args.Add("name", opt.SecretName)
	}

	secrets, err := c.SecretList(ctx, types.SecretListOptions{
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	// create map of matching secrets for easy lookup
	foundSecrets := make(map[string]string)
	foundSecretNames := []string{}
	for _, secret := range secrets {
		foundSecrets[secret.Spec.Annotations.Name] = secret.ID
		foundSecretNames = append(foundSecretNames, secret.Spec.Annotations.Name)
	}

	// mimics the simple syntax for `docker service create --secret foo`
	// and the code is based on the docker cli
	for _, opts := range secretOpts.Value() {

		secretName := opts.SecretName
		if _, exists := requestedSecrets[secretName]; exists {
			return nil, fmt.Errorf("duplicate secret target for %s not allowed", secretName)
		}

		id, ok := foundSecrets[secretName]
		if !ok {
			return nil, fmt.Errorf("secret not found: %s; possible choices:\n%v", secretName, foundSecretNames)
		}

		options := new(swarm.SecretReference)
		*options = *opts
		options.SecretID = id

		requestedSecrets[secretName] = true
		values = append(values, options)
	}

	return values, nil
}
