package models

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be at least 6 characters")
)

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

func (u *User) Validate(passwordPlain string) error {
	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	re := regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`)
	if !re.MatchString(u.Email) {
		return ErrInvalidEmail
	}
	if len(passwordPlain) < 6 {
		return ErrInvalidPassword
	}
	return nil
}

type Account struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	UserID    uuid.UUID       `db:"user_id" json:"user_id"`
	Number    string          `db:"number" json:"number"`
	Balance   decimal.Decimal `db:"balance" json:"balance"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

type Card struct {
	ID        uuid.UUID `db:"id" json:"id"`
	AccountID uuid.UUID `db:"account_id" json:"account_id"`
	NumberEnc []byte    `db:"number_enc" json:"-"`
	ExpiryEnc []byte    `db:"expiry_enc" json:"-"`
	CVVHash   string    `db:"cvv_hash" json:"-"`
	HMAC      string    `db:"hmac" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Transaction struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	From      *uuid.UUID      `db:"from_account_id" json:"from_account_id,omitempty"`
	To        *uuid.UUID      `db:"to_account_id" json:"to_account_id,omitempty"`
	Amount    decimal.Decimal `db:"amount" json:"amount"`
	Type      string          `db:"transaction_type" json:"transaction_type"`
	Note      string          `db:"note" json:"note,omitempty"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

type Credit struct {
	ID         uuid.UUID       `db:"id" json:"id"`
	UserID     uuid.UUID       `db:"user_id" json:"user_id"`
	AccountID  uuid.UUID       `db:"account_id" json:"account_id"`
	Principal  decimal.Decimal `db:"principal" json:"principal"`
	AnnualRate decimal.Decimal `db:"annual_rate" json:"annual_rate"`
	TermMonths int             `db:"term_months" json:"term_months"`
	StartAt    time.Time       `db:"start_at" json:"start_at"`
	Remaining  decimal.Decimal `db:"remaining" json:"remaining"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
}

type PaymentSchedule struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	CreditID  uuid.UUID       `db:"credit_id" json:"credit_id"`
	DueDate   time.Time       `db:"due_date" json:"due_date"`
	Amount    decimal.Decimal `db:"amount" json:"amount"`
	Principal decimal.Decimal `db:"principal" json:"principal"`
	Interest  decimal.Decimal `db:"interest" json:"interest"`
	Paid      bool            `db:"paid" json:"paid"`
}

// DTO
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type PaymentRequest struct {
	CardNumber string          `json:"card_number"`
	Amount     decimal.Decimal `json:"amount"`
	Merchant   string          `json:"merchant"`
}
type TransferRequest struct {
	FromAccountID uuid.UUID       `json:"from_account_id"`
	ToAccountID   uuid.UUID       `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
}
type DepositRequest struct {
	ToAccountID uuid.UUID       `json:"to_account_id"`
	Amount      decimal.Decimal `json:"amount"`
}
type ApplyCreditRequest struct {
	AccountID  uuid.UUID       `json:"account_id"`
	Principal  decimal.Decimal `json:"principal"`
	TermMonths int             `json:"term_months"`
}
