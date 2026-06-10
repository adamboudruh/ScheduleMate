package database

import (
	"database/sql"
	// "os"
	// "path/filepath"
	_ "embed"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

//go:embed schema.sql
var schema string

func InitDB(path string) error {
	if path == "" {
		path = "./schedulemate.db"
	}
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	// SQLite is a single local file; using one connection guarantees the
	// foreign_keys pragma below actually applies to every query (a pooled
	// connection could otherwise run with FK enforcement off, breaking
	// ON DELETE CASCADE). Single-connection is fine for a desktop app.
	DB.SetMaxOpenConns(1)

	// for dev, just put db in current directory, but for prod it should be in user config dir
	// dataDir, err := os.UserConfigDir()
	// if err != nil {
	// 	return err
	// }
	// appDir := filepath.Join(dataDir, "ScheduleMate")
	// if err := os.MkdirAll(appDir, 0755); err != nil {
	//     return err
	// }
	// dbPath := filepath.Join(appDir, "schedulemate.db")
	// DB, err = sql.Open("sqlite", dbPath)

	_, err = DB.Exec(`PRAGMA foreign_keys = ON;`) // sqlite disables foreign keys by default
	if err != nil {
		return err
	}

	return createTables()
}

func createTables() error {
	_, err := DB.Exec(schema)
	return err
}
