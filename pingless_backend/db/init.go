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

	if err := insertInitialRolesAndPermissions(db); err != nil {
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
	if err := createPermissionTable(db); err != nil {
		return err
	}
	if err := createRoleTable(db); err != nil {
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

	role_id INTEGER NOT NULL DEFAULT 2 REFERENCES roles(id) ON DELETE SET NULL,

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
		id INTEGER PRIMARY KEY CHECK (id = 1), 
		name TEXT NOT NULL DEFAULT "",

		icon_url TEXT,
		banner_url TEXT,

		allow_gif_pfp BOOLEAN NOT NULL DEFAULT TRUE,
		invite_only BOOLEAN NOT NULL DEFAULT TRUE,

		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Insert default row with id = 1 if not exists
	insert := `
	INSERT INTO server_settings (id, name)
	SELECT 1, ''
	WHERE NOT EXISTS (
		SELECT 1 FROM server_settings WHERE id = 1
	);`
	_, err := db.Exec(insert)
	return err
}

func createPermissionTable(db *sqlx.DB) error {
	schema := `CREATE TABLE IF NOT EXISTS permissions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	can_server_setting BOOLEAN NOT NULL DEFAULT FALSE
);
`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	return nil
}
func createRoleTable(db *sqlx.DB) error {
	schema := `CREATE TABLE IF NOT EXISTS roles (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	description TEXT DEFAULT '',
	permission_id INTEGER NOT NULL UNIQUE,
	FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);
`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	return nil
}
func insertInitialRolesAndPermissions(db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// Skip if roles already exist
	var count int
	if err := tx.Get(&count, `SELECT COUNT(*) FROM roles`); err != nil {
		tx.Rollback()
		return err
	}
	if count > 0 {
		return tx.Commit()
	}

	// Insert permissions
	_, err = tx.Exec(`INSERT INTO permissions (can_server_setting) VALUES (TRUE);`)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`INSERT INTO permissions DEFAULT VALUES;`)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert roles (permission_id 1 = Owner, 2 = Member)
	_, err = tx.Exec(`
		INSERT INTO roles (name, description, permission_id) VALUES
		('Owner', 'Full access to everything', 1),
		('Member', 'Default user with basic access', 2);
	`)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
