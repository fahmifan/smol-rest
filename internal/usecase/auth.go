package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fahmifan/smol/internal/auth"
	"github.com/fahmifan/smol/internal/datastore"
	"github.com/fahmifan/smol/internal/datastore/sqlcpg"
	"github.com/fahmifan/smol/internal/model"
	"github.com/golang-jwt/jwt/v4"
	"github.com/markbates/goth"
	"github.com/rs/zerolog/log"
)

type Auther struct {
	JWTKey  []byte
	Queries *sqlcpg.Queries
}

// UserToken represent user data in the token
type UserToken struct {
	ID    string
	Email string
	Role  auth.Role
}

// AuthenticateToken authenticate the access token and return a subset of user data.
// It doesn't do a query and only return information on the token they are user ID, Email & Role
func (a *Auther) AuthenticateToken(accessToken string) (UserToken, bool) {
	claims, err := a.parseJWTToken(accessToken)
	if err != nil {
		log.Error().Err(err).Msg("parse jwt")
		return UserToken{}, false
	}

	if err != nil {
		log.Error().Err(err).Msg("parse id")
		return UserToken{}, false
	}
	user := UserToken{
		ID:    claims.ID,
		Email: claims.Email,
		Role:  claims.GetRoleModel(),
	}
	return user, true
}

func (a *Auther) parseJWTToken(token string) (claims jwtClaims, err error) {
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return a.JWTKey, nil
	})
	if err != nil {
		return claims, err
	}

	if tkn != nil && !tkn.Valid {
		return claims, ErrInvalidToken
	}

	return claims, nil
}

// LoginFromGoth authenticate user & create new session from the Goth oauth2 flow
func (a *Auther) LoginFromGoth(ctx context.Context, guser goth.User) (sess sqlcpg.Session, err error) {
	user, err := a.Queries.FindUserByEmail(ctx, guser.Email)
	if err == sql.ErrNoRows {
		err = datastore.ErrNotFound
		return
	}
	if err != nil {
		err = fmt.Errorf("unable to find user by email: %w", err)
		return
	}

	sess, err = a.createSession(ctx, user)
	return
}

type RegisterParams struct {
	Role  auth.Role
	Name  string
	Email string
}

func (a *Auther) RegisterFromGoth(ctx context.Context, arg RegisterParams) (sess sqlcpg.Session, err error) {
	user, err := a.Queries.SaveUser(ctx, sqlcpg.SaveUserParams{
		ID:    model.NewID().String(),
		Name:  sql.NullString{String: arg.Name, Valid: true},
		Email: arg.Email,
		Role:  arg.Role,
	})
	if err != nil {
		err = fmt.Errorf("unable to save user: %w", err)
		return
	}

	sess, err = a.createSession(ctx, user)
	return
}

func (a *Auther) RefreshToken(ctx context.Context, refreshToken string) (sess sqlcpg.Session, err error) {
	_, err = a.parseRefreshToken(refreshToken)
	if err != nil {
		return
	}

	oldSess, err := a.Queries.FindSessionByRefreshToken(ctx, refreshToken)
	// oldSess, err := s.DataStore.FindSessionByRefreshToken(ctx, req.RT)
	if err != nil {
		return
	}

	if isSessionExpired(oldSess) {
		return
	}

	user, err := a.Queries.FindUserByID(ctx, oldSess.UserID)
	if err != nil {
		return
	}

	accessTokenExpiredAt := newAccessTokenExpireTime()
	accessToken, err := a.generateAccessToken(user, accessTokenExpiredAt)
	if err != nil {
		return
	}

	sess, err = a.Queries.UpdateAccessToken(ctx, sqlcpg.UpdateAccessTokenParams{
		AccessToken:          accessToken,
		AccessTokenExpiredAt: accessTokenExpiredAt,
		ID:                   oldSess.ID,
	})
	if err != nil {
		return
	}

	return
}

func (a *Auther) parseRefreshToken(token string) (sessID string, err error) {
	claims := jwt.MapClaims{}
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return a.JWTKey, nil
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

	if err != nil {
		return sub, fmt.Errorf("unable to parse sub into sessID: %w", err)
	}

	sessID = sub
	return
}

func (a *Auther) createSession(ctx context.Context, user sqlcpg.User) (sess sqlcpg.Session, err error) {
	accessTokenExpiredAt := newAccessTokenExpireTime()
	accessToken, err := a.generateAccessToken(user, accessTokenExpiredAt)
	if err != nil {
		err = fmt.Errorf("unable to generate access token: %w", err)
		return
	}

	expiredAt := newRefreshTokenExpireTime()
	refreshToken, err := a.generateRefreshToken(user.ID, expiredAt)
	if err != nil {
		err = fmt.Errorf("unable to generate refresh token: %w", err)
		return
	}

	sess, err = a.Queries.SaveSession(ctx, sqlcpg.SaveSessionParams{
		ID:                    model.NewID().String(),
		UserID:                user.ID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiredAt: expiredAt,
		AccessToken:           accessToken,
		AccessTokenExpiredAt:  accessTokenExpiredAt,
	})
	if err != nil {
		err = fmt.Errorf("unable to save session: %w", err)
		return
	}

	return
}

type jwtClaims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// GetRoleModel ..
func (c jwtClaims) GetRoleModel() auth.Role {
	return auth.Role(c.Role)
}

func (a *Auther) generateAccessToken(user sqlcpg.User, expiredAt time.Time) (string, error) {
	claims := &jwtClaims{
		ID:    user.ID,
		Email: user.Email,
		Role:  string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        model.NewID().String(),
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expiredAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.JWTKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Auther) generateRefreshToken(userID string, expiredAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": expiredAt.Unix(),
	})
	rt, err := token.SignedString(a.JWTKey)
	if err != nil {
		return "", err
	}

	return rt, nil
}

// 1 month
func newRefreshTokenExpireTime() time.Time {
	return time.Now().Add(time.Hour * 24 * 30)
}

// 1 hour
func newAccessTokenExpireTime() time.Time {
	return time.Now().Add(time.Hour)
}

func isSessionExpired(s sqlcpg.Session) bool {
	return time.Now().After(s.RefreshTokenExpiredAt)
}
