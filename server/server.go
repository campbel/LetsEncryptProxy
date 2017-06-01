package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// Options for configuring a Let's Encrypt Proxy server
type Options struct {
	Domains     []string
	Port        string
	CertCache   string
	ProxyTarget *url.URL
	HealthCheck string
}

// New returns a Let's Encrypt Proxy server using the given options
func New(options Options) *http.Server {

	mux := http.NewServeMux()

	if options.HealthCheck != "" {
		mux.HandleFunc(options.HealthCheck, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "OK")
		})
	}

	mux.Handle("/", httputil.NewSingleHostReverseProxy(options.ProxyTarget))

	var server *http.Server

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(options.Domains...),
		Cache:      autocert.DirCache(options.CertCache),
	}

	server = &http.Server{
		Addr: options.Port,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: mux,
	}

	handleSignals(server)

	return server
}

func handleSignals(server *http.Server) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		fmt.Print("shutting down... ")
		if err := server.Shutdown(ctx); err != nil {
			fmt.Println(err)
		}
		fmt.Println("done")
	}()
}
