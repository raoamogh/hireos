package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory      uint32 = 64 * 1024
	argon2Iterations  uint32 = 3
	argon2Parallelism uint8  = 4
	argon2SaltLength         = 16
	argon2KeyLength   uint32 = 32
)

var ErrInvalidPasswordHash = errors.New("invalid password hash")

func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argon2Iterations,
		argon2Memory,
		argon2Parallelism,
		argon2KeyLength,
	)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Iterations,
		argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")

	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return false, ErrInvalidPasswordHash
	}

	var memory, iterations uint32
	var parallelism uint8

	n, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&memory,
		&iterations,
		&parallelism,
	)

	if err != nil || n != 3 {
		return false, ErrInvalidPasswordHash
	}

	if memory != argon2Memory ||
		iterations != argon2Iterations ||
		parallelism != argon2Parallelism ||
		parts[3] != fmt.Sprintf(
			"m=%d,t=%d,p=%d",
			memory,
			iterations,
			parallelism,
		) {
		return false, ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != argon2SaltLength {
		return false, ErrInvalidPasswordHash
	}

	expectedHash, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(expectedHash) != int(argon2KeyLength) {
		return false, ErrInvalidPasswordHash
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		argon2KeyLength,
	)

	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1, nil
}
