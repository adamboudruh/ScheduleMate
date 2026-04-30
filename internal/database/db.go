package database

import (
	"database/sql"
	// "os"
	// "path/filepath"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(path string) error {
	if path == "" {
		path = "./schedulemate.db"
	}
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}

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
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS employee (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			name          TEXT NOT NULL,
			role          TEXT,
			desired_hours INTEGER NOT NULL,
			wage          REAL
		);

		CREATE TABLE IF NOT EXISTS availability (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL,
			day_of_week INTEGER NOT NULL,
			start_time  TEXT NOT NULL,
			end_time    TEXT NOT NULL,
			FOREIGN KEY (employee_id) REFERENCES employee(id) ON DELETE CASCADE,
			UNIQUE (employee_id, day_of_week)
		);

		CREATE TABLE IF NOT EXISTS schedule (
			id      INTEGER PRIMARY KEY AUTOINCREMENT,
			week_of TEXT NOT NULL UNIQUE,
			notes   TEXT
		);

		CREATE TABLE IF NOT EXISTS shift (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			schedule_id INTEGER NOT NULL,
			employee_id INTEGER NOT NULL,
			day_of_week INTEGER NOT NULL,
			start_time  TEXT NOT NULL,
			end_time    TEXT NOT NULL,
			FOREIGN KEY (schedule_id) REFERENCES schedule(id) ON DELETE CASCADE,
			FOREIGN KEY (employee_id) REFERENCES employee(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS store_hours (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			day_of_week INTEGER NOT NULL UNIQUE,
			open_time   TEXT NOT NULL,
			close_time  TEXT NOT NULL
		);
	`)
	return err
}
