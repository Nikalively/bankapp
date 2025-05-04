package repo

import (
	"bankapp/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TransactionRepo struct {
	db *sqlx.DB
}

func NewTransactionRepo(db *sqlx.DB) *TransactionRepo {
	return &TransactionRepo{db}
}

func (r *TransactionRepo) Create(t *models.Transaction) error {
	t.ID = uuid.New()
	_, err := r.db.NamedExec(`
        INSERT INTO transactions
            (id, from_account_id, to_account_id, amount, transaction_type, note)
        VALUES
            (:id, :from_account_id, :to_account_id, :amount, :transaction_type, :note)
    `, t)
	return err
}

func (r *TransactionRepo) CreateTx(tx TxContext, t *models.Transaction) error {
	t.ID = uuid.New()
	_, err := tx.NamedExec(`
        INSERT INTO transactions
            (id, from_account_id, to_account_id, amount, transaction_type, note)
        VALUES
            (:id, :from_account_id, :to_account_id, :amount, :transaction_type, :note)
    `, t)
	return err
}
