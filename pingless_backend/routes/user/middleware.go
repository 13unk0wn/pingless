package user

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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
		}
	})
}
