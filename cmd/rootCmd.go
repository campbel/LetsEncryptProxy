package cmd

import (
	"errors"
	"fmt"
	"log"
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

		if len(domains) == 0 {
			return errors.New("must supply atleast one domain")
		}

		// Provide an HTTP Health Check for LBs
		healthServer := &http.Server{
			Addr: ":8500",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "OK")
			}),
		}
		go func() {
			if err := healthServer.ListenAndServe(); err != nil {
				log.Println("health server error:", err)
			}
		}()

		// Redirect HTTP to HTTPS
		redirectServer := &http.Server{
			Addr: ":80",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Println("redirecting request", r.Host, r.URL.Path)
				target := "https://" + r.Host + r.URL.Path
				if len(r.URL.RawQuery) > 0 {
					target += "?" + r.URL.RawQuery
				}
				http.Redirect(w, r, target, http.StatusMovedPermanently)
			}),
		}
		go func() {
			if err := redirectServer.ListenAndServe(); err != nil {
				log.Println("redirect server error:", err)
			}
		}()

		// Setup proxy server
		proxyTarget, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		proxyServer := server.New(server.Options{
			Domains:     domains,
			Port:        ":443",
			CertCache:   "/var/certs",
			ProxyTarget: proxyTarget,
		})

		log.Println("starting up...")
		if err := proxyServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	},
}
