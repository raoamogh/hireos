//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/raoamogh/hireos/internal/identity"
)

var testUserSequence atomic.Uint64

func createTestUser(t *testing.T) identity.UserID {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email := fmt.Sprintf(
		"action-token-%d-%d@example.test",
		time.Now().UnixNano(),
		testUserSequence.Add(1),
	)

	var userID string
	err := integrationPool.QueryRow(
		ctx,
		`INSERT INTO users (email, display_name)
		 VALUES ($1, $2)
		 RETURNING id::text`,
		email,
		"Action Token Integration Test",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	_, err = integrationPool.Exec(
		ctx,
		`INSERT INTO user_password_credentials (user_id, password_hash)
		 VALUES ($1, $2)`,
		userID,
		"test-password-hash",
	)
	if err != nil {
		t.Fatalf("create test password credential: %v", err)
	}

	id := identity.UserID(userID)

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cleanupCancel()

		_, cleanupErr := integrationPool.Exec(
			cleanupCtx,
			`DELETE FROM users WHERE id = $1`,
			string(id),
		)
		if cleanupErr != nil {
			t.Errorf("clean up test user: %v", cleanupErr)
		}
	})

	return id
}

func issueTestToken(
	t *testing.T,
	userID identity.UserID,
	purpose identity.ActionTokenPurpose,
	tokenHash [32]byte,
	expiresAt time.Time,
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := integrationRepo.WithinTransaction(
		ctx,
		func(tx identity.ActionTokenTransaction) error {
			if err := tx.LockUser(ctx, userID); err != nil {
				return err
			}

			if err := tx.InvalidateOutstanding(ctx, userID, purpose); err != nil {
				return err
			}

			return tx.Create(ctx, userID, purpose, tokenHash, expiresAt)
		},
	)
	if err != nil {
		t.Fatalf("issue test token: %v", err)
	}
}

func consumeTestToken(
	ctx context.Context,
	tokenHash [32]byte,
	purpose identity.ActionTokenPurpose,
) (identity.UserID, error) {
	var userID identity.UserID

	err := integrationRepo.WithinTransaction(
		ctx,
		func(tx identity.ActionTokenTransaction) error {
			var err error
			userID, err = tx.Consume(ctx, tokenHash, purpose)
			return err
		},
	)

	return userID, err
}

func TestActionTokenIssuanceInvalidatesPreviousToken(t *testing.T) {
	userID := createTestUser(t)
	purpose := identity.ActionTokenPurposePasswordReset

	firstHash := sha256.Sum256([]byte("first-token"))
	secondHash := sha256.Sum256([]byte("second-token"))
	expiry := time.Now().Add(time.Hour)

	issueTestToken(t, userID, purpose, firstHash, expiry)
	issueTestToken(t, userID, purpose, secondHash, expiry)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := consumeTestToken(ctx, firstHash, purpose); !errors.Is(
		err,
		identity.ErrInvalidActionToken,
	) {
		t.Fatalf("expected previous token to be invalid, got %v", err)
	}

	gotUserID, err := consumeTestToken(ctx, secondHash, purpose)
	if err != nil {
		t.Fatalf("consume latest token: %v", err)
	}

	if gotUserID != userID {
		t.Fatalf("expected user %q, got %q", userID, gotUserID)
	}
}

func TestActionTokenCanOnlyBeConsumedOnce(t *testing.T) {
	userID := createTestUser(t)
	purpose := identity.ActionTokenPurposeEmailVerification
	tokenHash := sha256.Sum256([]byte("single-use-token"))

	issueTestToken(t, userID, purpose, tokenHash, time.Now().Add(time.Hour))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := consumeTestToken(ctx, tokenHash, purpose); err != nil {
		t.Fatalf("first consumption should succeed: %v", err)
	}

	if _, err := consumeTestToken(ctx, tokenHash, purpose); !errors.Is(
		err,
		identity.ErrInvalidActionToken,
	) {
		t.Fatalf("expected second consumption to fail, got %v", err)
	}
}

func TestActionTokenRejectsExpiredToken(t *testing.T) {
	userID := createTestUser(t)
	purpose := identity.ActionTokenPurposePasswordReset
	tokenHash := sha256.Sum256([]byte("expired-token"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := integrationPool.Exec(
		ctx,
		`INSERT INTO user_action_tokens (
			user_id,
			purpose,
			token_hash,
			created_at,
			expires_at
		)
		VALUES ($1, $2, $3, statement_timestamp() - interval '2 hours',
			statement_timestamp() - interval '1 hour')`,
		string(userID),
		string(purpose),
		tokenHash[:],
	)
	if err != nil {
		t.Fatalf("insert expired test token: %v", err)
	}

	if _, err := consumeTestToken(ctx, tokenHash, purpose); !errors.Is(
		err,
		identity.ErrInvalidActionToken,
	) {
		t.Fatalf("expected expired token to be rejected, got %v", err)
	}
}

func TestActionTokenConsumptionRollsBackWhenPasswordUpdateFails(t *testing.T) {
	userID := createTestUser(t)
	purpose := identity.ActionTokenPurposePasswordReset
	tokenHash := sha256.Sum256([]byte("rollback-token"))

	issueTestToken(t, userID, purpose, tokenHash, time.Now().Add(time.Hour))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := integrationPool.Exec(
		ctx,
		`DELETE FROM user_password_credentials WHERE user_id = $1`,
		string(userID),
	)
	if err != nil {
		t.Fatalf("remove test password credential: %v", err)
	}

	err = integrationRepo.WithinTransaction(
		ctx,
		func(tx identity.ActionTokenTransaction) error {
			consumedUserID, err := tx.Consume(ctx, tokenHash, purpose)
			if err != nil {
				return err
			}

			if consumedUserID != userID {
				return fmt.Errorf("consumed token belongs to unexpected user")
			}

			return tx.UpdatePasswordCredential(
				ctx,
				consumedUserID,
				"replacement-test-password-hash",
			)
		},
	)

	if !errors.Is(err, identity.ErrPasswordCredentialNotFound) {
		t.Fatalf(
			"expected missing password credential error, got %v",
			err,
		)
	}

	_, err = integrationPool.Exec(
		ctx,
		`INSERT INTO user_password_credentials (user_id, password_hash)
		 VALUES ($1, $2)`,
		string(userID),
		"restored-test-password-hash",
	)
	if err != nil {
		t.Fatalf("restore test password credential: %v", err)
	}

	gotUserID, err := consumeTestToken(ctx, tokenHash, purpose)
	if err != nil {
		t.Fatalf("token should remain usable after rollback: %v", err)
	}

	if gotUserID != userID {
		t.Fatalf("expected user %q, got %q", userID, gotUserID)
	}
}

func TestConcurrentActionTokenConsumptionOnlySucceedsOnce(t *testing.T) {
	userID := createTestUser(t)
	purpose := identity.ActionTokenPurposePasswordReset
	tokenHash := sha256.Sum256([]byte("concurrent-token"))

	issueTestToken(t, userID, purpose, tokenHash, time.Now().Add(time.Hour))

	const workers = 2

	start := make(chan struct{})
	results := make(chan error, workers)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start

			ctx, cancel := context.WithTimeout(
				context.Background(),
				10*time.Second,
			)
			defer cancel()

			_, err := consumeTestToken(ctx, tokenHash, purpose)
			results <- err
		}()
	}

	close(start)

	wg.Wait()
	close(results)

	successes := 0
	invalidTokens := 0

	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, identity.ErrInvalidActionToken):
			invalidTokens++
		default:
			t.Fatalf("unexpected concurrent consumption error: %v", err)
		}
	}

	if successes != 1 || invalidTokens != 1 {
		t.Fatalf(
			"expected exactly one success and one rejection; got %d successes and %d rejections",
			successes,
			invalidTokens,
		)
	}
}
