// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"testing"

	"github.com/openfaas/faas/gateway/requests"
)

// Test_ParseMemory exploratory testing to document how to convert
// from Docker limits notation to bytes value.
func Test_ParseMemoryInMegabytes(t *testing.T) {
	value := "512 m"

	val, err := parseMemory(value)
	if err != nil {
		t.Error(err)
	}

	if val != megaBytes(512) {
		t.Errorf("want: %d got: %d", 1024, val)
	}
}

func megaBytes(mbs int64) int64 {
	return 1024 * 1024 * mbs
}

func Test_ParseInvalidMemoryInMegabytes(t *testing.T) {
	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{
			Memory: "wrong",
		},
		Limits: &requests.FunctionResources{},
	}

	res := buildResources(&req)

	if res.Reservations != nil {
		t.Fatalf("Expected nil reservation due to incorrect value provided")
	}
}

func TestBuildSwarmResourcesAddsCPULimits(t *testing.T) {
	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{},
		Limits: &requests.FunctionResources{
			CPU: "100",
		},
	}

	res := buildResources(&req)

	if res.Limits.NanoCPUs != 100 {
		t.Fatalf("Expected CPU limit of 100, got %d", res.Limits.NanoCPUs)
	}
}

func TestBuildSwarmResourcesAddsCPUReservations(t *testing.T) {
	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{
			CPU: "100",
		},
		Limits: &requests.FunctionResources{},
	}

	res := buildResources(&req)

	if res.Reservations.NanoCPUs != 100 {
		t.Fatalf("Expected CPU limit of 100, got %d", res.Reservations.NanoCPUs)
	}
}

func TestBuildSwarmResourcesWithInvalidCPUSetsReservationsTo0(t *testing.T) {
	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{
			Memory: "invalid",
		},
		Limits: &requests.FunctionResources{},
	}

	res := buildResources(&req)

	if res.Reservations != nil {
		t.Fatalf("Expected cpu reservation to be nil due to invalid inputs")
	}
}
