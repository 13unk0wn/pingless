package fileutil

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"pingless/internal/auditlog"
	"time"

	"github.com/chai2010/webp"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

type FileUploadConfig struct {
	formFieldName    string
	maxFileSize      int64
	allowedMimeTypes map[string]bool
	uploadSubDir     string
	dbColumnName     string
	fileExtension    string
	isWebp           bool
}

func NewFileUploadConfig(
	formFieldName string,
	maxFileSize int64,
	allowedMimeTypes map[string]bool,
	uploadSubDir string,
	dbColumnName string,
	fileExtension string,
	isWebp bool,
) *FileUploadConfig {
	return &FileUploadConfig{
		formFieldName:    formFieldName,
		maxFileSize:      maxFileSize,
		allowedMimeTypes: allowedMimeTypes,
		uploadSubDir:     uploadSubDir,
		dbColumnName:     dbColumnName,
		fileExtension:    fileExtension,
		isWebp:           isWebp,
	}
}

func generateUniqueFileName(imageType, extension string) string {
	timestamp := time.Now().Unix()
	randomStr := fmt.Sprintf("%d", timestamp)[:8]
	return fmt.Sprintf("%s_%d_%s%s", imageType, timestamp, randomStr, extension)
}
func generateFileHash(file multipart.File) (string, error) {
	hash := md5.New()
	file.Seek(0, io.SeekStart)
	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	file.Seek(0, io.SeekStart)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func CheckMimeType(file multipart.File, allowed map[string]bool) error {
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

func SaveRawFile(src multipart.File, dstPath string) error {
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

func ConvertToWebP(src io.Reader, dstPath string) error {
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
func HandleFileUpload(w http.ResponseWriter, r *http.Request, db *sqlx.DB, config *FileUploadConfig) {
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

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Println("User not found:", err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Get the file
	file, header, err := r.FormFile(config.formFieldName)
	if err != nil {
		http.Error(w, "Cannot read uploaded image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check MIME type
	if err := CheckMimeType(file, config.allowedMimeTypes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileName := generateUniqueFileName(config.uploadSubDir, config.fileExtension)
	fileHash, err := generateFileHash(file)
	if err != nil {
		http.Error(w, "Cannot hash the value", http.StatusInternalServerError)
		return
	}

	// Rewind file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Failed to rewind file", http.StatusInternalServerError)
		return
	}

	// Save file with generated filename
	dstPath := filepath.Join("uploads", config.uploadSubDir, fileName)
	if config.isWebp {
		err = ConvertToWebP(file, dstPath)
	} else {
		err = SaveRawFile(file, dstPath)
	}
	if err != nil {
		log.Println("File saving failed:", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	fileInfo, err := os.Stat(dstPath)
	if err != nil {
		log.Println("Failed to get file info:", err)
		http.Error(w, "Failed to process file", http.StatusInternalServerError)
		return
	}

	// Store image metadata in database
	query := `
		INSERT INTO images (user_id, image_type, file_name, file_size, mime_type, hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, image_type) DO UPDATE SET
			file_name = excluded.file_name,
			file_size = excluded.file_size,
			mime_type = excluded.mime_type,
			hash = excluded.hash,
			updated_at = excluded.updated_at
	`

	now := time.Now()
	_, err = db.Exec(query,
		userID,
		config.uploadSubDir, // image_type (pfp or banner)
		fileName,
		fileInfo.Size(),
		header.Header.Get("Content-Type"),
		fileHash,
		now,
		now,
	)
	if err != nil {
		log.Println("Failed to store image metadata:", err)
		http.Error(w, "Failed to store image metadata", http.StatusInternalServerError)
		return
	}

	// Return JSON response with image URL
	imageURL := fmt.Sprintf("/images/%s", fileName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "File uploaded successfully",
		"image_url": imageURL,
		"file_name": fileName,
	})
}

func ServerFileUpload(w http.ResponseWriter, r *http.Request, db *sqlx.DB, config *FileUploadConfig) {
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
	if err := CheckMimeType(file, config.allowedMimeTypes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Rewind file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Failed to rewind file", http.StatusInternalServerError)
		return
	}

	// Save file
	dstPath := filepath.Join("uploads", "server", config.uploadSubDir, "server"+config.fileExtension)
	if config.isWebp {
		err = ConvertToWebP(file, dstPath)
	} else {
		err = SaveRawFile(file, dstPath)
	}
	if err != nil {
		log.Println("File saving failed:", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Update DB with new path
	query := fmt.Sprintf("UPDATE server_settings SET %s = ? WHERE id = 1", config.dbColumnName)
	_, err = db.Exec(query, dstPath)
	if err != nil {
		log.Println("DB update error:", err)
		http.Error(w, "Database update failed", http.StatusInternalServerError)
		return
	}

	auditlog.Record(db, auditlog.AuditLog{
		UserName: username,
		Action:   fmt.Sprintf("change_server_%s", config.dbColumnName),
		Target:   config.dbColumnName,
		Metadata: map[string]string{
			//TODO When add image link can be done when we serve image
		},
	})
	// Success
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("File uploaded successfully\n"))
}
