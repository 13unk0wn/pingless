package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func Init() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", "./pingless.db")
	if err != nil {
		return nil, err
	}

	// Force connection and file creation
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Run migrations
	if err := makeMigration(db); err != nil {
		return nil, err
	}

	return db, nil
}

func makeMigration(db *sqlx.DB) error {
	if err := makeUserMigration(db); err != nil {
		return err
	}
	if err := createEmailVerificationTable(db); err != nil {
		return err
	}
	if err := createSettingsTable(db); err != nil {
		return err
	}
	if err := createServerSettingsTable(db); err != nil {
		return err
	}
	return nil
}

func makeUserMigration(db *sqlx.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	email VARCHAR(319) NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,

	pfp TEXT,                  -- avatar URL or path
	header TEXT,               -- banner URL or path
	bio TEXT DEFAULT '',

	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

	_, err := db.Exec(schema)
	return err
}

func createEmailVerificationTable(db *sqlx.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS email_verifications (
	email VARCHAR(319) PRIMARY KEY,
	otp_hash TEXT NOT NULL,
	verified BOOLEAN NOT NULL DEFAULT FALSE,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`
	_, err := db.Exec(schema)
	return err
}

func createSettingsTable(db *sqlx.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);`
	_, err := db.Exec(schema)
	return err
}

func createServerSettingsTable(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS server_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),  -- Singleton enforcement
    name TEXT NOT NULL CHECK(length(name) BETWEEN 1 AND 100),

    icon_url TEXT,        -- Server icon path or URL
    banner_url TEXT,      -- Optional banner image

    allow_gif_pfp BOOLEAN NOT NULL DEFAULT FALSE,
    invite_only BOOLEAN NOT NULL DEFAULT TRUE,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
	_, err := db.Exec(schema)
	return err
}
