package utils

import (
	"github.com/golang-jwt/jwt/v5"
)

// Sign JWT token
func SignJwtToken(jwtSecret string, claims jwt.MapClaims) (string, error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}