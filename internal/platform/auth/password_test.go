package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestHashPasswordAndVerifyPassword(t *testing.T) {
	password := "Correct-Horse-Battery-Staple-2026!"

	encodedHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if !strings.HasPrefix(encodedHash, "$argon2id$v=19$") {
		t.Fatalf("unexpected hash format")
	}

	matches, err := VerifyPassword(password, encodedHash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !matches {
		t.Fatal("correct password did not match")
	}

	matches, err = VerifyPassword("incorrect-password", encodedHash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if matches {
		t.Fatal("incorrect password matched")
	}
}

func TestHashPasswordUsesUniqueSalts(t *testing.T) {
	password := "same-password"

	first, err := HashPassword(password)
	if err != nil {
		t.Fatalf("first HashPassword() error = %v", err)
	}

	second, err := HashPassword(password)
	if err != nil {
		t.Fatalf("second HashPassword() error = %v", err)
	}

	if first == second {
		t.Fatal("identical passwords produced identical encoded hashes")
	}
}

func TestVerifyPasswordRejectsMalformedHashes(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"plaintext", "not-a-hash"},
		{"wrong algorithm", "$argon2i$v=19$m=65536,t=3,p=4$abc$def"},
		{"wrong version", "$argon2id$v=16$m=65536,t=3,p=4$abc$def"},
		{"missing fields", "$argon2id$v=19$m=65536,t=3,p=4"},
		{"excessive memory", "$argon2id$v=19$m=4294967295,t=3,p=4$abc$def"},
		{"zero iterations", "$argon2id$v=19$m=65536,t=0,p=4$abc$def"},
		{"invalid parallelism", "$argon2id$v=19$m=65536,t=3,p=255$abc$def"},
		{"invalid base64", "$argon2id$v=19$m=65536,t=3,p=4$!invalid!$!invalid!"},
		{"trailing parameter data", "$argon2id$v=19$m=65536,t=3,p=4,extra=1$abc$def"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := VerifyPassword("password", tt.hash)

			if !errors.Is(err, ErrInvalidPasswordHash) {
				t.Fatalf("expected ErrInvalidPasswordHash, got %v", err)
			}
			if matches {
				t.Fatal("malformed hash matched")
			}
		})
	}
}

func TestVerifyPasswordRejectsTamperedHash(t *testing.T) {
	encodedHash, err := HashPassword("original-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		t.Fatalf("unexpected encoded hash format")
	}

	// Change one valid Base64 character without changing the hash length.
	if parts[5][0] == 'A' {
		parts[5] = "B" + parts[5][1:]
	} else {
		parts[5] = "A" + parts[5][1:]
	}

	tampered := strings.Join(parts, "$")

	matches, err := VerifyPassword("original-password", tampered)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if matches {
		t.Fatal("tampered hash matched")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	for b.Loop() {
		_, err := HashPassword("benchmark-password-2026")
		if err != nil {
			b.Fatal(err)
		}
	}
}
