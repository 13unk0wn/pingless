package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type ImageResponse struct {
	ID        int    `json:"id"`
	ImageType string `json:"image_type"`
	URL       string `json:"url"`
	FileSize  int64  `json:"file_size"`
	MimeType  string `json:"mime_type"`
	UpdatedAt string `json:"updated_at"`
}

// GetUserImages returns all images for a specific user
func GetUserImages(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	var images []struct {
		ID        int    `db:"id"`
		ImageType string `db:"image_type"`
		FileName  string `db:"file_name"`
		FileSize  int64  `db:"file_size"`
		MimeType  string `db:"mime_type"`
		UpdatedAt string `db:"updated_at"`
	}

	query := `
		SELECT i.id, i.image_type, i.file_name, i.file_size, i.mime_type, i.updated_at
		FROM images i
		JOIN users u ON i.user_id = u.id
		WHERE u.username = ?
		ORDER BY i.image_type
	`

	err := db.Select(&images, query, username)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Convert to response format with URLs
	var response []ImageResponse
	for _, img := range images {
		response = append(response, ImageResponse{
			ID:        img.ID,
			ImageType: img.ImageType,
			URL:       "/images/" + img.FileName,
			FileSize:  img.FileSize,
			MimeType:  img.MimeType,
			UpdatedAt: img.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetImageInfo returns metadata for a specific image by ID
func GetImageInfo(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	imageID := r.URL.Query().Get("id")
	if imageID == "" {
		http.Error(w, "Image ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(imageID)
	if err != nil {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	var image struct {
		ID        int    `db:"id"`
		ImageType string `db:"image_type"`
		FileName  string `db:"file_name"`
		FileSize  int64  `db:"file_size"`
		MimeType  string `db:"mime_type"`
		UpdatedAt string `db:"updated_at"`
	}

	err = db.Get(&image, "SELECT id, image_type, file_name, file_size, mime_type, updated_at FROM images WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	response := ImageResponse{
		ID:        image.ID,
		ImageType: image.ImageType,
		URL:       "/images/" + image.FileName,
		FileSize:  image.FileSize,
		MimeType:  image.MimeType,
		UpdatedAt: image.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserImageByType returns a specific image type for a user
func GetUserImageByType(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	username := r.URL.Query().Get("username")
	imageType := r.URL.Query().Get("type")

	if username == "" || imageType == "" {
		http.Error(w, "Username and image type required", http.StatusBadRequest)
		return
	}

	if imageType != "pfp" && imageType != "banner" {
		http.Error(w, "Invalid image type. Must be 'pfp' or 'banner'", http.StatusBadRequest)
		return
	}

	var image struct {
		ID        int    `db:"id"`
		ImageType string `db:"image_type"`
		FileName  string `db:"file_name"`
		FileSize  int64  `db:"file_size"`
		MimeType  string `db:"mime_type"`
		UpdatedAt string `db:"updated_at"`
	}

	query := `
		SELECT i.id, i.image_type, i.file_name, i.file_size, i.mime_type, i.updated_at
		FROM images i
		JOIN users u ON i.user_id = u.id
		WHERE u.username = ? AND i.image_type = ?
	`

	err := db.Get(&image, query, username, imageType)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	response := ImageResponse{
		ID:        image.ID,
		ImageType: image.ImageType,
		URL:       "/images/" + image.FileName,
		FileSize:  image.FileSize,
		MimeType:  image.MimeType,
		UpdatedAt: image.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
