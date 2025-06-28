package user

import "time"

type EmailVerification struct {
	Email     string    `db:"email"`
	OtpHash   string    `db:"otp_hash"`
	Verified  bool      `db:"verified"`
	CreatedAt time.Time `db:"created_at"`
}

type EmailSend struct {
	Email string `json:"email"`
}

type OtpVerifyModel struct {
	Email string `json:"email"`
	Otp   string `json:"otp"`
}

type CreateUserModel struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
	Username string `json:"username" db:"username"`
}

type VerifyUserModel struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type UpdateBioModel struct {
	Bio string `json:"bio" db:"bio"`
}

type changePasswordModel struct {
	Password    string `json:"password" db:"password"`
	NewPassword string `json : "new_password"`
}
