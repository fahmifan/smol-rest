package sqlite

import (
	"database/sql"
	_ "embed"

	"github.com/fahmifan/smol/backend/model/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

//go:embed migration.sql
var migrationSQL string

func Migrate(db *sql.DB) {
	res, err := db.Exec(migrationSQL)
	models.PanicErr(err)

	rows, err := res.RowsAffected()
	models.PanicErr(err)
	log.Info().Int64("rowsAffected", rows).Msg("Migrate sqlite3")
}

func MustOpen() *sql.DB {
	db, err := sql.Open("sqlite3", "smol_sqlite3.db?cache=shared&mode=rwc&_journal_mode=WAL")
	models.PanicErr(err)
	return db
}

type SQLite struct {
	DB *sql.DB
}
