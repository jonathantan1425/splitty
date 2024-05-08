package main

import (
	"database/sql"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
)

type OwingHistory struct {
	ID            int
	ChatID        int
	OwingUsername string
	OwedUsername  string
	Amount        float64
}

func GetDB() (*sql.DB, error) {
	once.Do(func() {
		slog.Info("Opening database")
		db, dbError = sql.Open("sqlite3", "./data/sqlite.db")
		if dbError != nil {
			slog.Error("Error opening database:", dbError)
		}
		if db == nil {
			slog.Error("db nil")
		}
		sql := `
		CREATE TABLE IF NOT EXISTS owing_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chat_id INTEGER,
			debtor_id TEXT,
			creditor_id TEXT,
			amount REAL
		);
		
		CREATE INDEX IF NOT EXISTS chat_id_idx ON owing_history(chat_id, debtor_id, creditor_id);

		CREATE TABLE IF NOT EXISTS chat_user_identities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chat_id INTEGER,
			username TEXT,
			user_id INTEGER
		);

		CREATE INDEX IF NOT EXISTS chat_id_idx ON chat_user_identities(chat_id, username, user_id);
		`

		_, err := db.Exec(sql)
		if err != nil {
			slog.Error("Error creating table:", err)
		}
		slog.Info("Database init complete")
	})

	return db, nil
}
