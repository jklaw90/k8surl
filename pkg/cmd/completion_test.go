package cmd

import (
	"reflect"
	"slices"
	"testing"
)

func TestExtractPositionalArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "no args",
			args: nil,
			want: nil,
		},
		{
			name: "only positional",
			args: []string{"get", "pods", "my-pod"},
			want: []string{"get", "pods", "my-pod"},
		},
		{
			name: "flags with values are skipped",
			args: []string{"--context", "prod", "get", "pods"},
			want: []string{"get", "pods"},
		},
		{
			name: "namespace flag short",
			args: []string{"-n", "kube-system", "get", "pods"},
			want: []string{"get", "pods"},
		},
		{
			name: "namespace flag long",
			args: []string{"--namespace", "kube-system", "get", "pods"},
			want: []string{"get", "pods"},
		},
		{
			name: "flags mixed throughout",
			args: []string{"get", "--context", "prod", "pods", "-n", "default"},
			want: []string{"get", "pods"},
		},
		{
			name: "boolean flags skipped",
			args: []string{"-A", "get", "pods"},
			want: []string{"get", "pods"},
		},
		{
			name: "selector flag",
			args: []string{"get", "pods", "-l", "app=nginx"},
			want: []string{"get", "pods"},
		},
		{
			name: "all flags no positional",
			args: []string{"--context", "prod", "-n", "default"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPositionalArgs(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractPositionalArgs(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestFindContextInArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no context",
			args: []string{"get", "pods"},
			want: "",
		},
		{
			name: "context present",
			args: []string{"--context", "prod", "get", "pods"},
			want: "prod",
		},
		{
			name: "context at end",
			args: []string{"get", "pods", "--context", "staging"},
			want: "staging",
		},
		{
			name: "context flag without value",
			args: []string{"get", "--context"},
			want: "",
		},
		{
			name: "empty args",
			args: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findContextInArgs(tt.args)
			if got != tt.want {
				t.Errorf("findContextInArgs(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestFindNamespaceInArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no namespace",
			args: []string{"get", "pods"},
			want: "",
		},
		{
			name: "short flag",
			args: []string{"-n", "kube-system", "get", "pods"},
			want: "kube-system",
		},
		{
			name: "long flag",
			args: []string{"get", "pods", "--namespace", "monitoring"},
			want: "monitoring",
		},
		{
			name: "flag without value",
			args: []string{"get", "-n"},
			want: "",
		},
		{
			name: "empty args",
			args: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findNamespaceInArgs(tt.args)
			if got != tt.want {
				t.Errorf("findNamespaceInArgs(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestKubectlCompletionFunc_Verbs(t *testing.T) {
	// With no positional args, should suggest "get".
	completions, _ := kubectlCompletionFunc(nil, nil, "")
	if !reflect.DeepEqual(completions, []string{"get"}) {
		t.Errorf("expected [get] with no args, got %v", completions)
	}

	// With only flags, should still suggest "get" since there are no positional args.
	completions, _ = kubectlCompletionFunc(nil, []string{"--context", "prod"}, "")
	if !reflect.DeepEqual(completions, []string{"get"}) {
		t.Errorf("expected [get] with only flags, got %v", completions)
	}
}

func TestKubectlCompletionFunc_FlagNames(t *testing.T) {
	// Typing "--" should suggest flag names.
	completions, _ := kubectlCompletionFunc(nil, nil, "--")
	if len(completions) == 0 {
		t.Fatal("expected flag completions, got none")
	}
	if !slices.Contains(completions, "--context") {
		t.Errorf("expected --context in flag completions, got %v", completions)
	}

	// Typing "-" should also suggest flags.
	completions, _ = kubectlCompletionFunc(nil, nil, "-")
	if len(completions) == 0 {
		t.Fatal("expected flag completions for '-', got none")
	}
}

func TestKubectlCompletionFunc_FlagValues(t *testing.T) {
	// After --context, should return context list (may be empty in test env, but shouldn't panic).
	completions, _ := kubectlCompletionFunc(nil, []string{"--context"}, "")
	_ = completions // just verify no panic

	// After -n, should return namespace list (may be empty in test env).
	completions, _ = kubectlCompletionFunc(nil, []string{"-n"}, "")
	_ = completions

	// After --namespace, same.
	completions, _ = kubectlCompletionFunc(nil, []string{"--namespace"}, "")
	_ = completions
}

func TestKubectlCompletionFunc_AfterResourceType(t *testing.T) {
	// After "get pods", should attempt to fetch resource names.
	// In a test environment without a cluster, this returns nil — just verify no panic.
	completions, _ := kubectlCompletionFunc(nil, []string{"get", "pods"}, "")
	_ = completions

	// With flags mixed in.
	completions, _ = kubectlCompletionFunc(nil, []string{"--context", "prod", "get", "pods"}, "")
	_ = completions
}

func TestKubectlCompletionFunc_PastResourceName(t *testing.T) {
	// After "get pods my-pod", no more completions expected.
	completions, _ := kubectlCompletionFunc(nil, []string{"get", "pods", "my-pod"}, "")
	if completions != nil {
		t.Errorf("expected nil completions after resource name, got %v", completions)
	}
}
