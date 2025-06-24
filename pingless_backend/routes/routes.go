package routes

import (
	"fmt"
	"log"
	"net/http"
	"pingless/routes/user"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

func Routes(db *sqlx.DB) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	port, err := GetPort(db)
	if err != nil {
		log.Fatal("Error in getting Port ", err)
	}
	r.Post("/api/emailVerification", func(w http.ResponseWriter, r *http.Request) {
		user.Email(w, r, db)
	})

	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

func GetPort(db *sqlx.DB) (int, error) {
	var val string
	err := db.Get(&val, "SELECT value FROM settings WHERE key = 'port'")
	if err != nil {
		return 0, fmt.Errorf("failed to fetch port: %w", err)
	}

	port, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid port value: %s", val)
	}

	return port, nil
}
