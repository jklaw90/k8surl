package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/jklaw90/k8surl/pkg/parser"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
)

var config *Config

func NewK8surlCmd() *cobra.Command {
	k8surlCmd := &cobra.Command{
		Use:   "k8surl",
		Short: "CLI to read k8s resources and open urls based on your template config",
		Example: `kubectl get pod <pod-name> -o json | k8surl pod
kubectl k8surl pod <pod-name>`,
		SilenceUsage:       true,
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs, // allows us to not require -- for kubectl args
		ValidArgsFunction: kubectlCompletionFunc,
		Run: func(cmd *cobra.Command, args []string) {
			obj, kind, err := commandInitializer(cmd, args)
			cobra.CheckErr(err)

			kt, ok := config.KindAndTemplates[strings.ToLower(kind)]
			if !ok {
				cobra.CheckErr(fmt.Errorf("%s is not defined in the root of the configuration file", kind))
			}

			if !parser.Allowed(kind, []string{kind}) {
				cobra.CheckErr(fmt.Errorf("%s is not allowed in this command", kind))
			}

			output(obj, kt.Templates, kt.Urls)
		},
	}

	createSubCommandsFromConfig(k8surlCmd)

	return k8surlCmd
}

func createSubCommandsFromConfig(rootCmd *cobra.Command) {
	browser.Stdout = nil // reduce noise from browser package

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".k8surl")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			cobra.CheckErr(fmt.Errorf("config file not found: create ~/.k8surl.yaml to get started\n  see examples at: https://github.com/jklaw90/k8surl#examples"))
		}
		cobra.CheckErr(err)
	}
	cobra.CheckErr(viper.Unmarshal(&config))

	configCmd := &cobra.Command{
		Use:          "config",
		Short:        "view the current configuration",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			configPath := viper.GetViper().ConfigFileUsed()
			config, err := os.ReadFile(configPath)
			cobra.CheckErr(err)
			fmt.Printf("Current configuration (%s):\n\n", configPath)
			fmt.Println(string(config))
		},
	}
	rootCmd.AddCommand(configCmd)

	for k, v := range config.Commands {
		dynamicCmd := &cobra.Command{
			Use:                k,
			Short:              fmt.Sprintf("run templates for %s (kinds: %s)", k, strings.Join(v.Kinds, ", ")),
			ValidArgsFunction:  kubectlCompletionFunc,
			DisableFlagParsing: true,
			DisableSuggestions: true,
			Run: func(cmd *cobra.Command, args []string) {
				obj, kind, err := commandInitializer(cmd, args)
				cobra.CheckErr(err)

				if !parser.Allowed(kind, v.Kinds) {
					cobra.CheckErr(fmt.Errorf("%s is not allowed in this command", kind))
				}
				output(obj, v.Templates, v.Urls)
			},
		}
		if v.Short != nil {
			dynamicCmd.Short = *v.Short
		}
		if v.Example != nil {
			dynamicCmd.Example = *v.Example
		}
		rootCmd.AddCommand(dynamicCmd)
	}
}

// commandInitializer is a helper function to decode the input from the command line or stdin
func commandInitializer(cmd *cobra.Command, args []string) (runtime.Object, string, error) {
	if slices.ContainsFunc(args, func(arg string) bool {
		return arg == "--help" || arg == "-h"
	}) {
		cmd.Help()
		return nil, "", fmt.Errorf("")
	}

	var reader io.Reader
	var cmdArgs []string
	if len(args) > 0 {
		if _, err := exec.LookPath("kubectl"); err != nil {
			return nil, "", fmt.Errorf("kubectl not found in PATH: pipe input via stdin instead (e.g. cat resource.json | k8surl)")
		}

		if !slices.ContainsFunc(args, func(arg string) bool {
			return strings.EqualFold(arg, "-o") || strings.EqualFold(arg, "--output")
		}) {
			cmdArgs = append(args, []string{"-o", "json"}...)
		}

		kubectlCmd := exec.Command("kubectl", cmdArgs...)
		rawOutput, err := kubectlCmd.Output()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
				return nil, "", fmt.Errorf("kubectl %s: %s", strings.Join(cmdArgs, " "), strings.TrimSpace(string(exitErr.Stderr)))
			}
			return nil, "", err
		}
		reader = bytes.NewReader(rawOutput)
	} else {
		reader = cmd.InOrStdin()
	}

	o, kind, err := parser.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("kubectl %s returned decoding input: %w", strings.Join(cmdArgs, " "), err)
	}
	return o, kind, nil
}

var urlPattern = regexp.MustCompile(`https?://`)

// splitURLs splits a string that may contain multiple concatenated URLs
// into individual URLs by finding http:// or https:// boundaries.
func splitURLs(s string) []string {
	locs := urlPattern.FindAllStringIndex(s, -1)
	if len(locs) == 0 {
		return nil
	}
	var urls []string
	for i, loc := range locs {
		start := loc[0]
		var end int
		if i+1 < len(locs) {
			end = locs[i+1][0]
		} else {
			end = len(s)
		}
		urls = append(urls, s[start:end])
	}
	return urls
}

func output(obj runtime.Object, templates []string, urlTemplates []string) {
	renderedTemplates, err := parser.RenderTemplates(obj, templates)
	cobra.CheckErr(err)

	urls, err := parser.RenderTemplates(obj, urlTemplates)
	cobra.CheckErr(err)

	for _, tpl := range renderedTemplates {
		fmt.Println(tpl)
	}

	for _, url := range urls {
		for _, u := range splitURLs(url) {
			fmt.Fprintf(os.Stderr, "Opening: %s\n", u)
			if err := browser.OpenURL(u); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to open URL %s: %v\n", u, err)
			}
		}
	}
}
