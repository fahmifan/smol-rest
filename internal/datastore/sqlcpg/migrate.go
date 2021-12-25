package sqlcpg

import (
	"context"
	_ "embed"
	"time"

	"github.com/fahmifan/smol/internal/model/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
)

//go:embed schema.sql
var migrationSQL string

func Migrate(conn *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	res, err := conn.Exec(ctx, migrationSQL)
	models.PanicErr(err)

	rows := res.RowsAffected()
	log.Info().Int64("rowsAffected", rows).Msg("Migrate sqlite3")
}
