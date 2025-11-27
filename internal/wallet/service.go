package wallet

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

type OperationRequest struct {
	WalletID  uuid.UUID
	Operation OperationType
	Amount    int64
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) HandleOperation(ctx context.Context, req OperationRequest) (wallet Wallet, err error) {
	if req.WalletID == uuid.Nil {
		return wallet, ErrWalletNotFound
	}
	if req.Amount <= 0 {
		return wallet, ErrInvalidAmount
	}
	if req.Operation != OperationDeposit && req.Operation != OperationWithdraw {
		return wallet, ErrInvalidOperation
	}
	return s.repo.AdjustBalance(ctx, req.WalletID, req.Operation, req.Amount)
}

func (s Service) GetBalance(ctx context.Context, id uuid.UUID) (wallet Wallet, err error) {
	if id == uuid.Nil {
		return wallet, ErrWalletNotFound
	}
	return s.repo.GetBalance(ctx, id)
}
