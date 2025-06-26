package serversetup

import (
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

func CanchangeServerSettings(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var canChange bool

			claims, ok := r.Context().Value("props").(jwt.MapClaims)
			if !ok {
				log.Println("Invalid token claims context")
				http.Error(w, "Invalid token claims", http.StatusInternalServerError)
				return
			}
			err := db.Get(&canChange, `
	        SELECT p.can_server_setting
	        FROM users u
	        JOIN roles r ON u.role_id = r.id
	        JOIN permissions p ON r.permission_id = p.id
	        WHERE u.username = ?
           `, claims["username"])
			if err != nil {
				log.Println(err)
				http.Error(w, "DB ERROR", http.StatusInternalServerError)
				return
			}
			if !canChange {
				log.Println(err)
				http.Error(w, "UNAUTHORIZED", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
