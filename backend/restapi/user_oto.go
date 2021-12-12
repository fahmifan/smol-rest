package restapi

import (
	"context"

	"github.com/fahmifan/smol/backend/model"
	"github.com/fahmifan/smol/backend/restapi/gen"
	"github.com/rs/zerolog/log"
)

func (s SmolService) FindCurrentUser(ctx context.Context, _ gen.Empty) (*gen.User, error) {
	sess := s.session.GetUser(ctx)
	if sess.Role == model.RoleGuest {
		return nil, nil
	}

	log.Info().Msg(sess.UserID.String())

	user, err := s.DataStore.FindUserByID(ctx, sess.UserID)
	if err != nil {
		log.Error().Err(err).Msg("FindUserByID")
		return nil, err
	}

	res := &gen.User{
		ID:    user.ID.String(),
		Email: user.Email,
		Role:  user.Role.String(),
	}
	return res, nil
}

func (s SmolService) LogoutUser(ctx context.Context, _ gen.Empty) (*gen.Empty, error) {
	s.session.PopUser(ctx)
	return nil, nil
}
