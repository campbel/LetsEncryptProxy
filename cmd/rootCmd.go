package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/campbel/LetsEncryptProxy/server"
	"github.com/spf13/cobra"
)

var (
	domains     []string
	healthCheck string
)

func init() {
	RootCmd.PersistentFlags().StringSliceVarP(&domains, "domain", "D", []string{}, "domains associated with requested cert")
	RootCmd.PersistentFlags().StringVarP(&healthCheck, "health", "H", "", "path to place an HTTP health check (ex. /health)")
}

// RootCmd to be called from main
var RootCmd = &cobra.Command{
	Use:     "leproxy [PROXY URL]",
	Short:   "Leproxy to easily enable HTTPS",
	Long:    "Leproxy is a single host reverse proxy with Let's Encrypt integration",
	Example: "leproxy -D example.com -D www.example.com http://example.io",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) != 1 {
			return errors.New("must supply one argument, the proxy URL")
		}

		proxyTarget, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		proxyServer := server.New(server.Options{
			Domains:     domains,
			Port:        ":443",
			CertCache:   "/var/certs",
			ProxyTarget: proxyTarget,
			HealthCheck: healthCheck,
		})

		fmt.Println("starting up...")
		if err := proxyServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	},
}
