package cmd

import (
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// kubectlCompletionFunc provides shell completions for kubectl flags, resource types, and resource names.
func kubectlCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Handle flag value completions.
	if len(args) > 0 {
		prev := args[len(args)-1]
		if prev == "--context" {
			return getKubeContexts(), cobra.ShellCompDirectiveNoFileComp
		}
		if prev == "-n" || prev == "--namespace" {
			ctx := findContextInArgs(args)
			return getKubeNamespaces(ctx), cobra.ShellCompDirectiveNoFileComp
		}
	}

	// Handle flag name completions.
	if strings.HasPrefix(toComplete, "--") || strings.HasPrefix(toComplete, "-") {
		flags := []string{"--context", "--namespace", "-n", "-l", "-A", "--all-namespaces", "--selector"}
		return flags, cobra.ShellCompDirectiveNoFileComp
	}

	// Extract positional args (skip flags and their values).
	positional := extractPositionalArgs(args)
	ctx := findContextInArgs(args)
	ns := findNamespaceInArgs(args)

	// No positional args yet: suggest "get" as the verb.
	if len(positional) == 0 {
		return []string{"get"}, cobra.ShellCompDirectiveNoFileComp
	}

	// After "get": suggest resource types.
	if len(positional) == 1 && strings.EqualFold(positional[0], "get") {
		return getKubeResourceTypes(ctx), cobra.ShellCompDirectiveNoFileComp
	}

	// After "get <resource>": suggest resource names.
	if len(positional) == 2 && strings.EqualFold(positional[0], "get") {
		return getKubeResourceNames(positional[1], ctx, ns), cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

// getKubeContexts returns context names from the kubeconfig.
func getKubeContexts() []string {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig, err := rules.Load()
	if err != nil {
		return nil
	}
	contexts := make([]string, 0, len(kubeConfig.Contexts))
	for name := range kubeConfig.Contexts {
		contexts = append(contexts, name)
	}
	return contexts
}

// getKubeNamespaces returns namespace names by shelling out to kubectl.
func getKubeNamespaces(context string) []string {
	args := []string{"get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}"}
	if context != "" {
		args = append(args, "--context", context)
	}
	out, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		return nil
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil
	}
	return strings.Split(raw, " ")
}

// findContextInArgs scans args for --context <value> and returns the value.
func findContextInArgs(args []string) string {
	for i, arg := range args {
		if arg == "--context" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// findNamespaceInArgs scans args for -n/--namespace <value> and returns the value.
func findNamespaceInArgs(args []string) string {
	for i, arg := range args {
		if (arg == "-n" || arg == "--namespace") && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// flagsWithValues lists flags that consume the next argument as a value.
var flagsWithValues = map[string]bool{
	"--context":   true,
	"-n":          true,
	"--namespace": true,
	"-l":          true,
	"--selector":  true,
}

// extractPositionalArgs filters out flags and their values, returning only positional arguments.
func extractPositionalArgs(args []string) []string {
	var positional []string
	skip := false
	for _, arg := range args {
		if skip {
			skip = false
			continue
		}
		if flagsWithValues[arg] {
			skip = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		positional = append(positional, arg)
	}
	return positional
}

// getKubeResourceTypes returns available resource types from the cluster.
func getKubeResourceTypes(context string) []string {
	args := []string{"api-resources", "--no-headers", "--verbs=get", "-o", "name"}
	if context != "" {
		args = append(args, "--context", context)
	}
	out, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		return nil
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}

// getKubeResourceNames returns resource names for a given type.
func getKubeResourceNames(resourceType, context, namespace string) []string {
	args := []string{"get", resourceType, "-o", "jsonpath={.items[*].metadata.name}"}
	if context != "" {
		args = append(args, "--context", context)
	}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	out, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		return nil
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil
	}
	return strings.Split(raw, " ")
}
