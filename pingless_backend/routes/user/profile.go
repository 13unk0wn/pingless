package user

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/chai2010/webp"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

/*
NOTE : This file is contains endpoint for All the Profile Options


TODO:
[ ] Delete Profile picture if it already exist
*/

func UpdatePfp(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
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

	// Parse uploaded form (max 5MB)
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		log.Println(err)
		http.Error(w, "File size exceeds 2MB", http.StatusBadRequest)
		return
	}

	// Read image file
	file, _, err := r.FormFile("pfp")
	if err != nil {
		log.Println(err)
		http.Error(w, "Cannot read uploaded image", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Detect image content type
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(buf)

	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowed[contentType] {
		http.Error(w, "Only PNG, JPEG, and WebP images are allowed", http.StatusBadRequest)
		return
	}

	// Reset file pointer before decoding
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to rewind file", http.StatusInternalServerError)
		return
	}

	// Convert to WebP and save
	pfpPath := filepath.Join("uploads", "pfp", username+".webp")
	err = convertToWebP(file, pfpPath)
	if err != nil {
		log.Println("Image conversion failed:", err)
		http.Error(w, "Failed to convert image to WebP", http.StatusInternalServerError)
		return
	}

	// Update DB with new path
	_, err = db.Exec("UPDATE users SET pfp = ? WHERE username = ?", pfpPath, username)
	if err != nil {
		log.Println("DB update error:", err)
		http.Error(w, "Database update failed", http.StatusInternalServerError)
		return
	}

	// Success
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Image uploaded successfully\n"))
}

func UpdatePfpGif(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {

	// Extract JWT claims
	claims, ok := r.Context().Value("props").(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusInternalServerError)
		return
	}
	username, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "Invalid token payload", http.StatusUnauthorized)
		return
	}

	// Limit request body size to 8 MB before parsing
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20) // 8 MB
	err := r.ParseMultipartForm(8 << 20)
	if err != nil {
		http.Error(w, "File size exceeds 8MB", http.StatusRequestEntityTooLarge)
		return
	}

	// Get the file
	file, _, err := r.FormFile("pfp")
	if err != nil {
		http.Error(w, "Cannot read uploaded image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check MIME
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(buf)
	if contentType != "image/gif" {
		http.Error(w, "Not a valid GIF", http.StatusBadRequest)
		return
	}

	// Rewind
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, "Failed to rewind file", http.StatusInternalServerError)
		return
	}

	// Create destination path
	dstPath := filepath.Join("uploads", "pfp", username+".gif")
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	out, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Efficient copy without loading full buffer into memory
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}
	// Update DB with new path
	_, err = db.Exec("UPDATE users SET pfp = ? WHERE username = ?", dstPath, username)
	if err != nil {
		log.Println("DB update error:", err)
		http.Error(w, "Database update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("GIF uploaded successfully\n"))
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
	opts := &webp.Options{Lossless: true}
	return webp.Encode(out, img, opts)
}
