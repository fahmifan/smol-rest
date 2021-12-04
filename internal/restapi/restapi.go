package restapi

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Port       string
	httpServer *http.Server
}

type Server struct {
	*ServerConfig
}

func NewServer(cfg *ServerConfig) *Server {
	return &Server{cfg}
}

func (s *Server) Stop(ctx context.Context) {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown server")
	}
}

func (s *Server) Run() {
	s.httpServer = &http.Server{Addr: s.Port, Handler: s.route()}
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("")
	}
}

func (s *Server) route() chi.Router {
	apiV1 := chi.NewRouter()
	apiV1.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", s.handlePing())
	})

	router := chi.NewRouter()
	router.Mount("/", apiV1)
	return router
}

func (s *Server) handlePing() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("pong"))
	}
}
