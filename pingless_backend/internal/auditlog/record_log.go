package auditlog

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
)

type AuditLog struct {
	UserName string
	Action   string
	Target   string
	Metadata map[string]string
}

func Record(db *sqlx.DB, log AuditLog) error {
	meta, _ := json.Marshal(log.Metadata)
	_, err := db.Exec(`
		INSERT INTO audit_log (user_name, action, target, metadata)
		VALUES (?, ?, ?, ?)`,
		log.UserName, log.Action, log.Target, string(meta),
	)
	return err
}
