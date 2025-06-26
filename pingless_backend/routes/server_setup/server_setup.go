package serversetup

import (
	"encoding/json"
	"log"
	"net/http"
	"pingless/routes/user"
	"strings"

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

	_, err := db.Exec("UPDATE server_settings SET name =  ? WHERE id = 1", server.ServerName)
	if err != nil {
		http.Error(w, "DB ERROR", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
