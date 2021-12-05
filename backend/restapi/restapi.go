package restapi

import (
	"context"
	"net/http"

	"github.com/fahmifan/smol/backend/restapi/generated"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/pacedotdev/oto/otohttp"
	"github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Port       string
	httpServer *http.Server
	session    *SessionManager
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
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	rpcRoute := "/api/oto"
	router.Mount(rpcRoute, s.initOTO(rpcRoute))

	restRoute := "/api/rest"
	router.Mount(restRoute, s.initREST())

	return router
}

func (s *Server) initOTO(rpcRoute string) http.Handler {
	greeter := GreeterService{}
	server := otohttp.NewServer()
	server.Basepath = fmtBasepath(rpcRoute)
	generated.RegisterGreeterService(server, greeter)
	return server
}

func (s *Server) initREST() http.Handler {
	router := chi.NewRouter()
	router.Get("/ping", s.handlePing())
	return router
}

func (s *Server) handlePing() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("pong"))
	}
}

func fmtBasepath(str string) string {
	if val := str[len(str)-1]; string(val) == "/" {
		return str
	}
	return str + "/"
}
