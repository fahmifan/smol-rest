package restapi

import (
	"net/http"

	"github.com/fahmifan/smol/backend/datastore/sqlite"
	"github.com/fahmifan/smol/backend/model"
	"github.com/fahmifan/smol/backend/model/models"
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

func (s *Server) handleLoginProviderCallback() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		guser, err := gothic.CompleteUserAuth(rw, r)
		if err != nil {
			log.Error().Err(err).Stack().Msg("")
			return
		}

		var sess *Session
		ctx := r.Context()

		oldUser, err := s.DataStore.FindUserByEmail(ctx, guser.Email)
		switch err {
		default:
			log.Error().Err(err).Msg("save user")
			writeJSON(rw, http.StatusBadRequest, Map{"error": ErrInternal.Error()})
			return
		case nil:
			sess = &Session{
				UserID: oldUser.ID.String(),
				Role:   oldUser.Role,
			}
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

			sess = &Session{
				UserID: newUser.ID.String(),
				Role:   newUser.Role,
			}
		}

		s.session.PutUser(r.Context(), sess)
		http.Redirect(rw, r, "/subpage", http.StatusSeeOther)
	}
}

type Map map[string]interface{}

func writeJSON(rw http.ResponseWriter, status int, body interface{}) {
	rw.WriteHeader(status)
	rw.Write(models.JSON(body))
}
