package services

import (
	"errors"
	"time"

	"digital-memorial/api/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

type Claims struct {
	UserID uint   `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func IssueToken(userID uint, email, role string, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(config.Get().JWTSecret))
	return signed, expiresAt, err
}

func keyFunc(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}
	return []byte(config.Get().JWTSecret), nil
}

// ParseToken verifies the token's signature and expiry, returning an error if
// either check fails.
func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ParseTokenAllowExpired verifies the token's signature but tolerates an
// expired exp claim, returning wasExpired=true in that case. Used by the
// logout handler so it can still identify (and audit-log) who an
// idle-expired session belonged to.
func ParseTokenAllowExpired(tokenStr string) (claims *Claims, wasExpired bool, err error) {
	claims = &Claims{}
	token, parseErr := jwt.ParseWithClaims(tokenStr, claims, keyFunc)
	if parseErr == nil && token.Valid {
		return claims, false, nil
	}
	if errors.Is(parseErr, jwt.ErrTokenExpired) {
		return claims, true, nil
	}
	return nil, false, errors.New("invalid token")
}
