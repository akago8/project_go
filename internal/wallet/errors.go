package wallet

import "errors"

var (
	ErrInvalidOperation  = errors.New("invalid operation")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
)
