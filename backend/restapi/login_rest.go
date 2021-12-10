package restapi

import (
	"net/http"

	"github.com/fahmifan/smol/backend/model"
	"github.com/markbates/goth/gothic"
	"github.com/rs/zerolog/log"
)

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
		s.session.PutUser(r.Context(), user)

		http.Redirect(rw, r, "/subpage", http.StatusSeeOther)
	}
}
