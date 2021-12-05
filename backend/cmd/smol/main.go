package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/fahmifan/smol/backend/config"
	"github.com/fahmifan/smol/backend/datastore/sqlite"
	"github.com/fahmifan/smol/backend/restapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

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
	return &cobra.Command{
		Use:   "server",
		Short: "run web server",
		Run: func(cmd *cobra.Command, args []string) {
			log.Logger = zerolog.New(zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: zerolog.TimeFieldFormat,
			}).With().Timestamp().Caller().Logger()

			db := sqlite.MustOpen()
			defer db.Close()

			datastore := sqlite.SQLite{DB: db}

			sqlite.Migrate(db)

			server := restapi.NewServer(&restapi.ServerConfig{
				Port:      config.Port(),
				DB:        db,
				DataStore: datastore,
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
		},
	}
}
