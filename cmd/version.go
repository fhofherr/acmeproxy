package cmd

import (
	"fmt"

	"github.com/fhofherr/acmeproxy/pkg/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show acmeproxy's version",
	Run: func(cmd *cobra.Command, args []string) {
		if version.GitTag != "" {
			fmt.Printf("Version: %s\n", version.GitTag)
		}
		fmt.Printf("Build Time: %s\n", version.BuildTime)
		fmt.Printf("Git Hash: %s\n", version.GitHash)
	},
}
