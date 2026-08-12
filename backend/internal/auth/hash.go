// Package auth implements password hashing/verification and JWT issuing.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const saltBytes = 16

// GenerateSalt returns a random hex-encoded salt for per-user password hashing.
func GenerateSalt() (string, error) {
	b := make([]byte, saltBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generate salt: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashPassword computes HMAC-SHA256(key=secret, message=salt+password), hex-encoded.
// See plan.md Key Decisions #1: the server-side secret keys the HMAC, the per-user
// salt keeps identical passwords from producing identical hashes across accounts.
func HashPassword(secret, salt, password string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(salt + password))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyPassword recomputes the HMAC and compares it against expectedHash with
// hmac.Equal for constant-time comparison — never == or bytes.Compare.
func VerifyPassword(secret, salt, password, expectedHash string) bool {
	computed := HashPassword(secret, salt, password)
	return hmac.Equal([]byte(computed), []byte(expectedHash))
}

// VerifyAdminPassword compares a submitted password against the configured admin
// password using the same constant-time HMAC scheme. The Admin identity has no
// database row (see spec.md Assumptions), so there's no stored hash to compare
// against — both sides are hashed at request time instead.
func VerifyAdminPassword(secret, submitted, configured string) bool {
	a := HashPassword(secret, "", submitted)
	b := HashPassword(secret, "", configured)
	return hmac.Equal([]byte(a), []byte(b))
}
