package wallet

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository interface {
	AdjustBalance(ctx context.Context, id uuid.UUID, op OperationType, amount int64) (Wallet, error)
	GetBalance(ctx context.Context, id uuid.UUID) (Wallet, error)
}

type SQLRepository struct {
	db      *sql.DB
	retries int
}

func NewRepository(db *sql.DB, retries int) *SQLRepository {
	return &SQLRepository{db: db, retries: retries}
}

func (r *SQLRepository) AdjustBalance(ctx context.Context, id uuid.UUID, op OperationType, amount int64) (result Wallet, err error) {
	if amount <= 0 {
		return result, ErrInvalidAmount
	}
	for attempt := 0; attempt <= r.retries; attempt++ {
		result, err = r.adjustOnce(ctx, id, op, amount)
		if err == nil {
			return result, nil
		}
		if !isRetryable(err) {
			return result, err
		}
	}
	return result, err
}

func (r *SQLRepository) adjustOnce(ctx context.Context, id uuid.UUID, op OperationType, amount int64) (wallet Wallet, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return wallet, err
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `SELECT id, balance, updated_at FROM wallets WHERE id = $1 FOR UPDATE`, id)
	var current Wallet
	switch scanErr := row.Scan(&current.ID, &current.Balance, &current.UpdatedAt); scanErr {
	case nil:
	default:
		if errors.Is(scanErr, sql.ErrNoRows) {
			if op == OperationWithdraw {
				return wallet, ErrWalletNotFound
			}
			return r.insertWallet(ctx, tx, id, amount)
		}
		return wallet, scanErr
	}
	var newBalance int64
	if op == OperationDeposit {
		newBalance = current.Balance + amount
	} else if op == OperationWithdraw {
		newBalance = current.Balance - amount
		if newBalance < 0 {
			return wallet, ErrInsufficientFunds
		}
	} else {
		return wallet, ErrInvalidOperation
	}
	row = tx.QueryRowContext(ctx, `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2 RETURNING id, balance, updated_at`, newBalance, id)
	if err := row.Scan(&wallet.ID, &wallet.Balance, &wallet.UpdatedAt); err != nil {
		return wallet, err
	}
	if err := tx.Commit(); err != nil {
		return wallet, err
	}
	return wallet, nil
}

func (r *SQLRepository) insertWallet(ctx context.Context, tx *sql.Tx, id uuid.UUID, amount int64) (wallet Wallet, err error) {
	row := tx.QueryRowContext(ctx, `INSERT INTO wallets (id, balance) VALUES ($1, $2) RETURNING id, balance, updated_at`, id, amount)
	if err := row.Scan(&wallet.ID, &wallet.Balance, &wallet.UpdatedAt); err != nil {
		return wallet, err
	}
	if err := tx.Commit(); err != nil {
		return wallet, err
	}
	return wallet, nil
}

func (r *SQLRepository) GetBalance(ctx context.Context, id uuid.UUID) (wallet Wallet, err error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, balance, updated_at FROM wallets WHERE id = $1`, id)
	if err := row.Scan(&wallet.ID, &wallet.Balance, &wallet.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return wallet, ErrWalletNotFound
		}
		return wallet, err
	}
	return wallet, nil
}

func isRetryable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40001" || pgErr.Code == "40P01"
	}
	return false
}
