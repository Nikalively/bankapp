package repo

import (
	"database/sql"
	"errors"

	"bankapp/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db}
}

func (r *UserRepo) Create(u *models.User) error {
	u.ID = uuid.New()
	_, err := r.db.NamedExec(`
        INSERT INTO users (id, username, email, password_hash)
        VALUES (:id, :username, :email, :password_hash)
    `, u)
	return err
}

func (r *UserRepo) GetByUsername(username string) (*models.User, error) {
	var u models.User
	err := r.db.Get(&u, `
        SELECT id, username, email, password_hash, created_at
        FROM users WHERE username=$1
    `, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.Get(&u, `
        SELECT id, username, email, password_hash, created_at
        FROM users WHERE email=$1
    `, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.Get(&u, `
        SELECT id, username, email, password_hash, created_at
        FROM users WHERE id=$1
    `, id)
	return &u, err
}
