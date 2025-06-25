package routes

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"pingless/routes/user"
	"strconv"
)

func Routes(db *sqlx.DB) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	port, err := GetPort(db)
	if err != nil {
		log.Fatal("Error in getting Port ", err)
	}
	r.Post("/api/user/email_verification", func(w http.ResponseWriter, r *http.Request) {
		user.Email(w, r, db)
	})
	r.Post("/api/user/otp_verification", func(w http.ResponseWriter, r *http.Request) {
		user.OtpVerify(w, r, db)
	})
	r.Post("/api/user/create_user", func(w http.ResponseWriter, r *http.Request) {
		user.CreateUser(w, r, db)
	})

	r.Post("/api/user/verify_user", func(w http.ResponseWriter, r *http.Request) {
		user.VerifyUser(w, r, db)
	})
	r.With(user.VerifiyAccessToken).Post("/api/user/upload_pfp", func(w http.ResponseWriter, r *http.Request) {
		user.UpdatePfp(w, r, db)
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
