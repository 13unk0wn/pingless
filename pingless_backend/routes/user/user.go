package user

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
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
func createRefreshToken(username string) (string, error) {
	secretKey := os.Getenv("SECRETKEY")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add((time.Hour * 24) * 30).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	return tokenString, err
}

func createAccessToken(username string) (string, error) {
	secretKey := os.Getenv("SECRETKEY")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	return tokenString, err
}

func VerifyUser(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	var signin VerifyUserModel

	if err := json.NewDecoder(r.Body).Decode(&signin); err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var storedHash string

	// Fetch the stored password hash and verification status from the database
	// It's important to fetch the hash here, not generate a new one.
	if err := db.QueryRow("SELECT password_hash FROM users WHERE username = ?", signin.Username).Scan(&storedHash); err != nil {
		log.Printf("Database fetch error for user %s: %v", signin.Username, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Compare the incoming plaintext password with the stored hash
	// Use bcrypt.CompareHashAndPassword for this purpose
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(signin.Password)); err != nil {
		// If passwords don't match or there's an error during comparison (e.g., bad hash format)
		if err == bcrypt.ErrMismatchedHashAndPassword {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized) // Incorrect password
			return
		}
		log.Printf("Bcrypt comparison error for user %s: %v", signin.Username, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If we reach here, the password is correct and the user is verified.
	refreshToken, err := createRefreshToken(signin.Username)
	if err != nil {
		log.Println(err)
		http.Error(w, "JWT ERROR", http.StatusInternalServerError)
		return
	}

	// TODO : use secure in production
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
	})

	accessToken, err := createAccessToken(signin.Username)
	if err != nil {
		log.Println(err)
		http.Error(w, "JWT ERROR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})

}
