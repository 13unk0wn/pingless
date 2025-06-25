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
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Not the same signing method")
			}
			return []byte(secretKey), nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), "props", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
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
