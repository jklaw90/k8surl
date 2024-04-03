package main

import (
	"os"

	"github.com/jklaw90/k8surl/pkg/cmd"
)

func main() {
	err := cmd.NewK8surlCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
