package utils

import (
	"time"

	"github.com/o1egl/paseto"
)

// Sign PASETO token
func SignPasetoToken(secretKey string, userID string, email string, expiry time.Duration) (string, error) {
	now := time.Now()
	json := map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"exp":     now.Add(expiry).Unix(),
		"iat":     now.Unix(),
	}
	token, err := paseto.NewV2().Encrypt([]byte(secretKey), json, nil)
	if err != nil {
		return "", err
	}
	return token, nil
}
