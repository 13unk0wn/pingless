package fileutil

import (
	"fmt"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"pingless/internal/auditlog"

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
	dstPath := filepath.Join("uploads", config.uploadSubDir, username+config.fileExtension)
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
