package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/server"
	"github.com/fhofherr/golf/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	printErrorAndExit(1,
		viper.BindPFlag(flagACMEDirectoryURLName, serveCmd.Flags().Lookup(flagACMEDirectoryURLName)))
	printErrorAndExit(1,
		viper.BindPFlag(flagHTTPAPIAddrName, serveCmd.Flags().Lookup(flagHTTPAPIAddrName)))
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the acmeproxy server",
	Long: `
Start the acmeproxy server.

The acmeproxy server obtains certificates from an ACME compliant certificate
authority. Depending on the operation mode requested by the client, it either
stores the certificates, or directly passes them on to the client. If the
acmeproxy server stores the certificates locally it takes care of renewing
them before they expire.

The server should work as expected out-of-the box. Certain settings can be
overridden using command line flags. Some flags can also be set using
environment variables. Those flags are marked with [*]. The name of the
environment variable corresponds to the flag name prefixed with 'ACMEPROXY_' and
all hyphens replaced underscores. For example the name of the environment
variable matching the flag '--http-api-addr' would be 'ACMEPROXY_HTTP_API_ADDR'.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO (fhofherr) configure Logger
		var logger log.Logger

		s := &server.Server{
			ACMEDirectoryURL: viper.GetString(flagACMEDirectoryURLName),
			HTTPAPIAddr:      viper.GetString(flagHTTPAPIAddrName),
			Logger:           logger,
		}
		err := s.Start()
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
