package serversetup

/*
ADD LOG SUPPORT
*/

import (
	"encoding/json"
	"log"
	"net/http"
	"pingless/internal/auditlog"
	"pingless/internal/fileutil"
	"pingless/routes/user"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

const SERVER_NAME_LENGTH int = 200

func CreateOwner(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	var user user.CreateUserModel

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if user is verified
	var verified bool
	err := db.QueryRow("SELECT verified FROM email_verifications WHERE email = ?", user.Email).Scan(&verified)
	if err != nil {
		log.Println(err)
		http.Error(w, "DB ERROR", http.StatusInternalServerError)
		return
	}
	if !verified {
		log.Println(err)
		http.Error(w, "UNAUTHORIZED", http.StatusBadRequest)
		return
	}

	var exists bool
	err = db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM users WHERE role_id = 1)`)
	if err != nil {
		log.Println(err)
		http.Error(w, "DB ERROR", http.StatusInternalServerError)
		return
	}
	if exists {
		log.Println(err)
		http.Error(w, "Owner already present", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		http.Error(w, "Hash Error", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)
	if err := createOwnerQuery(db, &user); err != nil {
		log.Println(err)
		http.Error(w, "DB ERROR", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Owner Created\n"))
}

func createOwnerQuery(db *sqlx.DB, user *user.CreateUserModel) error {
	_, err := db.Exec(`
		INSERT INTO users (username, email, password_hash, role_id)
		VALUES (?, ?, ?, 1)
	`, user.Username, user.Email, user.Password)
	return err
}

func SetServerName(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	claims, ok := r.Context().Value("props").(jwt.MapClaims)
	if !ok {
		log.Println("Invalid token claims context")
		http.Error(w, "Invalid token claims", http.StatusInternalServerError)
		return
	}
	username, err := claims["username"].(string)
	if !err {
		log.Println("username claim not a string")
		http.Error(w, "Invalid token payload", http.StatusUnauthorized)
		return
	}
	var server SetServerNameStruct

	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	server.ServerName = strings.Trim(server.ServerName, " ")
	log.Println(server.ServerName)

	len := len(server.ServerName)
	if len > SERVER_NAME_LENGTH || len == 0 {
		log.Println("Server name length more than 200")
		http.Error(w, "Allowed length 1 ≤ len ≤ 200", http.StatusBadRequest)
		return
	}

	row := db.QueryRowx(`
	UPDATE server_settings 
	SET name = ? 
	WHERE id = 1 
	RETURNING name
`, server.ServerName)

	var oldName string
	if err := row.Scan(&oldName); err != nil {
		log.Println(err)
		http.Error(w, "DB ERROR", http.StatusInternalServerError)
		return
	}
	auditlog.Record(db, auditlog.AuditLog{
		UserName: username,
		Action:   "change_server_name",
		Target:   "server_name",
		Metadata: map[string]string{
			"old": oldName,
			"new": server.ServerName,
		},
	})
	w.WriteHeader(http.StatusAccepted)
}

func SetServerProfile(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileutil.NewFileUploadConfig(
		"pfp",
		5<<20, // 5MB
		map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true},
		"pfp",
		"pfp",
		".webp",
		true,
	)

	fileutil.ServerFileUpload(w, r, db, config)
}

func SetServerProfileGif(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileutil.NewFileUploadConfig(
		"pfp",
		8<<20, // 8MB
		map[string]bool{"image/gif": true},
		"pfp",
		"pfp",
		".gif",
		false,
	)

	fileutil.ServerFileUpload(w, r, db, config)
}

func SetServerBanner(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileutil.NewFileUploadConfig(
		"banner",
		5<<20, // 5MB
		map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true},
		"banner",
		"header",
		".webp",
		true,
	)

	fileutil.ServerFileUpload(w, r, db, config)
}
func SetServerBannerGif(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileutil.NewFileUploadConfig(
		"banner",
		8<<20, // 8MB
		map[string]bool{"image/gif": true},
		"banner",
		"header",
		".gif",
		false,
	)
	fileutil.ServerFileUpload(w, r, db, config)
}
