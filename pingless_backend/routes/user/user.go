package user

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func CreateUser(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	var user CreateUserModel

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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		http.Error(w, "Hash Error", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)
	if err := insertUser(db, &user); err != nil {
		log.Println(err)
		http.Error(w, "DB ERROR", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User Created\n"))
}

func insertUser(db *sqlx.DB, user *CreateUserModel) error {
	_, err := db.Exec("INSERT INTO users (email,username,password_hash) VALUES (?,?,?)", user.Email, user.Username, user.Password)
	return err
}
