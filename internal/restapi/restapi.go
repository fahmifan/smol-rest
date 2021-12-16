package restapi

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/jordan-wright/unindexed"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/swaggo/http-swagger/example/go-chi/docs"

	"github.com/fahmifan/smol/internal/config"
	"github.com/fahmifan/smol/internal/datastore/sqlite"
	"github.com/fahmifan/smol/internal/model"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Port          int
	DB            *sql.DB
	DataStore     sqlite.SQLite
	ServerBaseURL string
	EnableSwagger bool

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

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("%s/assets/swagger/swagger.json", s.ServerBaseURL)),
	))

	// serve static file, use unindexed.Dir to prevent directory traversal
	router.Handle("/assets/swagger/*", http.StripPrefix("/assets/swagger/", http.FileServer(unindexed.Dir("./swagger"))))

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
	router.Method("POST", "/auth/refresh", s.mdAuthorizedAny()(s.handleRefreshToken()))

	router.Method("POST", "/todos", s.mdAuthorizedAny(model.Create_Todo)(s.handleCreateTodo()))
	router.Method("GET", "/todos", s.mdAuthorizedAny(model.View_AllSelfTodo)(s.handleFindAllTodos()))

	return router
}

func (s *Server) handlePing() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("pong"))
	}
}
