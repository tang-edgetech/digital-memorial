package services

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken returns a hex-encoded random token of n bytes, used for
// the CSRF double-submit cookie.
func GenerateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
