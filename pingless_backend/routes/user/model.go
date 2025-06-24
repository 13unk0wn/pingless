package user

import "time"

type EmailVerification struct {
	Email     string    `db:"email"`
	OtpHash   string    `db:"otp_hash"`
	Verified  bool      `db:"verified"`
	CreatedAt time.Time `db:"created_at"`
}

type EmailSend struct {
	Email string `db:"email"`
}
