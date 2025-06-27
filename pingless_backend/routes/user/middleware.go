package user

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

func VerifiyAccessToken(next http.Handler) http.Handler {
	secretKey := os.Getenv("SECRETKEY")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			log.Println("Malformed token")
			http.Error(w, "Malformed Token", http.StatusUnauthorized)
			return
		}

		jwtToken := authHeader[1]
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Unexpected signing method")
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			log.Println("JWT parse error:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Println("Invalid token or claims")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "props", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IsGifAllowed(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var gifAllowed string
			err := db.QueryRow("SELECT value FROM settings WHERE key = 'GifAllowed'").Scan(&gifAllowed)
			if err != nil {
				http.Error(w, "DB ERROR", http.StatusUnauthorized)
				return
			}
			log.Println(gifAllowed)
			if gifAllowed != "true" {
				http.Error(w, "Server Dont allow GIF pfp/header", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
func IsInviteOnly(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var inviteOnly string
			err := db.QueryRow("SELECT value FROM settings WHERE key = 'inviteOnly'").Scan(&inviteOnly)
			if err != nil {
				http.Error(w, "DB ERROR", http.StatusUnauthorized)
				return
			}
			if inviteOnly == "true" {
				http.Error(w, "Server is invite only", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
