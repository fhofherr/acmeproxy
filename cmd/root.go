package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "acmeproxy",
	Short: "acmeproxy is a proxy for ACME compliant certificate authorities",
}

// Execute starts acmeproxy and executes the command given on the command line.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
