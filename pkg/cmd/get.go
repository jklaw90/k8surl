package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jklaw90/k8surl/pkg/parser"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use: "get",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdArgs := append([]string{"get", "-o", "json"}, args...)
		kubectlCmd := exec.Command("kubectl", cmdArgs...)
		rawOutput, err := kubectlCmd.Output()
		if err != nil {
			return fmt.Errorf("invoking kubectl as %v %v", kubectlCmd.Args, err)
		}

		obj, kind, err := parser.Decode(bytes.NewReader(rawOutput))
		cobra.CheckErr(err)

		kt, ok := config.KindAndTemplates[strings.ToLower(kind)]
		if !ok {
			cobra.CheckErr(fmt.Sprintf("%s isn't defined in the root of the configuration file", kind))
		}

		if !parser.Allowed(kind, []string{kind}) {
			cobra.CheckErr(fmt.Sprintf("%s isn't allowed in this command", kind))
		}

		output(obj, kt.Templates, kt.Urls)
		return nil
	},
}
