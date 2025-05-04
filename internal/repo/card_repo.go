package repo

import (
	"bankapp/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CardRepo struct {
	db *sqlx.DB
}

func NewCardRepo(db *sqlx.DB) *CardRepo {
	return &CardRepo{db}
}

func (r *CardRepo) Create(c *models.Card) error {
	c.ID = uuid.New()
	_, err := r.db.NamedExec(`
        INSERT INTO cards (id, account_id, number_enc, expiry_enc, cvv_hash, hmac)
        VALUES (:id, :account_id, :number_enc, :expiry_enc, :cvv_hash, :hmac)
    `, c)
	return err
}

func (r *CardRepo) GetByAccountID(accountID uuid.UUID) ([]models.Card, error) {
	var list []models.Card
	err := r.db.Select(&list, `
        SELECT id, account_id, number_enc, expiry_enc, cvv_hash, hmac, created_at
        FROM cards WHERE account_id=$1
    `, accountID)
	return list, err
}

func (r *CardRepo) GetByHMAC(hmacHex string) (*models.Card, error) {
	var c models.Card
	err := r.db.Get(&c, `
        SELECT id, account_id, number_enc, expiry_enc, cvv_hash, hmac, created_at
        FROM cards WHERE hmac=$1
    `, hmacHex)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
