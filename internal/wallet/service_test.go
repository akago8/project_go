package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type repoStub struct {
	adjust func(ctx context.Context, id uuid.UUID, op OperationType, amount int64) (Wallet, error)
	get    func(ctx context.Context, id uuid.UUID) (Wallet, error)
}

func (r repoStub) AdjustBalance(ctx context.Context, id uuid.UUID, op OperationType, amount int64) (Wallet, error) {
	return r.adjust(ctx, id, op, amount)
}

func (r repoStub) GetBalance(ctx context.Context, id uuid.UUID) (Wallet, error) {
	return r.get(ctx, id)
}

func TestHandleOperation(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	stub := repoStub{
		adjust: func(ctx context.Context, walletID uuid.UUID, op OperationType, amount int64) (Wallet, error) {
			return Wallet{ID: walletID, Balance: 2000, UpdatedAt: now}, nil
		},
		get: func(ctx context.Context, walletID uuid.UUID) (Wallet, error) {
			return Wallet{ID: walletID, Balance: 2000, UpdatedAt: now}, nil
		},
	}
	service := NewService(stub)
	result, err := service.HandleOperation(context.Background(), OperationRequest{
		WalletID:  id,
		Operation: OperationDeposit,
		Amount:    1000,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2000), result.Balance)
	require.Equal(t, id, result.ID)
}

func TestHandleOperationValidation(t *testing.T) {
	service := NewService(repoStub{})
	_, err := service.HandleOperation(context.Background(), OperationRequest{})
	require.Error(t, err)
	require.Equal(t, ErrWalletNotFound, err)
}

func TestGetBalance(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	service := NewService(repoStub{
		adjust: func(ctx context.Context, walletID uuid.UUID, op OperationType, amount int64) (Wallet, error) {
			return Wallet{}, nil
		},
		get: func(ctx context.Context, walletID uuid.UUID) (Wallet, error) {
			return Wallet{ID: walletID, Balance: 500, UpdatedAt: now}, nil
		},
	})
	result, err := service.GetBalance(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, int64(500), result.Balance)
}
