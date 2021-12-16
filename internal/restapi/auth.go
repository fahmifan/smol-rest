package restapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/fahmifan/smol/internal/config"
	"github.com/fahmifan/smol/internal/datastore/sqlite"
	"github.com/fahmifan/smol/internal/model"
	"github.com/golang-jwt/jwt/v4"
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

		var user model.User
		ctx := r.Context()

		oldUser, err := s.DataStore.FindUserByEmail(ctx, guser.Email)
		switch err {
		default:
			log.Error().Err(err).Msg("save user")
			writeJSON(rw, http.StatusBadRequest, Map{"error": ErrInternal.Error()})
			return
		case nil:
			user = oldUser
		case sqlite.ErrNotFound:
			guserRawData := &GoogleUserRawData{}
			guserRawData.Parse(guser.RawData)

			newUser, err := model.NewUser(model.RoleUser, guser.Name, guserRawData.Email)
			if err != nil {
				log.Error().Err(err).Msg("create new user")
				writeJSON(rw, http.StatusBadRequest, Map{"error": ErrInvalidArgument.Error()})
				return
			}

			err = s.DataStore.SaveUser(ctx, newUser)
			if err != nil {
				log.Error().Err(err).Msg("save user")
				writeJSON(rw, http.StatusBadRequest, Map{"error": ErrInvalidArgument.Error()})
				return
			}
			user = newUser
		}

		expiredAt := time.Now().Add(time.Hour * 30)
		refreshToken, err := generateRefreshToken(user.ID, expiredAt)
		if err != nil {
			log.Error().Err(err).Msg("generate refresh token")
			writeJSON(rw, http.StatusBadRequest, Map{"error": ErrInternal.Error()})
			return
		}

		accessToken, err := generateAccessToken(user, time.Now().Add(time.Hour))
		if err != nil {
			log.Error().Err(err).Msg("create access token")
			writeJSON(rw, http.StatusBadRequest, Map{"error": ErrInternal.Error()})
			return
		}

		sessModel := model.NewSession(user.ID, refreshToken, expiredAt)
		err = s.DataStore.CreateSession(ctx, sessModel)
		if err != nil {
			log.Error().Err(err).Msg("create session")
			writeJSON(rw, http.StatusBadRequest, Map{"error": ErrInternal.Error()})
			return
		}

		writeJSON(rw, http.StatusOK, LoginResponse{
			UserID:       user.ID.String(),
			RefreshToken: refreshToken,
			AccessToken:  accessToken,
		})
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
			httpError(rw, err, ErrInternal)
			return
		}

		ctx := r.Context()
		oldSess, err := s.DataStore.FindSessionByRefreshToken(ctx, req.RT)
		if errors.Is(err, sqlite.ErrNotFound) {
			httpError(rw, err, ErrNotFound)
			return
		}
		if err != nil {
			httpError(rw, err, ErrInternal)
			return
		}

		user, err := s.DataStore.FindUserByID(ctx, oldSess.UserID)
		if err != nil {
			httpError(rw, err, ErrNotFound)
			return
		}

		refreshToken, err := generateRefreshToken(user.ID, newAccessTokenExpireTime())
		if err != nil {
			httpError(rw, err, ErrInternal)
			return
		}

		newSess := model.NewSession(user.ID, refreshToken, newRefreshTokenExpireTime())
		err = s.DataStore.CreateSession(ctx, newSess)
		if err != nil {
			httpError(rw, err, ErrInternal)
			return
		}
		accessToken, err := generateAccessToken(user, newAccessTokenExpireTime())
		if err != nil {
			httpError(rw, err, ErrInternal)
			return
		}

		httpOK(rw, LoginResponse{
			RefreshToken: newSess.RefreshToken,
			AccessToken:  accessToken,
		})
	}
}

func (s *Server) authorizedAny(perms ...model.Permission) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			token, err := parseTokenFromHeader(r.Header)
			if err != nil {
				log.Error().Err(err).Msg("unable parse token from header")
				httpError(rw, err, ErrInvalidArgument)
				return
			}

			user, ok := auth(token)
			if !ok {
				httpError(rw, nil, ErrUnauthorized)
				return
			}

			setUserToCtx(r.Context(), user)
			if user.IsEmpty() {
				httpError(rw, nil, ErrUnauthorized)
				return
			}

			if !user.Role.GrantedAny(perms...) {
				httpError(rw, nil, ErrPermissionDenied)
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
var jwtKey = []byte(config.JWTSecret())

// Claims jwt claim
type Claims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// GetRoleModel ..
func (c Claims) GetRoleModel() model.Role {
	return model.ParseRole(c.Role)
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

func auth(token string) (model.User, bool) {
	claims, err := parseJWTToken(token)
	if err != nil {
		log.Error().Err(err).Msg("parse jwt")
		return model.User{}, false
	}

	id, err := ulid.Parse(claims.ID)
	if err != nil {
		log.Error().Err(err).Msg("parse id")
		return model.User{}, false
	}
	user := model.User{
		ID:    id,
		Email: claims.Email,
		Role:  claims.GetRoleModel(),
	}

	return user, true
}
