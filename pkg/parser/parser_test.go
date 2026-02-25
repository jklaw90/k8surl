package parser

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAllowed(t *testing.T) {
	tests := []struct {
		name   string
		target string
		kinds  []string
		want   bool
	}{
		{"exact match", "Pod", []string{"Pod"}, true},
		{"case insensitive", "pod", []string{"Pod"}, true},
		{"case insensitive reverse", "Pod", []string{"pod"}, true},
		{"wildcard", "Anything", []string{"*"}, true},
		{"not in list", "Service", []string{"Pod", "Deployment"}, false},
		{"empty list", "Pod", []string{}, false},
		{"wildcard with others", "Service", []string{"Pod", "*"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Allowed(tt.target, tt.kinds)
			if got != tt.want {
				t.Errorf("Allowed(%q, %v) = %v, want %v", tt.target, tt.kinds, got, tt.want)
			}
		})
	}
}

func TestAllowed_DoesNotMutateInput(t *testing.T) {
	kinds := []string{"Pod", "Service", "Deployment"}
	original := make([]string, len(kinds))
	copy(original, kinds)

	Allowed("pod", kinds)

	for i, k := range kinds {
		if k != original[i] {
			t.Errorf("Allowed() mutated input: kinds[%d] = %q, want %q", i, k, original[i])
		}
	}
}

func TestDecode(t *testing.T) {
	t.Run("valid pod JSON", func(t *testing.T) {
		json := `{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod","namespace":"default"}}`
		obj, kind, err := Decode(strings.NewReader(json))
		if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if kind != "Pod" {
			t.Errorf("Decode() kind = %q, want %q", kind, "Pod")
		}
		if obj == nil {
			t.Error("Decode() returned nil object")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, _, err := Decode(strings.NewReader("{not json"))
		if err == nil {
			t.Error("Decode() expected error for invalid JSON, got nil")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		_, _, err := Decode(strings.NewReader(""))
		if err == nil {
			t.Error("Decode() expected error for empty input, got nil")
		}
	})
}

func TestRenderTemplates(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "default",
		},
	}

	t.Run("simple template", func(t *testing.T) {
		result, err := RenderTemplates(pod, []string{"{.metadata.name}"})
		if err != nil {
			t.Fatalf("RenderTemplates() error = %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("RenderTemplates() returned %d results, want 1", len(result))
		}
		if result[0] != "my-pod" {
			t.Errorf("RenderTemplates() = %q, want %q", result[0], "my-pod")
		}
	})

	t.Run("multiple templates", func(t *testing.T) {
		result, err := RenderTemplates(pod, []string{"{.metadata.name}", "{.metadata.namespace}"})
		if err != nil {
			t.Fatalf("RenderTemplates() error = %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("RenderTemplates() returned %d results, want 2", len(result))
		}
		if result[0] != "my-pod" {
			t.Errorf("result[0] = %q, want %q", result[0], "my-pod")
		}
		if result[1] != "default" {
			t.Errorf("result[1] = %q, want %q", result[1], "default")
		}
	})

	t.Run("invalid template syntax", func(t *testing.T) {
		_, err := RenderTemplates(pod, []string{"{.metadata.name"})
		if err == nil {
			t.Error("RenderTemplates() expected error for invalid template, got nil")
		}
	})

	t.Run("empty templates", func(t *testing.T) {
		result, err := RenderTemplates(pod, []string{})
		if err != nil {
			t.Fatalf("RenderTemplates() error = %v", err)
		}
		if result != nil {
			t.Errorf("RenderTemplates() = %v, want nil", result)
		}
	})
}
