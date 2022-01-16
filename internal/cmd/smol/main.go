package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/fahmifan/flycasbin/acl"
	"github.com/fahmifan/smol/internal/auth"
	"github.com/fahmifan/smol/internal/config"
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
		Run:   runServer,
	}

	cmd.Flags().Bool("enable-swagger", false, "enable swagger docs")
	return cmd
}

func runServer(cmd *cobra.Command, args []string) {
	enableSwagger := models.StringToBool(cmd.Flag("enable-swagger").Value.String())
	dbPool := sqlcpg.MustOpen(config.PostgresDSN())
	defer dbPool.Close()

	sqlcpg.Migrate(dbPool)
	mustInitACL()

	queries := sqlcpg.New(dbPool)
	auther := &usecase.Auther{
		JWTKey:  []byte(config.JWTSecret()),
		Queries: queries,
	}
	todoer := &usecase.Todoer{
		Queries: queries,
	}

	server := restapi.NewServer(&restapi.ServerConfig{
		Port:          config.Port(),
		ServerBaseURL: config.ServerBaseURL(),
		EnableSwagger: enableSwagger,
		Auther:        auther,
		Queries:       queries,
		Todoer:        todoer,
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

// init ACL panic on error
func mustInitACL() {
	acl, err := acl.NewACL(auth.Policies)
	models.PanicErr(err)

	// set acl to packages
	auth.SetACL(acl)
}
