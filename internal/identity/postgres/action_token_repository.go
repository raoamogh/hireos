package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raoamogh/hireos/internal/identity"
	db "github.com/raoamogh/hireos/internal/platform/database/sqlc"
)

type ActionTokenRepository struct {
	pool *pgxpool.Pool
}

var _ identity.ActionTokenRepository = (*ActionTokenRepository)(nil)

func NewActionTokenRepository(pool *pgxpool.Pool) *ActionTokenRepository {
	return &ActionTokenRepository{
		pool: pool,
	}
}

func (r *ActionTokenRepository) WithinTransaction(
	ctx context.Context,
	fn func(identity.ActionTokenTransaction) error,
) (err error) {
	if r.pool == nil {
		return errors.New("PostgreSQL connection pool is nil")
	}

	if fn == nil {
		return errors.New("action token transaction callback is nil")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin action token transaction: %w", err)
	}

	committed := false

	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback(context.Background())
			panic(recovered)
		}

		if !committed {
			if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil &&
				!errors.Is(rollbackErr, pgx.ErrTxClosed) {
				err = errors.Join(
					err,
					fmt.Errorf("rollback action token transaction: %w", rollbackErr),
				)
			}
		}
	}()

	queries := db.New(tx)
	transaction := &actionTokenTransaction{
		queries: queries,
	}

	if err := fn(transaction); err != nil {
		return fmt.Errorf("execute action token transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit action token transaction: %w", err)
	}

	committed = true
	return nil
}

type actionTokenTransaction struct {
	queries *db.Queries
}

func parseUserID(userID identity.UserID) (pgtype.UUID, error) {
	var id pgtype.UUID

	if err := id.Scan(string(userID)); err != nil {
		return pgtype.UUID{}, fmt.Errorf("parse user ID: %w", err)
	}

	if !id.Valid {
		return pgtype.UUID{}, errors.New("user ID is invalid")
	}

	return id, nil
}

func formatUserID(id pgtype.UUID) (identity.UserID, error) {
	if !id.Valid {
		return "", errors.New("database returned an invalid user ID")
	}

	b := id.Bytes

	return identity.UserID(fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)), nil
}

func validatePurpose(
	purpose identity.ActionTokenPurpose,
) error {
	switch purpose {
	case identity.ActionTokenPurposeEmailVerification,
		identity.ActionTokenPurposePasswordReset:
		return nil
	default:
		return fmt.Errorf(
			"%w: %q",
			identity.ErrInvalidActionTokenPurpose,
			purpose,
		)
	}
}

func (t *actionTokenTransaction) LockUser(
	ctx context.Context,
	userID identity.UserID,
) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	_, err = t.queries.LockUserForActionToken(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("lock user for action token: %w", err)
	}

	return nil
}

func (t *actionTokenTransaction) InvalidateOutstanding(
	ctx context.Context,
	userID identity.UserID,
	purpose identity.ActionTokenPurpose,
) error {
	if err := validatePurpose(purpose); err != nil {
		return err
	}

	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = t.queries.InvalidateOutstandingUserActionTokens(
		ctx,
		db.InvalidateOutstandingUserActionTokensParams{
			UserID:  id,
			Purpose: string(purpose),
		},
	)
	if err != nil {
		return fmt.Errorf("invalidate outstanding action tokens: %w", err)
	}

	return nil
}

func (t *actionTokenTransaction) Create(
	ctx context.Context,
	userID identity.UserID,
	purpose identity.ActionTokenPurpose,
	tokenHash [32]byte,
	expiresAt time.Time,
) error {
	if err := validatePurpose(purpose); err != nil {
		return err
	}

	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	_, err = t.queries.CreateUserActionToken(
		ctx,
		db.CreateUserActionTokenParams{
			UserID:    id,
			Purpose:   string(purpose),
			TokenHash: tokenHash[:],
			ExpiresAt: pgtype.Timestamptz{
				Time:  expiresAt,
				Valid: true,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("create action token: %w", err)
	}

	return nil
}

func (t *actionTokenTransaction) Consume(
	ctx context.Context,
	tokenHash [32]byte,
	purpose identity.ActionTokenPurpose,
) (identity.UserID, error) {
	if err := validatePurpose(purpose); err != nil {
		return "", err
	}

	result, err := t.queries.ConsumeUserActionToken(
		ctx,
		db.ConsumeUserActionTokenParams{
			TokenHash: tokenHash[:],
			Purpose:   string(purpose),
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", identity.ErrInvalidActionToken
	}
	if err != nil {
		return "", fmt.Errorf("consume action token: %w", err)
	}

	userID, err := formatUserID(result.UserID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (t *actionTokenTransaction) UpdatePasswordCredential(
	ctx context.Context,
	userID identity.UserID,
	passwordHash string,
) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	rowsAffected, err := t.queries.UpdateUserPasswordCredential(
		ctx,
		db.UpdateUserPasswordCredentialParams{
			UserID:       id,
			PasswordHash: passwordHash,
		},
	)
	if err != nil {
		return fmt.Errorf("update password credential: %w", err)
	}

	if rowsAffected == 0 {
		return identity.ErrPasswordCredentialNotFound
	}

	if rowsAffected != 1 {
		return fmt.Errorf(
			"update password credential: expected one affected row, got %d",
			rowsAffected,
		)
	}

	return nil
}

func (t *actionTokenTransaction) MarkEmailVerified(
	ctx context.Context,
	userID identity.UserID,
) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	rowsAffected, err := t.queries.MarkUserEmailVerified(ctx, id)
	if err != nil {
		return fmt.Errorf("mark user email verified: %w", err)
	}

	if rowsAffected == 0 {
		return identity.ErrUserNotFound
	}

	if rowsAffected != 1 {
		return fmt.Errorf(
			"mark user email verified: expected one affected row, got %d",
			rowsAffected,
		)
	}

	return nil
}
