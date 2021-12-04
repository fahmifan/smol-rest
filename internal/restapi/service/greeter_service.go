package service

import (
	"context"

	"github.com/fahmifan/smol/internal/restapi/generated"
)

var _ generated.GreeterService = &GreeterService{}

// GreeterService makes nice greetings.
type GreeterService struct{}

// Greet makes a greeting.
func (GreeterService) Greet(ctx context.Context, r generated.GreetRequest) (*generated.GreetResponse, error) {
	resp := &generated.GreetResponse{
		Greeting: "Hello " + r.Name,
	}
	return resp, nil
}
