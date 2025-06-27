package user

import (
	"encoding/json"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"pingless/internal/fileutil"
)

/*
NOTE : This file is contains endpoint for All the Profile Options

TODO:
[ ] Delete Profile picture if it already exist
*/
const MAX_BIO_SIZE int = 200

func UpdatePfp(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileutil.NewFileUploadConfig(
		"pfp",
		5<<20, // 5MB
		map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true},
		"pfp",
		"pfp",
		".webp",
		true,
	)
	fileutil.HandleFileUpload(w, r, db, config)
}

func UpdatePfpGif(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileutil.NewFileUploadConfig(
		"pfp",
		8<<20, // 8MB
		map[string]bool{"image/gif": true},
		"pfp",
		"pfp",
		".gif",
		false,
	)
	fileutil.HandleFileUpload(w, r, db, config)
}

func UpdateBanner(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileutil.NewFileUploadConfig(
		"banner",
		5<<20, // 5MB
		map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true},
		"banner",
		"banner",
		".webp",
		true,
	)
	fileutil.HandleFileUpload(w, r, db, config)
}

func UpdateBannerGif(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileutil.NewFileUploadConfig(
		"banner",
		8<<20, // 8MB
		map[string]bool{"image/gif": true},
		"banner",
		"banner",
		".gif",
		false,
	)
	fileutil.HandleFileUpload(w, r, db, config)
}

func UpdateBio(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	// Checking claims
	claims, ok := r.Context().Value("props").(jwt.MapClaims)
	if !ok {
		log.Println("Invalid token claims context")
		http.Error(w, "Invalid token claims", http.StatusInternalServerError)
		return
	}

	// Converting to string
	username, err := claims["username"].(string)
	if !err {
		log.Println("username claim not a string")
		http.Error(w, "Invalid token payload", http.StatusUnauthorized)
		return
	}

	// json decoding
	var bio UpdateBioModel
	if err := json.NewDecoder(r.Body).Decode(&bio); err != nil {
		http.Error(w, "Invalid Json", http.StatusBadRequest)
		return
	}

	bio.Bio = strings.Trim(bio.Bio, " ") // trim whitespace
	if len(bio.Bio) > MAX_BIO_SIZE {
		http.Error(w, "Max length exceded", http.StatusBadRequest)
		return
	}
	// UPDATE in DB
	_, success := db.Exec("UPDATE users SET bio = ? WHERE username = ?", bio.Bio, username)
	if success != nil {
		http.Error(w, "DB ERROR", http.StatusInternalServerError)
		return
	}

	//Response
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Bio Updated\n"))
}
