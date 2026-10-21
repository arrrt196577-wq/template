package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
)

type Server struct {
	server *http.Server
	cfg    config.HTTP
	log    *slog.Logger
}

func New(cfg config.HTTP, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		server: &http.Server{
			Addr:              cfg.Address,
			Handler:           handler,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
			MaxHeaderBytes:    1 << 20,
			// net/http's default error logger may dump connection details.
			ErrorLog: slog.NewLogLogger(log.Handler(), slog.LevelError),
		},
		cfg: cfg,
		log: log,
	}
}

func (server *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", server.cfg.Address)
	if err != nil {
		return errors.New("HTTP listener initialization failed")
	}
	return server.Serve(ctx, listener)
}

// Serve takes ownership of listener and blocks until shutdown finishes.
func (server *Server) Serve(ctx context.Context, listener net.Listener) error {
	server.server.BaseContext = func(net.Listener) context.Context {
		// Keep in-flight requests alive while Shutdown drains them.
		return context.WithoutCancel(ctx)
	}
	result := make(chan error, 1)
	go func() { result <- server.server.Serve(listener) }()
	server.log.Info("http server started", "address", listener.Addr().String())
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errors.New("HTTP server stopped unexpectedly")
	case <-ctx.Done():
		server.log.Info("http server draining")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), server.cfg.ShutdownTimeout)
		defer cancel()
		if err := server.server.Shutdown(shutdownCtx); err != nil {
			_ = server.server.Close()
			<-result
			return errors.New("HTTP graceful shutdown deadline exceeded")
		}
		<-result
		server.log.Info("http server stopped")
		return nil
	}
}
