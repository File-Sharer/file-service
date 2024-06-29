package server

import (
	"context"
	"net/http"

	"github.com/File-Sharer/file-service/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func New() *Server {
	return &Server{}
}

func (s *Server) Run(cfg *config.ServerConfig) error {
	s.httpServer = &http.Server{
		Addr: ":" + cfg.Port,
		Handler: cfg.Handler,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
		ReadTimeout: cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
