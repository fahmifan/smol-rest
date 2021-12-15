package restapi

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fahmifan/smol/internal/config"
	"github.com/fahmifan/smol/internal/datastore/sqlite"
	"github.com/fahmifan/smol/internal/model"
	generated "github.com/fahmifan/smol/internal/restapi/gen"
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
	Port          int
	DB            *sql.DB
	DataStore     sqlite.SQLite
	ServerBaseURL string

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

	baseURL := strings.TrimSuffix(s.ServerBaseURL, "/") + "/api/rest/auth/login/provider/callback?provider=google"
	cookieStore := sessions.NewCookieStore([]byte("secret"))
	gothic.Store = cookieStore
	goth.UseProviders(google.New(config.GoogleClientID(), config.GoogleClientSecret(), baseURL))

	rpcRoute := "/api/oto"
	router.Mount(rpcRoute, s.routeOTO(rpcRoute))

	restRoute := "/api/rest"
	router.Mount(restRoute, s.routeREST())

	router.Group(func(r chi.Router) {
		r.Use(s.mdNonLoginOnly)
		router.Get("/index", renderHTML("index"))
		router.Get("/", renderHTML("index"))
	})
	router.Group(func(r chi.Router) {
		r.Use(s.mdLoginOnly)
		r.Get("/dashboard", renderHTML("dashboard"))
	})

	router.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(unindexed.Dir("./frontend/dist/assets"))))

	return router
}

func (s *Server) mdLoginOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sess := s.session.GetUser(r.Context())
		if sess.Role == model.RoleGuest {
			http.Redirect(rw, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(rw, r)
	})
}

func (s *Server) mdNonLoginOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sess := s.session.GetUser(r.Context())
		if sess.Role != model.RoleGuest {
			http.Redirect(rw, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(rw, r)
	})
}

func renderHTML(page string) http.HandlerFunc {
	writeInternalErr := func(rw http.ResponseWriter) {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("error"))
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		f, err := os.Open(fmt.Sprintf("./frontend/dist/%s/index.html", page))
		if err != nil {
			writeInternalErr(rw)
			return
		}
		defer f.Close()

		s, err := f.Stat()
		if err != nil {
			writeInternalErr(rw)
			return
		}
		if s.IsDir() {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("not found"))
			return
		}

		bt, err := io.ReadAll(f)
		if err != nil {
			log.Error().Err(err).Msg("open index")
			writeInternalErr(rw)
			return
		}
		rw.Header().Add("application/type", "text/html")
		rw.Write(bt)
	}
}

func (s *Server) routeOTO(rpcRoute string) http.Handler {
	greeter := SmolService{s}
	server := otohttp.NewServer()
	server.Basepath = fmtBasepath(rpcRoute)
	generated.RegisterSmolService(server, greeter)
	return server
}

func (s *Server) routeREST() http.Handler {
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
