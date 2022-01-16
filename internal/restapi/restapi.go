package restapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/fahmifan/smol/internal/auth"
	"github.com/fahmifan/smol/internal/config"
	"github.com/fahmifan/smol/internal/datastore/sqlcpg"
	"github.com/fahmifan/smol/internal/usecase"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/gorilla/sessions"
	"github.com/jordan-wright/unindexed"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/swaggo/http-swagger/example/go-chi/docs"
)

type ServerConfig struct {
	Port          int
	ServerBaseURL string
	EnableSwagger bool
	Queries       *sqlcpg.Queries
	Auther        *usecase.Auther
	Todoer        *usecase.Todoer

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
	s.httpServer = &http.Server{Addr: fmt.Sprintf(":%d", s.Port), Handler: s.route()}
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

	if s.EnableSwagger {
		router.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(fmt.Sprintf("%s/assets/swagger/swagger.json", s.ServerBaseURL)),
		))

		// serve static file, use unindexed.Dir to prevent directory traversal
		router.Handle("/assets/swagger/*", http.StripPrefix("/assets/swagger/", http.FileServer(unindexed.Dir("./swagger"))))
	}

	baseURL := strings.TrimSuffix(s.ServerBaseURL, "/") + "/api/auth/login/provider/callback?provider=google"
	cookieStore := sessions.NewCookieStore([]byte("secret"))
	gothic.Store = cookieStore
	goth.UseProviders(google.New(config.GoogleClientID(), config.GoogleClientSecret(), baseURL))

	router.Mount("/api", s.router())
	return router
}

func (s *Server) router() http.Handler {
	router := chi.NewRouter()
	router.Get("/ping", s.handlePing())
	router.Get("/auth/login/oauth2", s.handleLoginProvider())
	router.Get("/auth/login/provider/callback", s.handleLoginProviderCallback())
	router.Post("/auth/refresh", s.handleRefreshToken())

	router.Method(http.MethodGet, "/auth/logout", s.mdAuthorizedAny()(s.handleLogout()))
	router.Method(http.MethodPost, "/todos", s.mdAuthorizedAny(auth.Perm(auth.Create, auth.Todo))(s.handleCreateTodo()))
	router.Method(http.MethodGet, "/todos", s.mdAuthorizedAny(auth.Perm(auth.ViewSelf, auth.Todo))(s.handleFindAllTodos()))

	return router
}

func (s *Server) handlePing() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("pong"))
	}
}
