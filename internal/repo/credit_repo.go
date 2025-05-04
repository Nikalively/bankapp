package repo

import (
	"bankapp/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CreditRepo struct {
	db *sqlx.DB
}

func NewCreditRepo(db *sqlx.DB) *CreditRepo {
	return &CreditRepo{db}
}

func (r *CreditRepo) Create(c *models.Credit) error {
	c.ID = uuid.New()
	_, err := r.db.NamedExec(`
        INSERT INTO credits
          (id, user_id, account_id, principal, annual_rate, term_months, start_at, remaining)
        VALUES
          (:id, :user_id, :account_id, :principal, :annual_rate, :term_months, :start_at, :remaining)
    `, c)
	return err
}

func (r *CreditRepo) GetByUserID(userID uuid.UUID) ([]models.Credit, error) {
	var list []models.Credit
	err := r.db.Select(&list, `
        SELECT id, user_id, account_id, principal, annual_rate, term_months, start_at, remaining, created_at
        FROM credits WHERE user_id=$1
    `, userID)
	return list, err
}

func (r *CreditRepo) GetByID(id uuid.UUID) (*models.Credit, error) {
	var c models.Credit
	err := r.db.Get(&c, `
        SELECT id, user_id, account_id, principal, annual_rate, term_months, start_at, remaining, created_at
        FROM credits WHERE id=$1
    `, id)
	return &c, err
}

func (r *CreditRepo) WithTx(fn func(TxContext) error) error {
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

func (r *CreditRepo) CreateTx(tx TxContext, c *models.Credit) error {
	c.ID = uuid.New()
	_, err := tx.NamedExec(`
        INSERT INTO credits
          (id, user_id, account_id, principal, annual_rate, term_months, start_at, remaining)
        VALUES
          (:id, :user_id, :account_id, :principal, :annual_rate, :term_months, :start_at, :remaining)
    `, c)
	return err
}
