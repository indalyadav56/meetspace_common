package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtService interface {
	GenerateToken(userID string, duration time.Duration) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	GetClaims(token *jwt.Token) (jwt.MapClaims, error)
}

// JWTService implements TokenService
type JWTService struct {
	secretKey []byte
}

// NewJWTService creates a new JWTService
func NewJWTService(secretKey string) JwtService {
	return &JWTService{secretKey: []byte(secretKey)}
}

// GenerateToken generates a new JWT token
func (j *JWTService) GenerateToken(userID string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates the given token
func (j *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

// GetClaims retrieves the claims from a validated token as a map
func (j *JWTService) GetClaims(token *jwt.Token) (jwt.MapClaims, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token or claims")
	}
	return claims, nil
}
