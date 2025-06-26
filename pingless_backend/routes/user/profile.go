package user

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

/*
NOTE : This file is contains endpoint for All the Profile Options

TODO:
[ ] Delete Profile picture if it already exist
[ ] Implement Bio
*/
const MAX_BIO_SIZE int = 200

type fileUploadConfig struct {
	formFieldName    string
	maxFileSize      int64
	allowedMimeTypes map[string]bool
	uploadSubDir     string
	dbColumnName     string
	fileExtension    string
	isWebp           bool
}

func UpdatePfp(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileUploadConfig{
		formFieldName:    "pfp",
		maxFileSize:      5 << 20, // 5MB
		allowedMimeTypes: map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true},
		uploadSubDir:     "pfp",
		dbColumnName:     "pfp",
		fileExtension:    ".webp",
		isWebp:           true,
	}
	handleFileUpload(w, r, db, config)
}

func UpdatePfpGif(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileUploadConfig{
		formFieldName:    "pfp",
		maxFileSize:      8 << 20, // 8MB
		allowedMimeTypes: map[string]bool{"image/gif": true},
		uploadSubDir:     "pfp",
		dbColumnName:     "pfp",
		fileExtension:    ".gif",
		isWebp:           false,
	}
	handleFileUpload(w, r, db, config)
}

func UpdateBanner(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileUploadConfig{
		formFieldName:    "banner",
		maxFileSize:      5 << 20, // 5MB
		allowedMimeTypes: map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true},
		uploadSubDir:     "banner",
		dbColumnName:     "banner",
		fileExtension:    ".webp",
		isWebp:           true,
	}
	handleFileUpload(w, r, db, config)
}

func UpdateBannerGif(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	config := fileUploadConfig{
		formFieldName:    "banner",
		maxFileSize:      8 << 20, // 8MB
		allowedMimeTypes: map[string]bool{"image/gif": true},
		uploadSubDir:     "banner",
		dbColumnName:     "banner",
		fileExtension:    ".gif",
		isWebp:           false,
	}
	handleFileUpload(w, r, db, config)
}

func handleFileUpload(w http.ResponseWriter, r *http.Request, db *sqlx.DB, config fileUploadConfig) {
	// Extract JWT claims from context
	claims, ok := r.Context().Value("props").(jwt.MapClaims)
	if !ok {
		log.Println("Invalid token claims context")
		http.Error(w, "Invalid token claims", http.StatusInternalServerError)
		return
	}
	username, ok := claims["username"].(string)
	if !ok {
		log.Println("username claim not a string")
		http.Error(w, "Invalid token payload", http.StatusUnauthorized)
		return
	}

	// Limit request body size before parsing
	r.Body = http.MaxBytesReader(w, r.Body, config.maxFileSize)
	err := r.ParseMultipartForm(config.maxFileSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("File size exceeds %dMB", config.maxFileSize>>20), http.StatusRequestEntityTooLarge)
		return
	}

	// Get the file
	file, _, err := r.FormFile(config.formFieldName)
	if err != nil {
		http.Error(w, "Cannot read uploaded image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check MIME type
	if err := checkMimeType(file, config.allowedMimeTypes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Rewind file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Failed to rewind file", http.StatusInternalServerError)
		return
	}

	// Save file
	dstPath := filepath.Join("uploads", config.uploadSubDir, username+config.fileExtension)
	if config.isWebp {
		err = convertToWebP(file, dstPath)
	} else {
		err = saveRawFile(file, dstPath)
	}
	if err != nil {
		log.Println("File saving failed:", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Update DB with new path
	query := fmt.Sprintf("UPDATE users SET %s = ? WHERE username = ?", config.dbColumnName)
	_, err = db.Exec(query, dstPath, username)
	if err != nil {
		log.Println("DB update error:", err)
		http.Error(w, "Database update failed", http.StatusInternalServerError)
		return
	}

	// Success
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("File uploaded successfully\n"))
}

func checkMimeType(file multipart.File, allowed map[string]bool) error {
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read image")
	}
	contentType := http.DetectContentType(buf)
	if !allowed[contentType] {
		return fmt.Errorf("file type not allowed")
	}
	return nil
}

func saveRawFile(src multipart.File, dstPath string) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func convertToWebP(src io.Reader, dstPath string) error {
	// Decode the image (JPEG, PNG, etc.)
	img, _, err := image.Decode(src)
	if err != nil {
		return err
	}

	// Ensure the destination directory exists
	err = os.MkdirAll(filepath.Dir(dstPath), 0755)
	if err != nil {
		return err
	}

	// Create the destination file
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Encode to WebP
	op := &webp.Options{Lossless: true}
	return webp.Encode(out, img, op)
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
