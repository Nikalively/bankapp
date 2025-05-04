package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"bankapp/internal/models"
	"bankapp/internal/services"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type key int

const (
	ctxUserID key = iota
)

type Handler struct {
	svc *services.BankService
}

func NewHandler(svc *services.BankService) *Handler {
	return &Handler{svc: svc}
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, code int, msg string) {
	respondJSON(w, code, map[string]string{"error": msg})
}

// POST /register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	user, err := h.svc.RegisterUser(req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, user)
}

// POST /login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	token, err := h.svc.LoginUser(req)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		userID, err := h.svc.ParseToken(tokenStr)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(ctxUserID)
	if v == nil {
		return uuid.Nil, fmt.Errorf("no user in context")
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user id")
	}
	return id, nil
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	uid, err := userIDFromCtx(r.Context())
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	acc, err := h.svc.CreateAccount(uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, acc)
}

func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromCtx(r.Context())
	list, err := h.svc.GetUserAccounts(uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, list)
}

func (h *Handler) GenerateCard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account id")
		return
	}
	card, err := h.svc.GenerateCard(accID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, card)
}

func (h *Handler) GetCards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account id")
		return
	}
	list, err := h.svc.GetAccountCards(accID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, list)
}

// POST /payments
func (h *Handler) PayWithCard(w http.ResponseWriter, r *http.Request) {
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := h.svc.PayWithCard(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /transfers
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := h.svc.Transfer(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /deposits
func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req models.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := h.svc.Deposit(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /credits
func (h *Handler) ApplyCredit(w http.ResponseWriter, r *http.Request) {
	var req models.ApplyCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	credit, sched, err := h.svc.ApplyCredit(req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"credit":   credit,
		"schedule": sched,
	})
}
