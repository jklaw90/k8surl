package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jklaw90/k8surl/pkg/parser"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	config *Config
)

var k8surlCmd = &cobra.Command{
	Use:   "k8surl",
	Short: "CLI to read k8s resources and open links based on your template config",
	Long: `pipe in kubectl output to k8surl or pass in arguments like you would to kubectl.
The config should be formatted with kind and array of urls for the type like:
pod:
  - https://mydashboard?pod_name={.metadata.name}
ingress:
  - ....
To launch our pod urls enter:
kubectl get pod nginx-xyz -o json | k8surl
OR
k8surl get pod nginx-xyz (eventually....)`,
	Run: func(cmd *cobra.Command, args []string) {
		obj, kind, err := parser.Decode(cmd.InOrStdin())
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

func Execute() {
	err := k8surlCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	initConfig()
	browser.Stdout = nil // not sure if we should do something better here
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".k8surl")
	err = viper.ReadInConfig()
	cobra.CheckErr(err)

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Fprintln(os.Stderr)
		cobra.CheckErr(fmt.Errorf("failed to unmsrshal config: %s", err.Error()))
	}

	for k, v := range config.Commands {
		k, v := k, v
		dynamicCmd := &cobra.Command{
			Use:   k,
			Short: fmt.Sprintf("dynamic command for %s", k),
			Run: func(cmd *cobra.Command, args []string) {
				obj, kind, err := parser.Decode(cmd.InOrStdin())
				cobra.CheckErr(err)

				if !parser.Allowed(kind, v.Kinds) {
					cobra.CheckErr(fmt.Sprintf("%s is not allowed in this command", kind))
				}
				output(obj, v.Templates, v.Urls)
			},
		}
		k8surlCmd.AddCommand(dynamicCmd)
	}
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
		allUrls := strings.Split(url, "https://")
		for _, u := range allUrls {
			browser.OpenURL(fmt.Sprintf("https://%s", u)) // nolint: errcheck
		}
	}
}
