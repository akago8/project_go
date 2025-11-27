package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/akann77/wallet/internal/wallet"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type Handler struct {
	service wallet.Service
}

type operationPayload struct {
	WalletID      string `json:"walletId"`
	OperationType string `json:"operationType"`
	Amount        int64  `json:"amount"`
}

type walletResponse struct {
	WalletID string `json:"walletId"`
	Balance  int64  `json:"balance"`
}

func NewRouter(service wallet.Service) http.Handler {
	h := Handler{service: service}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Post("/api/v1/wallet", h.handleOperation)
	r.Get("/api/v1/wallets/{walletId}", h.handleGetBalance)
	return r
}

func (h Handler) handleOperation(w http.ResponseWriter, r *http.Request) {
	var payload operationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	id, err := uuid.Parse(payload.WalletID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid walletId")
		return
	}
	req := wallet.OperationRequest{
		WalletID:  id,
		Operation: wallet.OperationType(payload.OperationType),
		Amount:    payload.Amount,
	}
	result, err := h.service.HandleOperation(r.Context(), req)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, walletResponse{WalletID: result.ID.String(), Balance: result.Balance})
}

func (h Handler) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "walletId")
	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid walletId")
		return
	}
	result, err := h.service.GetBalance(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, walletResponse{WalletID: result.ID.String(), Balance: result.Balance})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch err {
	case wallet.ErrInvalidOperation, wallet.ErrInvalidAmount:
		writeError(w, http.StatusBadRequest, err.Error())
	case wallet.ErrWalletNotFound:
		writeError(w, http.StatusNotFound, err.Error())
	case wallet.ErrInsufficientFunds:
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
