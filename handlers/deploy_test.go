package handlers

import (
	"fmt"

	typesv1 "github.com/openfaas/faas-provider/types"

	"testing"
)

func Test_BuildLabels_Defaults(t *testing.T) {
	request := &typesv1.FunctionDeployment{}
	val, err := buildLabels(request)

	if err != nil {
		t.Fatalf("want: no error got: %v", err)
	}

	if len(val) != 2 {
		t.Errorf("want: %d entries in label map got: %d", 2, len(val))
	}

	if _, ok := val["com.openfaas.function"]; !ok {
		t.Errorf("want: '%s' entry in label map got: key not found", "com.openfaas.function")
	}

	if _, ok := val["function"]; !ok {
		t.Errorf("want: '%s' entry in label map got: key not found", "function")
	}
}

func Test_BuildLabels_WithAnnotations(t *testing.T) {
	request := &typesv1.FunctionDeployment{
		Labels:      &map[string]string{"function_name": "echo"},
		Annotations: &map[string]string{"current-time": "Wed 25 Jul 06:41:43 BST 2018"},
	}

	val, err := buildLabels(request)

	if err != nil {
		t.Fatalf("want: no error got: %v", err)
	}

	if len(val) != 4 {
		t.Errorf("want: %d entries in combined label annotation map got: %d", 4, len(val))
	}

	if _, ok := val[fmt.Sprintf("%scurrent-time", annotationLabelPrefix)]; !ok {
		t.Errorf("want: '%s' entry in combined label annotation map got: key not found", "annotation: current-time")
	}
}

func Test_BuildLabels_NoAnnotations(t *testing.T) {
	request := &typesv1.FunctionDeployment{
		Labels: &map[string]string{"function_name": "echo"},
	}

	val, err := buildLabels(request)

	if err != nil {
		t.Fatalf("want: no error got: %v", err)
	}

	if len(val) != 3 {
		t.Errorf("want: %d entries in combined label annotation map got: %d", 3, len(val))
	}

	if _, ok := val["function_name"]; !ok {
		t.Errorf("want: '%s' entry in combined label annotation map got: key not found", "function_name")
	}
}

func Test_BuildLabels_KeyClash(t *testing.T) {
	request := &typesv1.FunctionDeployment{
		Labels: &map[string]string{
			"function_name": "echo",
			fmt.Sprintf("%scurrent-time", annotationLabelPrefix): "foo",
		},
		Annotations: &map[string]string{"current-time": "Wed 25 Jul 06:41:43 BST 2018"},
	}

	_, err := buildLabels(request)

	if err == nil {
		t.Fatal("want: an error got: nil")
	}
}
