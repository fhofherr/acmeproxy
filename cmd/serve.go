package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the acmeproxy server",
	Long: `
The acmeproxy server obtains certificates from an ACME compliant certificate 
authority. Depending on the operation mode requested by the clientit either 
stores the certificates or directly passes them on to the client. If the 
acmeproxy server stores the certificates locally it takes care of renewing
them before they expire.`,
}
