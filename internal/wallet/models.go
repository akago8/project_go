package wallet

import (
	"time"

	"github.com/google/uuid"
)

type OperationType string

const (
	OperationDeposit  OperationType = "DEPOSIT"
	OperationWithdraw OperationType = "WITHDRAW"
)

type Wallet struct {
	ID        uuid.UUID
	Balance   int64
	UpdatedAt time.Time
}
