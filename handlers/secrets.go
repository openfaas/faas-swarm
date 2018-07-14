package handlers

import (
	"context"
	"fmt"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

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
