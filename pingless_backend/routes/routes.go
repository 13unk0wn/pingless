package routes

import (
	"fmt"
	"log"
	"net/http"
	serversetup "pingless/routes/server_setup"
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
	r.Post("/api/user/email_verification", func(w http.ResponseWriter, r *http.Request) {
		user.Email(w, r, db)
	})
	r.Post("/api/user/otp_verification", func(w http.ResponseWriter, r *http.Request) {
		user.OtpVerify(w, r, db)
	})
	r.Post("/api/user/verify_user", func(w http.ResponseWriter, r *http.Request) {
		user.VerifyUser(w, r, db)
	})
	r.Post("/api/server/create_owner", func(w http.ResponseWriter, r *http.Request) {
		serversetup.CreateOwner(w, r, db)
	})
	r.With(user.VerifiyAccessToken).Post("/api/user/upload_pfp", func(w http.ResponseWriter, r *http.Request) {
		user.UpdatePfp(w, r, db)
	})
	r.With(user.VerifiyAccessToken).With(user.IsGifAllowed(db)).Post("/api/user/upload_pfp_gif", func(w http.ResponseWriter, r *http.Request) {
		user.UpdatePfpGif(w, r, db)
	})
	r.With(user.VerifiyAccessToken).Post("/api/user/upload_banner", func(w http.ResponseWriter, r *http.Request) {
		user.UpdateBanner(w, r, db)
	})
	r.With(user.VerifiyAccessToken).With(user.IsGifAllowed(db)).Post("/api/user/upload_banner_gif", func(w http.ResponseWriter, r *http.Request) {
		user.UpdateBannerGif(w, r, db)
	})
	r.With(user.VerifiyAccessToken).Post("/api/user/upload_bio", func(w http.ResponseWriter, r *http.Request) {
		user.UpdateBio(w, r, db)
	})
	r.With(user.IsInviteOnly(db)).Post("/api/user/create_user", func(w http.ResponseWriter, r *http.Request) {
		user.CreateUser(w, r, db)
	})
	r.With(user.VerifiyAccessToken).With(serversetup.CanchangeServerSettings(db)).Post("/api/server/change_name", func(w http.ResponseWriter, r *http.Request) {
		serversetup.SetServerName(w, r, db)
	})
	r.With(user.VerifiyAccessToken).With(serversetup.CanchangeServerSettings(db)).Post("/api/server/change_profile", func(w http.ResponseWriter, r *http.Request) {
		serversetup.SetServerProfile(w, r, db)
	})
	r.With(user.VerifiyAccessToken).With(serversetup.CanchangeServerSettings(db)).Post("/api/server/change_profile_gif", func(w http.ResponseWriter, r *http.Request) {
		serversetup.SetServerProfileGif(w, r, db)
	})
	r.With(user.VerifiyAccessToken).With(serversetup.CanchangeServerSettings(db)).Post("/api/server/change_banner", func(w http.ResponseWriter, r *http.Request) {
		serversetup.SetServerBanner(w, r, db)
	})
	r.With(user.VerifiyAccessToken).With(serversetup.CanchangeServerSettings(db)).Post("/api/server/change_banner_gif", func(w http.ResponseWriter, r *http.Request) {
		serversetup.SetServerBannerGif(w, r, db)
	})
	r.Get("/api/user/images", func(w http.ResponseWriter, r *http.Request) {
		user.GetUserImages(w, r, db)
	})
	r.Get("/api/images/info", func(w http.ResponseWriter, r *http.Request) {
		user.GetImageInfo(w, r, db)
	})
	r.Get("/api/user/image", func(w http.ResponseWriter, r *http.Request) {
		user.GetUserImageByType(w, r, db)
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
