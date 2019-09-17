package app

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/ichiban/assets"

	"github.com/sirupsen/logrus"

	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	http.Server
}

func NewServer(config *Config, l *assets.Locator, handler http.Handler) *Server {
	return &Server{
		Server: http.Server{
			Addr:      config.Bind,
			TLSConfig: tlsConfig(config, l),
			Handler:   handler,
		},
	}
}

func (s *Server) Run() {
	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if s.TLSConfig == nil {
			if err := s.ListenAndServe(); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("stopped listen and serve")
			}
		} else {
			if err := s.ListenAndServeTLS("", ""); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("stopped listen and serve")
			}
		}
	}()

	<-q

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to shutdown")
	}
}

func tlsConfig(config *Config, l *assets.Locator) *tls.Config {
	if config.Host == "" {
		return nil
	}

	m := autocert.Manager{
		Cache:      autocert.DirCache(path.Join(l.Path, "certs")),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(config.Host),
	}

	return m.TLSConfig()
}
