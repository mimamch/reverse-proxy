package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/mimamch/reverse-proxy/internal/config"
	"github.com/mimamch/reverse-proxy/internal/modules/certificate"
	"github.com/mimamch/reverse-proxy/internal/modules/proxy"
	"github.com/mimamch/reverse-proxy/internal/utils"
	"github.com/mimamch/reverse-proxy/pkg/database"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sys/unix"
)

func main() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if err := utils.InitTrustedCIDRs(); err != nil {
		log.Fatalf("Failed to initialize Trusted CIDRs: %v", err)
	}

	r := chi.NewRouter()

	// r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	cfg := config.LoadConfig()
	db := database.ConnectPostgres(cfg)
	proxyService := proxy.NewService(proxy.NewRepository(db), proxy.NewProxyCache())
	proxyHandler := proxy.NewHandler(proxyService)

	r.HandleFunc("/*", proxyHandler.HandleRequest)

	n := runtime.NumCPU()
	go func() {
		log.Printf("Starting %d HTTP server(s) on :80", n)
		for range n {
			ln, err := reusePortListen(":80")
			if err != nil {
				log.Fatalf("Failed to listen on :80: %v", err)
			}
			server := &http.Server{
				Handler: r,
			}
			go server.Serve(ln)
		}
	}()

	certCache := certificate.NewCertCache()
	certRepo := certificate.NewRepository(db)
	certService := certificate.NewService(certRepo, certCache)
	manager := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Email:  cfg.Email,
	}
	tlsConfig := certificate.NewTLSConfig(certificate.NewCertCache(), certService, manager)
	// server := &http.Server{
	// 	Addr:      ":443",
	// 	TLSConfig: tlsConfig,
	// 	Handler:   r,
	// }

	go func() {
		log.Printf("Starting %d HTTPS server(s) on :443", n)
		for range n {
			ln, err := reusePortListen(":443")
			if err != nil {
				log.Fatalf("Failed to listen on :443: %v", err)
			}
			server := &http.Server{
				TLSConfig: tlsConfig,
				Handler:   r,
			}
			go server.ServeTLS(ln, "", "")
		}
	}()

	// run forever
	<-sigs
	log.Println("Signal received, shutting down gracefully...")
}

func reusePortListen(addr string) (net.Listener, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = unix.SetsockoptInt(
					int(fd),
					unix.SOL_SOCKET,
					unix.SO_REUSEPORT,
					1,
				)
			})
			return err
		},
	}
	return lc.Listen(context.Background(), "tcp", addr)
}
