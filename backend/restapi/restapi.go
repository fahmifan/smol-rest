package restapi

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/fahmifan/smol/backend/config"
	"github.com/fahmifan/smol/backend/datastore/sqlite"
	generated "github.com/fahmifan/smol/backend/restapi/gen"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/gorilla/sessions"
	"github.com/jordan-wright/unindexed"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/pacedotdev/oto/otohttp"
	"github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Port       int
	DB         *sql.DB
	DataStore  sqlite.SQLite
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

	router.Route("/", func(r chi.Router) {
		r.Get("/", renderHTML("index"))
		r.Get("/index", renderHTML("index"))
		r.Get("/dashboard", renderHTML("dashboard"))
		r.Get("/subpage", renderHTML("subpage"))
	})
	router.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(unindexed.Dir("./web/dist/assets"))))

	return router
}

func renderHTML(page string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		f, err := os.Open(fmt.Sprintf("./web/dist/%s/index.html", page))
		if err != nil {
			log.Error().Err(err).Msg("open index")
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("error"))
			return
		}
		defer f.Close()
		bt, err := io.ReadAll(f)
		if err != nil {
			log.Error().Err(err).Msg("open index")
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("error"))
			return
		}
		rw.Header().Add("application/type", "text/html")
		rw.Write(bt)
	}
}

func (s *Server) initOTO(rpcRoute string) http.Handler {
	greeter := SmolService{s}
	server := otohttp.NewServer()
	server.Basepath = fmtBasepath(rpcRoute)
	generated.RegisterSmolService(server, greeter)
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

func fmtBasepath(str string) string {
	if val := str[len(str)-1]; string(val) == "/" {
		return str
	}
	return str + "/"
}
