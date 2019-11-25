package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/api"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/golf-zap/golfzap"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	flagACMEDirectoryURLName = "acme-directory-url"
	flagHTTPAPIAddrName      = "http-api-addr"
)

func init() {
	serveCmd.Flags().String(flagACMEDirectoryURLName, acme.DefaultDirectoryURL,
		"Directory URL of the ACME server. [*]")
	serveCmd.Flags().String(flagHTTPAPIAddrName, ":http",
		"TCP address the HTTP API listens on. [*]")

	printErrorAndExit(
		viper.BindPFlag(flagACMEDirectoryURLName, serveCmd.Flags().Lookup(flagACMEDirectoryURLName)))
	printErrorAndExit(
		viper.BindPFlag(flagHTTPAPIAddrName, serveCmd.Flags().Lookup(flagHTTPAPIAddrName)))
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the acmeproxy server",
	Long: `
Start the acmeproxy server.

The acmeproxy server obtains certificates from an ACME compliant certificate
authority. Depending on the operation mode requested by the user, it either
stores the certificates, or directly passes them on to the user. If the
acmeproxy server stores the certificates locally it takes care of renewing
them before they expire.

The server should work as expected out-of-the box. Certain settings can be
overridden using command line flags. Some flags can also be set using
environment variables. Those flags are marked with [*]. The name of the
environment variable corresponds to the flag name prefixed with 'ACMEPROXY_' and
all hyphens replaced underscores. For example the name of the environment
variable matching the flag '--http-api-addr' would be 'ACMEPROXY_HTTP_API_ADDR'.`,
	Run: func(cmd *cobra.Command, args []string) {
		zapLogger, err := zap.NewProduction()
		if err != nil {
			printErrorAndExit(err)
		}
		defer zapLogger.Sync() //nolint: errcheck

		logger := golfzap.New(zapLogger)

		s := &api.Server{
			ACMEDirectoryURL: viper.GetString(flagACMEDirectoryURLName),
			HTTPAPIAddr:      viper.GetString(flagHTTPAPIAddrName),
			Logger:           logger,
		}
		err = s.Start()
		if err != nil {
			fmt.Printf("%+v", err)
			os.Exit(1)
		}
		defer errors.LogFunc(logger, func() error {
			return s.Shutdown(context.Background())
		})
		// Block until we are killed.
		select {}
	},
}
