package utils

import "golang.org/x/crypto/bcrypt"

// GeneratePassword func for generate password hash.
func GeneratePassword(password string) string {
	// Generate password hash.
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Return password hash as string.
	return string(hash)
}

// ComparePasswords func for compare password hash and password.
func ComparePasswords(hashedPassword, password string) bool {
	// Compare password hash and password.
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	// Return true, if password hash and password are match.
	return err == nil
}
