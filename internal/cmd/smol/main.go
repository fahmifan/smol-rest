package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/fahmifan/smol/internal/config"
	"github.com/fahmifan/smol/internal/datastore/postgres"
	"github.com/fahmifan/smol/internal/datastore/sqlcpg"
	"github.com/fahmifan/smol/internal/model/models"
	"github.com/fahmifan/smol/internal/restapi"
	"github.com/fahmifan/smol/internal/usecase"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// @title SMOL API
// @version 1.0
// @description API documentation for SMOL Service

// @BasePath /
func main() {
	cmd := &cobra.Command{
		Use:   "smol",
		Short: "smol cli",
	}
	cmd.AddCommand(
		serverCMD(),
	)
	cmd.Execute()
}

func serverCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "run web server",
	}

	cmd.Flags().Bool("enable-swagger", false, "enable swagger docs")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		enableSwagger := models.StringToBool(cmd.Flag("enable-swagger").Value.String())
		dbPool := postgres.MustOpen(config.PostgresDSN())
		defer dbPool.Close()

		sqlcpg.Migrate(dbPool)

		queries := sqlcpg.New(dbPool)
		auther := &usecase.Auther{
			JWTKey:  []byte(config.JWTSecret()),
			Queries: queries,
		}

		restapi.SetJWTKey(config.JWTSecret())
		server := restapi.NewServer(&restapi.ServerConfig{
			Port:          config.Port(),
			ServerBaseURL: config.ServerBaseURL(),
			EnableSwagger: enableSwagger,
			Auther:        auther,
			Queries:       queries,
		})
		log.Info().Int("port", config.Port()).Msg("server runs")
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

	return cmd
}
