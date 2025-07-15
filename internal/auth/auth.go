package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil })
	if err != nil {
		return uuid.Nil, err
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(subject)
}

func GetBearerToken(headers http.Header) (string, error) {
	header, ok := headers["Authorization"]
	if !ok {
		return "", fmt.Errorf("not found")
	}

	token_str := ""
	for _, h := range header {
		bear := strings.Split(h, " ")
		if bear[0] == "Bearer" {
			token := bear[1:]
			token_str = strings.Join(token, " ")
		}
	}

	return token_str, nil
}

func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	str := hex.EncodeToString(b)
	return str, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	header, ok := headers["Authorization"]
	if !ok {
		return "", fmt.Errorf("not found")
	}

	apiKey_str := ""
	for _, h := range header {
		bear := strings.Split(h, " ")
		if bear[0] == "ApiKey" {
			key := bear[1:]
			apiKey_str = strings.Join(key, " ")
		}
	}

	return apiKey_str, nil
}
