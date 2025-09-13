package dto

import (
	"time"
)

type CreateUserDto struct {
	FirstName     string `json:"first_name" validate:"required,min=2,max=50"`
	LastName      string `json:"last_name" validate:"required,min=2,max=50"`
	Email         string `json:"email" validate:"required,email"`
	PhoneNumber   string `json:"phone_number" validate:"required,e164"`
	PasswordHash  string `json:"password" validate:"required,min=8"`
	ProfilePicURL string `json:"profile_pic_url" validate:"omitempty,url"`
}

type LoginUserDto struct {
	Email    string `json:"email" validate:"required,min=2,max=50"`
	Password string `json:"password" validate:"required"`
}

type LoginUserResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	SessionID    string    `json:"session_id"`
	ExpiresIn    time.Time `json:"expires_in"`
}
