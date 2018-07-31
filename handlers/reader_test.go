package handlers

import (
	"fmt"
	"testing"
)

func Test_BuildLabelsAndAnnotationsFromServiceSpec_NoLabels(t *testing.T) {
	container := make(map[string]string)

	labels, annotation := buildLabelsAndAnnotations(container)

	if len(labels) != 0 {
		t.Errorf("want: %d entries labels got: %d", 0, len(labels))
	}

	if len(annotation) != 0 {
		t.Errorf("want: %d entries annotations got: %d", 0, len(annotation))
	}
}

func Test_BuildLabelsAndAnnotationsFromServiceSpec_Labels(t *testing.T) {
	container := map[string]string{
		"foo":  "baa",
		"fizz": "buzz",
	}

	labels, annotation := buildLabelsAndAnnotations(container)

	if len(labels) != 2 {
		t.Errorf("want: %d labels got: %d", 2, len(labels))
	}

	if len(annotation) != 0 {
		t.Errorf("want: %d annotations got: %d", 0, len(annotation))
	}

	if _, ok := labels["fizz"]; !ok {
		t.Errorf("want: '%s' entry in label map got: key not found", "fizz")
	}
}

func Test_BuildLabelsAndAnnotationsFromServiceSpec_Annotations(t *testing.T) {
	container := map[string]string{
		"foo":  "baa",
		"fizz": "buzz",
		fmt.Sprintf("%scurrent-time", annotationLabelPrefix): "Wed 25 Jul 07:10:34 BST 2018",
	}

	labels, annotation := buildLabelsAndAnnotations(container)

	if len(labels) != 2 {
		t.Errorf("want: %d labels got: %d", 2, len(labels))
	}

	if len(annotation) != 1 {
		t.Errorf("want: %d annotation got: %d", 1, len(annotation))
	}

	if _, ok := annotation[fmt.Sprintf("%scurrent-time", annotationLabelPrefix)]; !ok {
		t.Errorf("want: '%s' entry in annotation map got: key not found", "current-time")
	}
}
