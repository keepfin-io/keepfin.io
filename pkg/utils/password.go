package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// GeneratePassword hashes a password using Argon2id.
func GeneratePassword(password string) string {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s:%s", saltB64, hashB64)
}

// ComparePasswords checks if the password matches the hash.
func ComparePasswords(hashedPassword, password string) bool {
	parts := strings.Split(hashedPassword, ":")
	if len(parts) != 2 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	newHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(newHash) == base64.RawStdEncoding.EncodeToString(hash)
}
