package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGenerateActionToken(t *testing.T) {
	rawToken, tokenHash, err := GenerateActionToken()
	if err != nil {
		t.Fatalf("GenerateActionToken() error = %v", err)
	}

	if len(rawToken) != base64.RawURLEncoding.EncodedLen(actionTokenEntropyBytes) {
		t.Fatalf("unexpected token length: got %d", len(rawToken))
	}

	entropy, err := base64.RawURLEncoding.DecodeString(rawToken)
	if err != nil {
		t.Fatalf("generated token is not valid base64url: %v", err)
	}

	if len(entropy) != actionTokenEntropyBytes {
		t.Fatalf("unexpected entropy length: got %d", len(entropy))
	}

	expectedHash := sha256.Sum256(entropy)
	if tokenHash != expectedHash {
		t.Fatal("generated token hash does not match token entropy")
	}
}

func TestGenerateActionTokenProducesDifferentTokens(t *testing.T) {
	first, _, err := GenerateActionToken()
	if err != nil {
		t.Fatalf("first GenerateActionToken() error = %v", err)
	}

	second, _, err := GenerateActionToken()
	if err != nil {
		t.Fatalf("second GenerateActionToken() error = %v", err)
	}

	if first == second {
		t.Fatal("expected independently generated tokens to differ")
	}
}

func TestHashActionToken(t *testing.T) {
	rawToken, expectedHash, err := GenerateActionToken()
	if err != nil {
		t.Fatalf("GenerateActionToken() error = %v", err)
	}

	actualHash, err := HashActionToken(rawToken)
	if err != nil {
		t.Fatalf("HashActionToken() error = %v", err)
	}

	if actualHash != expectedHash {
		t.Fatal("HashActionToken() returned an unexpected hash")
	}
}

func TestHashActionTokenRejectsMalformedTokens(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{name: "empty token", token: ""},
		{name: "invalid base64", token: "not a valid token!"},
		{name: "too short", token: "abc"},
		{name: "wrong entropy length", token: base64.RawURLEncoding.EncodeToString([]byte("short"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HashActionToken(tt.token)
			if err == nil {
				t.Fatal("expected malformed token to be rejected")
			}
		})
	}
}
