package identity

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserNotFound               = errors.New("user not found")
	ErrInvalidActionToken         = errors.New("invalid or expired action token")
	ErrInvalidActionTokenPurpose  = errors.New("invalid action token purpose")
	ErrPasswordCredentialNotFound = errors.New("password credential not found")
)

type UserID string

type ActionTokenPurpose string

const (
	ActionTokenPurposeEmailVerification ActionTokenPurpose = "email_verification"
	ActionTokenPurposePasswordReset     ActionTokenPurpose = "password_reset"
)

type ActionTokenTransaction interface {
	LockUser(ctx context.Context, userID UserID) error

	InvalidateOutstanding(
		ctx context.Context,
		userID UserID,
		purpose ActionTokenPurpose,
	) error

	Create(
		ctx context.Context,
		userID UserID,
		purpose ActionTokenPurpose,
		tokenHash [32]byte,
		expiresAt time.Time,
	) error

	Consume(
		ctx context.Context,
		tokenHash [32]byte,
		purpose ActionTokenPurpose,
	) (UserID, error)

	UpdatePasswordCredential(
		ctx context.Context,
		userID UserID,
		passwordHash string,
	) error

	MarkEmailVerified(
		ctx context.Context,
		userID UserID,
	) error
}

type ActionTokenRepository interface {
	WithinTransaction(
		ctx context.Context,
		fn func(ActionTokenTransaction) error,
	) error
}
