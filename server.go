package queue

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/duh-rpc/duh-go"
	"github.com/kapetan-io/tackle/set"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	// TLS is the TLS config used for public server and clients
	TLS *duh.TLSConfig
	// ListenAddress is the address:port
	ListenAddress string
	// Logger is the logging implementation
	Logger duh.StandardLogger
	// Request Sleep Time
	RequestSleep time.Duration
}

func (c *Config) ClientTLS() *tls.Config {
	if c.TLS != nil {
		return c.TLS.ClientTLS
	}
	return nil
}

func (c *Config) ServerTLS() *tls.Config {
	if c.TLS != nil {
		return c.TLS.ServerTLS
	}
	return nil
}

type Server struct {
	logAdaptor *duh.HttpLogAdaptor
	client     *Client
	server     *http.Server
	wg         sync.WaitGroup
	Listener   net.Listener
	conf       Config
}

func NewServer(ctx context.Context, conf Config) (*Server, error) {
	set.Default(&conf.Logger, slog.Default())

	d := &Server{
		logAdaptor: duh.NewHttpLogAdaptor(conf.Logger),
		conf:       conf,
	}
	return d, d.Start(ctx)
}

func (s *Server) Start(ctx context.Context) error {
	registry := prometheus.NewRegistry()

	handler := NewHTTPHandler(promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	), s.conf)
	registry.MustRegister(handler)

	if s.conf.ServerTLS() != nil {
		if err := s.spawnHTTPS(ctx, handler); err != nil {
			return err
		}
	} else {
		if err := s.spawnHTTP(ctx, handler); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) MustClient() *Client {
	c, err := s.Client()
	if err != nil {
		panic(fmt.Sprintf("failed to init daemon client - '%s'", err))
	}
	return c
}

func (s *Server) Client() (*Client, error) {
	var err error
	if s.client != nil {
		return s.client, nil
	}

	if s.conf.TLS != nil {
		s.client, err = NewClient(WithTLS(s.conf.ClientTLS(), s.Listener.Addr().String()))
		return s.client, err
	}
	s.client, err = NewClient(WithNoTLS(s.Listener.Addr().String()))
	return s.client, err
}

func (s *Server) spawnHTTPS(ctx context.Context, mux http.Handler) error {
	srv := &http.Server{
		ErrorLog:  log.New(s.logAdaptor, "", 0),
		TLSConfig: s.conf.ServerTLS().Clone(),
		Addr:      s.conf.ListenAddress,
		Handler:   mux,
	}

	var err error
	s.Listener, err = net.Listen("tcp", s.conf.ListenAddress)
	if err != nil {
		return fmt.Errorf("while starting HTTPS listener: %w", err)
	}
	srv.Addr = s.Listener.Addr().String()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.conf.Logger.Info("HTTPS Listening ...", "address", s.Listener.Addr().String())
		if err := srv.ServeTLS(s.Listener, "", ""); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.conf.Logger.Error("while starting TLS HTTP server", "error", err)
			}
		}
	}()
	if err := duh.WaitForConnect(ctx, s.Listener.Addr().String(), s.conf.ClientTLS()); err != nil {
		return err
	}

	s.server = srv

	return nil
}

func (s *Server) spawnHTTP(ctx context.Context, h http.Handler) error {
	srv := &http.Server{
		ErrorLog: log.New(s.logAdaptor, "", 0),
		Addr:     s.conf.ListenAddress,
		Handler:  h,
	}
	var err error
	s.Listener, err = net.Listen("tcp", s.conf.ListenAddress)
	if err != nil {
		return fmt.Errorf("while starting HTTP listener: %w", err)
	}
	srv.Addr = s.Listener.Addr().String()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.conf.Logger.Info("HTTP Listening ...", "address", s.Listener.Addr().String())
		if err := srv.Serve(s.Listener); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.conf.Logger.Error("while starting HTTP server", "error", err)
			}
		}
	}()

	if err := duh.WaitForConnect(ctx, s.Listener.Addr().String(), nil); err != nil {
		return err
	}

	s.server = srv
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.conf.Logger.Info("Shutting down server", "address", s.server.Addr)
	_ = s.server.Shutdown(ctx)

	s.server = nil
	return nil
}
