package rest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/seggga/approve-analytics/internal/ports"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// Server is a REST server
// @title Blueprint Swagger API
// @version 1.0
// @description Swagger API for analytics service.
// @termsOfService http://swagger.io/terms/
// @securityDefinitions.basic Auth
// @authorizationurl /validate
// @name token
// @contact.name API Support
// @contact.email test@gmail.com
//
// @license.name MIT
// @license.url https://github.com/MartinHeinz/go-project-blueprint/blob/master/LICENSE
type Server struct {
	auth     ports.Auther
	server   *http.Server
	logger   *zap.Logger
	an       ports.Analyter
	listener net.Listener
}

// New ...
func New(logger *zap.Logger, auth ports.Auther, an ports.Analyter, port string) *Server {
	var err error
	s := &Server{
		auth:   auth,
		logger: logger,
		an:     an,
	}

	s.listener, err = net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Sugar().Fatalf("error creating listener on port %s: %v", port, err)
	}
	s.server = &http.Server{
		Handler: s.routes(),
	}

	return s
}

// Start starts the REST server
func (s *Server) Start() error {
	s.logger.Debug("starting REST server ...")

	if err := s.server.Serve(s.listener); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("cannot start REST server: %v", err)
	}
	return nil
}

// Stop grecefuly terminages the REST server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Debug("stopping REST server ...")

	return s.server.Shutdown(ctx)
}

func (s *Server) routes() http.Handler {
	r := chi.NewMux()

	r.Use(middleware.Heartbeat("/healthz"))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Mount("/swagger", httpSwagger.WrapHandler)

	r.Mount("/", s.Handlers())

	return r
}
