package restapi

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/fahmifan/smol/backend/config"
	"github.com/fahmifan/smol/backend/model"
	"github.com/fahmifan/smol/backend/restapi/generated"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/pacedotdev/oto/otohttp"
	"github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Port       int
	DB         *sql.DB
	httpServer *http.Server
	session    *SessionManager
}

type Server struct {
	*ServerConfig
}

func NewServer(cfg *ServerConfig) *Server {
	cfg.session = NewSessionManager(cfg.DB)
	return &Server{cfg}
}

func (s *Server) Stop(ctx context.Context) {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown server")
	}
}

func (s *Server) Run() {
	handler := s.session.session.LoadAndSave(s.route())
	s.httpServer = &http.Server{Addr: fmt.Sprintf(":%d", s.Port), Handler: handler}
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

	baseURL := "http://localhost:9000/api/rest/auth/login/provider/callback?provider=google"
	cookieStore := sessions.NewCookieStore([]byte("secret"))
	gothic.Store = cookieStore
	goth.UseProviders(google.New(config.GoogleClientID(), config.GoogleClientSecret(), baseURL))

	rpcRoute := "/api/oto"
	router.Mount(rpcRoute, s.initOTO(rpcRoute))

	restRoute := "/api/rest"
	router.Mount(restRoute, s.initREST())

	return router
}

func (s *Server) initOTO(rpcRoute string) http.Handler {
	greeter := GreeterService{s}
	server := otohttp.NewServer()
	server.Basepath = fmtBasepath(rpcRoute)
	generated.RegisterGreeterService(server, greeter)
	return server
}

func (s *Server) initREST() http.Handler {
	router := chi.NewRouter()
	router.Get("/ping", s.handlePing())
	router.Get("/auth/login/oauth2", s.handleLoginProvider())
	router.Get("/auth/login/provider/callback", s.handleLoginProviderCallback())
	return router
}

func (s *Server) handlePing() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("pong"))
	}
}

func (s *Server) handleLoginProvider() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Info().Msg("masook")
		if _, err := gothic.CompleteUserAuth(rw, r); err == nil {
			log.Error().Err(err).Msg("")
			http.Redirect(rw, r, "/", http.StatusSeeOther)
			return
		}

		gothic.BeginAuthHandler(rw, r)
	}
}

func (s *Server) handleLoginProviderCallback() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		guser, err := gothic.CompleteUserAuth(rw, r)
		if err != nil {
			log.Error().Err(err).Stack().Msg("")
			return
		}

		user := &Session{
			UserID: guser.UserID,
			Role:   model.RoleUser,
		}
		log.Debug().Interface("user", user).Msg("")
		s.session.PutUser(r.Context(), user)

		http.Redirect(rw, r, "/subpage", http.StatusSeeOther)
	}
}

func fmtBasepath(str string) string {
	if val := str[len(str)-1]; string(val) == "/" {
		return str
	}
	return str + "/"
}
