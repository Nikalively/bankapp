package repo

import (
	"bankapp/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

type ScheduleRepo struct {
	db *sqlx.DB
}

func NewScheduleRepo(db *sqlx.DB) *ScheduleRepo {
	return &ScheduleRepo{db}
}

func (r *ScheduleRepo) Create(s *models.PaymentSchedule) error {
	s.ID = uuid.New()
	_, err := r.db.NamedExec(`
        INSERT INTO payment_schedules
          (id, credit_id, due_date, amount, principal, interest)
        VALUES
          (:id, :credit_id, :due_date, :amount, :principal, :interest)
    `, s)
	return err
}

func (r *ScheduleRepo) CreateTx(tx TxContext, s *models.PaymentSchedule) error {
	s.ID = uuid.New()
	_, err := tx.NamedExec(`
        INSERT INTO payment_schedules
          (id, credit_id, due_date, amount, principal, interest)
        VALUES
          (:id, :credit_id, :due_date, :amount, :principal, :interest)
    `, s)
	return err
}

func (r *ScheduleRepo) GetByCreditID(creditID uuid.UUID) ([]models.PaymentSchedule, error) {
	var list []models.PaymentSchedule
	err := r.db.Select(&list, `
        SELECT id, credit_id, due_date, amount, principal, interest, paid
        FROM payment_schedules WHERE credit_id=$1
    `, creditID)
	return list, err
}

func (r *ScheduleRepo) GetDueSchedules(before time.Time) ([]models.PaymentSchedule, error) {
	var list []models.PaymentSchedule
	err := r.db.Select(&list, `
        SELECT id, credit_id, due_date, amount, principal, interest, paid
        FROM payment_schedules
        WHERE due_date <= $1 AND paid = false
    `, before)
	return list, err
}

func (r *ScheduleRepo) UpdatePaidTx(tx TxContext, id uuid.UUID, paid bool) error {
	_, err := tx.Exec(`
        UPDATE payment_schedules
        SET paid = $2
        WHERE id = $1
    `, id, paid)
	return err
}
