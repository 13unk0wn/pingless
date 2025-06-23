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
	return makeUserMigration(db)
}

func makeUserMigration(db *sqlx.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS user (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL,                             -- display name
	email VARCHAR(319) NOT NULL UNIQUE,
	password TEXT NOT NULL,
	otp TEXT NOT NULL,
	verified BOOLEAN NOT NULL DEFAULT FALSE,

	pfp TEXT,                                           -- profile image path or URL
	pfp_type TEXT CHECK(pfp_type IN ('image', 'gif')) DEFAULT 'image',

	header TEXT,                                        -- header image path or URL
	header_type TEXT CHECK(header_type IN ('image', 'gif')) DEFAULT 'image',

	bio TEXT DEFAULT '',                                -- short user description
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
	_, err := db.Exec(schema)
	return err
}
