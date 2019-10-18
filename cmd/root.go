package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	viperEnvPrefix = "acmeproxy"
)

func init() {
	viper.SetEnvPrefix(viperEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

var rootCmd = &cobra.Command{
	Use:   "acmeproxy",
	Short: "acmeproxy is a proxy for ACME compliant certificate authorities",
}

// Execute starts acmeproxy and executes the command given on the command line.
func Execute() {
	printErrorAndExit(rootCmd.Execute())
}

func printErrorAndExit(err error) {
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
