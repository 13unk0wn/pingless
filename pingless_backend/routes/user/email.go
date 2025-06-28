package user

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/smtp"
	"time"

	"github.com/jmoiron/sqlx"
)

/*
NOTE : This file deal with all the email and otp verification

TODO:
 [ ] Add option for custom email template
*/

// TODO : dont send email if already Verified
func Email(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	var email EmailSend

	if err := json.NewDecoder(r.Body).Decode(&email); err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	verified, created_at, exist, error := checkEmailExist(db, email.Email)
	if verified {
		http.Error(w, "BAD REQUEST", http.StatusBadRequest)
		return
	}
	if error != nil {
		log.Println(error)
		http.Error(w, "DB ERROR", http.StatusInternalServerError)
		return
	}
	if exist {
		if !time.Now().After(created_at.Add(10 * time.Minute)) {
			http.Error(w, "Email is already send wait 10 minutes to send again", http.StatusBadRequest)
			return
		}

		otp := generateOtp()
		hashOtp := HashOTP(otp)
		// Update created_at and Otp
		if err := UpdateCreatedAt(db, email.Email, hashOtp); err != nil {
			log.Println(err)
			http.Error(w, "DB ERROR", http.StatusInternalServerError)
			return
		}
		if err := sendEmail(db, email.Email, hashOtp); err != nil {
			log.Println(err)
			http.Error(w, "Verification email Cannot Be Send", http.StatusInternalServerError)
			return
		}
	} else {
		var verify EmailVerification

		verify.Email = email.Email
		otp := generateOtp()
		verify.OtpHash = HashOTP(otp)
		verify.CreatedAt = time.Now()
		verify.Verified = false

		if err := insertEmailVerification(db, &verify); err != nil {
			log.Println(err)
			http.Error(w, "Database Error", http.StatusInternalServerError)
			return
		}
		if err := sendEmail(db, email.Email, otp); err != nil {
			log.Println(err)
			http.Error(w, "Verification email Cannot Be Send", http.StatusInternalServerError)
			return
		}

	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Email Send\n"))
}

func checkEmailExist(db *sqlx.DB, email string) (bool, time.Time, bool, error) {
	var createdAt time.Time
	var verified bool
	err := db.QueryRow("SELECT created_at, verified FROM email_verifications WHERE email = ? LIMIT 1", email).Scan(&createdAt, &verified)
	if verified {
		return true, createdAt, false, nil
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, createdAt, false, nil
		} else {
			return false, createdAt, false, err
		}
	} else {
		return false, createdAt, true, nil
	}
}

func UpdateCreatedAt(db *sqlx.DB, email string, hashOtp string) error {
	_, err := db.Exec(`
		UPDATE email_verifications
		SET created_at = ?,
		otp_hash = ?
		WHERE email = ?`,
		time.Now(), hashOtp, email,
	)
	return err
}

// TODO : Reject Request when already verified
func OtpVerify(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	var otp OtpVerifyModel

	if err := json.NewDecoder(r.Body).Decode(&otp); err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	var hashedOrignalOtp string
	var created_at time.Time
	var verified bool
	err := db.QueryRow("SELECT otp_hash,created_at,verified FROM email_verifications WHERE email = ? ", otp.Email).Scan(&hashedOrignalOtp, &created_at, &verified)
	if err != nil {
		log.Println(err)
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}
	if verified {
		http.Error(w, "BAD REQUEST", http.StatusBadRequest)
		return
	}
	hashedUserOtp := HashOTP(otp.Otp)
	if hashedOrignalOtp != hashedUserOtp || time.Now().After(created_at.Add(10*time.Minute)) {
		log.Println(err)
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE email_verifications SET verified = TRUE where email =  ?", otp.Email)
	if err != nil {
		log.Println(err)
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Verified\n"))
}

func HashOTP(otp string) string {
	sum := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(sum[:])
}

func generateOtp() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func insertEmailVerification(db *sqlx.DB, verify *EmailVerification) error {
	_, err := db.Exec("INSERT INTO email_verifications (email,otp_hash,verified,created_at) VALUES (?,?,?,?)", verify.Email, verify.OtpHash, verify.Verified, verify.CreatedAt)
	return err
}

func sendEmail(db *sqlx.DB, to string, otp string) error {
	var from string
	var password string
	var host string
	var port string

	err := db.QueryRow("SELECT value FROM settings WHERE key = 'email'").Scan(&from)
	if err != nil {
		return err
	}

	err = db.QueryRow("SELECT value FROM settings WHERE key = 'password'").Scan(&password)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT value FROM settings WHERE key = 'emailHost'").Scan(&host)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT value FROM settings WHERE key = 'emailPort'").Scan(&port)
	if err != nil {
		return err
	}

	subject := "Your Pingless Verification Code"

	msg := fmt.Sprintf(`Subject: %s
MIME-Version: 1.0
Content-Type: text/html; charset="UTF-8"

<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Pingless Email Verification</title>
</head>
<body style="font-family: Arial, sans-serif; background-color: #f9fafb; margin: 0; padding: 0;">
  <div style="background-color: #ffffff; max-width: 480px; margin: 40px auto; padding: 32px; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.05);">
    <div style="font-size: 20px; font-weight: 600; color: #111827; margin-bottom: 24px;">
      Verify your email with Pingless
    </div>
    <div style="font-size: 14px; color: #4b5563;">
      Use the code below to verify your email address. This code will expire in 10 minutes.
    </div>
    <div style="font-size: 32px; letter-spacing: 4px; font-weight: bold; color: #111827; background-color: #f3f4f6; padding: 16px; text-align: center; border-radius: 6px; margin: 24px 0;">
      %s
    </div>
    <div style="font-size: 14px; color: #4b5563;">
      If you did not request this code, you can safely ignore this email.
    </div>
    <div style="font-size: 12px; color: #9ca3af; text-align: center; margin-top: 32px;">
      Pingless · A self-hosted async status board<br>
      You received this email because someone tried to sign up with your address.
    </div>
  </div>
</body>
</html>
`, subject, otp)

	body := []byte(msg)

	auth := smtp.PlainAuth("", from, password, host)

	err = smtp.SendMail(host+":"+port, auth, from, []string{to}, body)
	return err
}
