package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kubehelp",
		Short: "Kubernetes troubleshooting CLI",
		Long:  `kubehelp assists with troubleshooting Kubernetes deployments via subcommands.`,
	}

	rootCmd.AddCommand(diagnoseCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
