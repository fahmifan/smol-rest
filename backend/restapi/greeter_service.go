package restapi

import (
	"context"

	"github.com/fahmifan/smol/backend/restapi/generated"
	"github.com/rs/zerolog/log"
)

var _ generated.GreeterService = &GreeterService{}

// GreeterService makes nice greetings.
type GreeterService struct {
	*Server
}

// Greet makes a greeting.
func (g GreeterService) Greet(ctx context.Context, r generated.GreetRequest) (*generated.GreetResponse, error) {
	sess := g.session.GetUser(ctx)
	log.Debug().Interface("user", sess).Msg("")
	resp := &generated.GreetResponse{
		Greeting: "Hello " + r.Name,
	}
	return resp, nil
}
