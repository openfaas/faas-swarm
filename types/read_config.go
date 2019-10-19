// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package types

import (
	ftypes "github.com/openfaas/faas-provider/types"
)

// ReadConfig constitutes config from env variables
type ReadConfig struct {
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv ftypes.HasEnv) (SwarmConfig, error) {
	cfg := SwarmConfig{}

	faasCfg, err := ftypes.ReadConfig{}.Read(hasEnv)
	if err != nil {
		return cfg, err
	}

	cfg.DNSRoundRobin = ftypes.ParseBoolValue(hasEnv.Getenv("dnsrr"), false)
	cfg.FaaSConfig = *faasCfg

	return cfg, nil
}

// SwarmConfig contains the configuration for the process.
type SwarmConfig struct {
	// DNSRoundRobin controls how faas-swarm will lookup functions when proxying requests.
	// When
	//	DNSRoundRobin = true
	// faas-swarm will look up the function directly from Swarm's DNS via the tasks.functionName
	// when
	// 	DNSRoundRObin = false
	// faas-swarm will attempt to resolve the function by name, validating using the Swarm API
	DNSRoundRobin bool
	// FaasConfig contains the standard OpenFaaS provider configuration
	FaaSConfig ftypes.FaaSConfig
}
