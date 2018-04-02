// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
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

func Test_ParseMemoryLimits(t *testing.T) {
	want := int64(256)
	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{},
		Limits: &requests.FunctionResources{
			Memory: fmt.Sprintf("%d m", want),
		},
	}

	res := buildResources(&req)

	if res.Limits.MemoryBytes != megaBytes(want) {
		t.Fatalf("Limits.MemoryBytes want: %d, got: %d", megaBytes(want), res.Limits.MemoryBytes)
	}
}

func Test_ParseMemoryRequests(t *testing.T) {
	want := int64(256)
	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{
			Memory: fmt.Sprintf("%d m", want),
		},
		Limits: &requests.FunctionResources{},
	}

	res := buildResources(&req)

	if res.Reservations.MemoryBytes != megaBytes(want) {
		t.Fatalf("Reservations.MemoryBytes want: %d, got: %d", megaBytes(want), res.Reservations.MemoryBytes)
	}
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

func TestInvalidMemoryRequests_Ignored(t *testing.T) {
	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{
			Memory: "invalid",
		},
		Limits: &requests.FunctionResources{},
	}

	res := buildResources(&req)

	if res.Reservations != nil {
		t.Fatalf("Expected reservations to be nil due to invalid input")
	}
}

func TestInvalidMemoryLimits_Ignored(t *testing.T) {
	req := requests.CreateFunctionRequest{
		Limits: &requests.FunctionResources{
			Memory: "invalid",
		},
		Requests: &requests.FunctionResources{},
	}

	res := buildResources(&req)

	if res.Limits != nil {
		t.Fatalf("Expected limits to be nil due to invalid input")
	}
}

func TestBuildSwarmResourcesAddsCPULimits(t *testing.T) {
	want := int64(1000000)

	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{},
		Limits: &requests.FunctionResources{
			CPU: fmt.Sprintf("%d", want),
		},
	}

	res := buildResources(&req)

	if res.Limits.NanoCPUs != want {
		t.Fatalf("Expected CPU limit of %d, got %d", want, res.Limits.NanoCPUs)
	}
}

func TestBuildSwarmResourcesAddsCPUReservations(t *testing.T) {
	want := int64(1000000)
	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{
			CPU: fmt.Sprintf("%d", want),
		},
		Limits: &requests.FunctionResources{},
	}

	res := buildResources(&req)

	if res.Reservations.NanoCPUs != want {
		t.Fatalf("Expected CPU limit of %d, got %d", want, res.Reservations.NanoCPUs)
	}
}

func TestInvalidCPULimits_Ignored(t *testing.T) {
	req := requests.CreateFunctionRequest{
		Requests: &requests.FunctionResources{},
		Limits: &requests.FunctionResources{
			CPU: "invalid",
		},
	}

	res := buildResources(&req)

	if res.Limits != nil {
		t.Fatalf("Expected Limits to be nil due to invalid input")
	}
}

func TestInvalidCPURequests_Ignored(t *testing.T) {
	req := requests.CreateFunctionRequest{
		Limits: &requests.FunctionResources{},
		Requests: &requests.FunctionResources{
			CPU: "invalid",
		},
	}

	res := buildResources(&req)

	if res.Reservations != nil {
		t.Fatalf("Expected Requests to be nil due to invalid input")
	}
}
