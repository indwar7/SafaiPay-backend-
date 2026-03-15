package auth

type OTPRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
}

type OTPVerifyRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	OTP         string `json:"otp" validate:"required,len=6"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}
