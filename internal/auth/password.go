package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// Used during registration
// HashPassword converts a plaintext password into a bcrypt hash
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hashed), err
}

// Used during login
// ComparePassword verifies a plaintext password against a hash
func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}
