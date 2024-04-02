package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
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
		Use:                "k8surl",
		SilenceUsage:       true,
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs, // allows us to not require -- for kubectl args
		CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
		Short:              "CLI to read k8s resources and open urls based on your template config",
		Example: `kubectl get pod <pod-name> -o json | k8surl pod
kubectl k8surl pod <pod-name>`,
		Run: func(cmd *cobra.Command, args []string) {
			obj, kind, err := commandInitilizer(cmd, args)
			cobra.CheckErr(err)

			kt, ok := config.KindAndTemplates[strings.ToLower(kind)]
			if !ok {
				cobra.CheckErr(fmt.Sprintf("%s isn't defined in the root of the configuration file", kind))
			}

			if !parser.Allowed(kind, []string{kind}) {
				cobra.CheckErr(fmt.Sprintf("%s isn't allowed in this command", kind))
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
	cobra.CheckErr(viper.ReadInConfig())
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
		k, v := k, v
		dynamicCmd := &cobra.Command{
			Use:                k,
			DisableFlagParsing: true,
			DisableSuggestions: true,
			Short:              fmt.Sprintf("dynamic command for %s", k),
			Run: func(cmd *cobra.Command, args []string) {
				obj, kind, err := commandInitilizer(cmd, args)
				cobra.CheckErr(err)

				if !parser.Allowed(kind, v.Kinds) {
					cobra.CheckErr(fmt.Sprintf("%s is not allowed in this command", kind))
				}
				output(obj, v.Templates, v.Urls)
			},
		}
		rootCmd.AddCommand(dynamicCmd)
	}
}

// commandInitilizer is a helper function to decode the input from the command line or stdin
func commandInitilizer(cmd *cobra.Command, args []string) (runtime.Object, string, error) {
	if slices.ContainsFunc(args, func(arg string) bool {
		return strings.EqualFold(arg, "--help") || strings.EqualFold(arg, "-h")
	}) {
		cmd.Help()
		os.Exit(0)
	}

	var reader io.Reader
	var cmdArgs []string
	if len(args) > 0 {
		if !slices.ContainsFunc(args, func(arg string) bool {
			return strings.EqualFold(arg, "-o") || strings.EqualFold(arg, "--output")
		}) {
			cmdArgs = append(args, []string{"-o", "json"}...)
		}

		kubectlCmd := exec.Command("kubectl", cmdArgs...)
		rawOutput, err := kubectlCmd.Output()
		if err != nil {
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

func output(obj runtime.Object, templates []string, urlTemplates []string) {
	renderedTemplates, err := parser.RenderTemplates(obj, templates)
	cobra.CheckErr(err)

	urls, err := parser.RenderTemplates(obj, urlTemplates)
	cobra.CheckErr(err)

	for _, tpl := range renderedTemplates {
		fmt.Println(tpl)
	}

	for _, url := range urls {
		// TODO cleanup - template will be a single string we need to split somehow...
		allUrls := strings.Split(strings.TrimPrefix(url, "https://"), "https://")
		for _, u := range allUrls {
			browser.OpenURL(fmt.Sprintf("https://%s", u)) // nolint: errcheck
		}
	}
}
