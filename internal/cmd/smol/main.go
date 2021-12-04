package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/fahmifan/smol/internal/restapi"
	"github.com/rs/zerolog/log"
)

func main() {
	server := restapi.NewServer(&restapi.ServerConfig{
		Port: ":8080",
	})
	log.Info().Msg("run server at " + server.Port)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go server.Run()

	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// stop web & rest api server first to stop sending jobs to worker
	log.Info().Msg("server stopped")
	server.Stop(ctx)
}
