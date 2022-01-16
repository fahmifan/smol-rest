package restapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/fahmifan/smol/internal/auth"
	"github.com/fahmifan/smol/internal/datastore/sqlcpg"
	"github.com/fahmifan/smol/internal/usecase"
	"github.com/jackc/pgx/v4"
	"github.com/markbates/goth/gothic"
	"github.com/rs/zerolog/log"
)

type GoogleUserRawData struct {
	Email         string `json:"email"`
	ID            string `json:"id"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func (g *GoogleUserRawData) Parse(m Map) {
	if m == nil {
		return
	}

	g.Email = m["email"].(string)
	g.ID = m["id"].(string)
	g.Picture = m["picture"].(string)
}

func (s *Server) handleLoginProvider() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if _, err := gothic.CompleteUserAuth(rw, r); err == nil {
			log.Error().Err(err).Msg("")
			http.Redirect(rw, r, "/", http.StatusSeeOther)
			return
		}

		gothic.BeginAuthHandler(rw, r)
	}
}

type LoginResponse struct {
	UserID       string `json:"userID"`
	RefreshToken string `json:"refreshToken"`
	AccessToken  string `json:"accessToken"`
}

func (s *Server) handleLoginProviderCallback() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		guser, err := gothic.CompleteUserAuth(rw, r)
		if err != nil {
			log.Error().Err(err).Stack().Msg("")
			return
		}

		ctx := r.Context()
		_, err = s.Queries.FindUserByEmail(ctx, guser.Email)
		switch err {
		default:
			jsonError(rw, err)
			return
		case sql.ErrNoRows, pgx.ErrNoRows:
			guserRawData := &GoogleUserRawData{}
			guserRawData.Parse(guser.RawData)

			sess, err := s.Auther.RegisterFromGoth(ctx, usecase.RegisterParams{
				Role:  auth.RoleUser,
				Email: guserRawData.Email,
				Name:  guser.Name,
			})
			if err != nil {
				jsonError(rw, err)
				return
			}

			jsonOK(rw, LoginResponse{
				UserID:       sess.UserID,
				RefreshToken: sess.RefreshToken,
				AccessToken:  sess.AccessToken,
			})
		case nil:
			sess, err := s.Auther.LoginFromGoth(ctx, guser)
			if err != nil {
				jsonError(rw, err)
				return
			}

			jsonOK(rw, LoginResponse{
				UserID:       sess.UserID,
				RefreshToken: sess.RefreshToken,
				AccessToken:  sess.AccessToken,
			})
		}
	}
}

func (s *Server) handleRefreshToken() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		req := struct {
			RT string `json:"refreshToken"`
		}{}

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			jsonError(rw, err)
			return
		}

		ctx := r.Context()
		sess, err := s.Auther.RefreshToken(ctx, req.RT)
		if err != nil {
			jsonError(rw, err)
			return
		}

		jsonOK(rw, LoginResponse{
			UserID:       sess.UserID,
			RefreshToken: sess.RefreshToken,
			AccessToken:  sess.AccessToken,
		})
	}
}

func IsExpired(s sqlcpg.Session) bool {
	return time.Now().After(s.RefreshTokenExpiredAt)
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getUserFromCtx(ctx)
		_, err := s.Queries.DeleteSessionByUserID(ctx, user.ID)
		if err != nil {
			jsonError(rw, err)
			return
		}

		jsonOK(rw, Map{"status": "success"})
	}
}

func (s *Server) mdAuthorizedAny(perms ...auth.Permission) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			token, err := parseTokenFromHeader(r.Header)
			if err != nil {
				jsonError(rw, err)
				return
			}

			user, ok := s.Auther.AuthenticateToken(token)
			if !ok {
				jsonError(rw, ErrUnauthorized)
				return
			}
			r = r.WithContext(setUserToCtx(r.Context(), user))
			if len(perms) == 0 {
				next.ServeHTTP(rw, r)
				return
			}

			err = auth.GrantedAny(user.Role, perms...)
			if err != nil {
				jsonError(rw, nil)
				return
			}

			next.ServeHTTP(rw, r)
		})
	}
}

func parseTokenFromHeader(header http.Header) (string, error) {
	var token string

	authHeaders := strings.Split(header.Get("Authorization"), " ")
	if len(authHeaders) != 2 {
		return "", ErrMissingAuthorizationHeader
	}

	if authHeaders[0] != "Bearer" {
		err := ErrMissingAuthorizationHeader
		log.Error().Err(err).Str("Authorization", authHeaders[0]).Msg("")
		return token, err
	}

	token = strings.Trim(authHeaders[1], " ")
	if token == "" {
		err := ErrMissingAuthorizationHeader
		log.Error().Err(err).Msg("")
		return token, err
	}

	return token, nil
}
