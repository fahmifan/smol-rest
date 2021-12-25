package restapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fahmifan/smol/internal/datastore/sqlcpg"
	"github.com/fahmifan/smol/internal/model"
	"github.com/fahmifan/smol/internal/rbac"
	"github.com/fahmifan/smol/internal/usecase"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v4"
	"github.com/markbates/goth/gothic"
	"github.com/oklog/ulid/v2"
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

func (s *Server) HandleLoginProvider() http.HandlerFunc {
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

func (s *Server) HandleLoginProviderCallback() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		guser, err := gothic.CompleteUserAuth(rw, r)
		if err != nil {
			log.Error().Err(err).Stack().Msg("")
			return
		}

		ctx := r.Context()
		_, err = s.Queries.FindUserByEmail(ctx, guser.Email)
		switch err {
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

		case sql.ErrNoRows, pgx.ErrNoRows:
			guserRawData := &GoogleUserRawData{}
			guserRawData.Parse(guser.RawData)

			sess, err := s.Auther.RegisterFromGoth(ctx, usecase.RegisterParams{
				Role:  rbac.RoleUser,
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
		default:
			jsonError(rw, err)
			return
		}
	}
}

func (s *Server) HandleRefreshToken() http.HandlerFunc {
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

func (s *Server) HandleLogout() http.HandlerFunc {
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

func (s *Server) mdAuthorizedAny(perms ...rbac.Permission) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			token, err := parseTokenFromHeader(r.Header)
			if err != nil {
				log.Error().Err(err).Msg("unable parse token from header")
				jsonError(rw, ErrUnauthorized)
				return
			}

			user, ok := auth(token)
			if !ok {
				jsonError(rw, ErrUnauthorized)
				return
			}
			r = r.WithContext(setUserToCtx(r.Context(), user))
			if !user.Role.GrantedAny(perms...) {
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
		return "", ErrInvalidToken
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

// 1 hour
func newAccessTokenExpireTime() time.Time {
	return time.Now().Add(time.Hour)
}

// 1 month
func newRefreshTokenExpireTime() time.Time {
	return time.Now().Add(time.Hour * 24 * 30)
}

// Create the JWT key used to create the signature
var jwtKey []byte

func SetJWTKey(s string) {
	jwtKey = []byte(s)
}

// Claims jwt claim
type Claims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// GetRoleModel ..
func (c Claims) GetRoleModel() rbac.Role {
	return rbac.ParseRole(c.Role)
}

func generateAccessToken(user model.User, expiredAt time.Time) (string, error) {
	claims := &Claims{
		ID:    user.ID.String(),
		Email: user.Email,
		Role:  user.Role.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        model.NewID().String(),
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(expiredAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func generateRefreshToken(userID ulid.ULID, expiredAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID.String(),
		"exp": expiredAt.Unix(),
	})
	rt, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return rt, nil
}

func parseRefreshToken(token string) (sessID ulid.ULID, err error) {
	claims := jwt.MapClaims{}
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		log.Error().Err(err).Msg("")
		return sessID, err
	}

	if tkn != nil && !tkn.Valid {
		return sessID, ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return sessID, ErrInvalidToken
	}
	log.Debug().Str("sub", sub).Msg("")

	sessID, err = ulid.Parse(sub)
	if err != nil {
		return sessID, fmt.Errorf("unable to parse sub into sessID: %w", err)
	}
	return
}

func parseJWTToken(token string) (claims Claims, err error) {
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return claims, err
	}

	if tkn != nil && !tkn.Valid {
		return claims, ErrInvalidToken
	}

	return claims, nil
}

func auth(token string) (sqlcpg.User, bool) {
	claims, err := parseJWTToken(token)
	if err != nil {
		log.Error().Err(err).Msg("parse jwt")
		return sqlcpg.User{}, false
	}

	if err != nil {
		log.Error().Err(err).Msg("parse id")
		return sqlcpg.User{}, false
	}
	user := sqlcpg.User{
		ID:    claims.ID,
		Email: claims.Email,
		Role:  claims.GetRoleModel(),
	}

	return user, true
}
