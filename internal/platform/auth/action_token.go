package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

const actionTokenEntropyBytes = 32

func GenerateActionToken() (string, [sha256.Size]byte, error) {
	var entropy [actionTokenEntropyBytes]byte

	if _, err := io.ReadFull(rand.Reader, entropy[:]); err != nil {
		return "", [sha256.Size]byte{}, fmt.Errorf(
			"generate action token entropy: %w",
			err,
		)
	}

	rawToken := base64.RawURLEncoding.EncodeToString(entropy[:])
	tokenHash := sha256.Sum256(entropy[:])

	return rawToken, tokenHash, nil
}

func HashActionToken(rawToken string) ([sha256.Size]byte, error) {
	entropy, err := base64.RawURLEncoding.Strict().DecodeString(rawToken)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf(
			"decode action token: %w",
			err,
		)
	}

	if len(entropy) != actionTokenEntropyBytes {
		return [sha256.Size]byte{}, fmt.Errorf(
			"invalid action token length: expected %d bytes, got %d",
			actionTokenEntropyBytes,
			len(entropy),
		)
	}

	return sha256.Sum256(entropy), nil
}
