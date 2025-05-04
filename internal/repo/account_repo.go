package repo

import (
	"database/sql"
	"errors"
	"github.com/shopspring/decimal"

	"bankapp/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AccountRepo struct {
	db *sqlx.DB
}

func NewAccountRepo(db *sqlx.DB) *AccountRepo {
	return &AccountRepo{db}
}

func (r *AccountRepo) Create(a *models.Account) error {
	a.ID = uuid.New()
	_, err := r.db.NamedExec(`
        INSERT INTO accounts (id, user_id, number, balance)
        VALUES (:id, :user_id, :number, :balance)
    `, a)
	return err
}

func (r *AccountRepo) GetByUserID(userID uuid.UUID) ([]models.Account, error) {
	var list []models.Account
	err := r.db.Select(&list, `
        SELECT id, user_id, number, balance, created_at
        FROM accounts WHERE user_id=$1
    `, userID)
	return list, err
}

func (r *AccountRepo) GetByID(id uuid.UUID) (*models.Account, error) {
	var a models.Account
	err := r.db.Get(&a, `
        SELECT id, user_id, number, balance, created_at
        FROM accounts WHERE id=$1
    `, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	return &a, err
}

func (r *AccountRepo) WithTx(fn func(TxContext) error) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	ctx := tx.(TxContext)
	if err := fn(ctx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// обновление баланса
func (r *AccountRepo) UpdateBalanceTx(tx TxContext, id uuid.UUID, newBalance decimal.Decimal) error {
	_, err := tx.Exec(`
        UPDATE accounts SET balance=$2 WHERE id=$1
    `, id, newBalance)
	return err
}
