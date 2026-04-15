package appointments

// Data layer, should exclusively contain data access logic
import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AppointmentDL
type AppointmentDL struct {
	ID        int64     `json:"-" db:"id"`
	UserID    int64     `json:"-" db:"user_id"`
	TrainerID int64     `json:"-" db:"trainer_id"`
	End       time.Time `json:"-" db:"ended_at"`
	Start     time.Time `json:"-" db:"started_at"`
}

// AppointmentDAO
type AppointmentDAO interface {
	GetAppointments(ctx context.Context, trainerID int64) ([]AppointmentDL, error)
	CreateAppointment(ctx context.Context, apt AppointmentDL) (AppointmentDL, error)
}

// I tend to do this so you can easily mock the data layer for testing
func NewDAO(pool *pgxpool.Pool) AppointmentDAO {
	return &AppointmentDAOImpl{pool: pool}
}

// AppointmentDAOImpl
type AppointmentDAOImpl struct {
	pool *pgxpool.Pool
}

// GetAppointments
func (d *AppointmentDAOImpl) GetAppointments(ctx context.Context, trainerID int64) ([]AppointmentDL, error) {
	q := `
		SELECT id, user_id, trainer_id, started_at, ended_at
		FROM appointments
		WHERE trainer_id = $1
		ORDER BY started_at
	`
	rows, err := d.pool.Query(ctx, q, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AppointmentDL
	for rows.Next() {
		var a AppointmentDL
		if err := rows.Scan(&a.ID, &a.UserID, &a.TrainerID, &a.Start, &a.End); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateAppointment
func (d *AppointmentDAOImpl) CreateAppointment(ctx context.Context, apt AppointmentDL) (AppointmentDL, error) {
	const q = `
		INSERT INTO appointments (user_id, trainer_id, started_at, ended_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, trainer_id, started_at, ended_at
	`
	var created AppointmentDL
	err := d.pool.QueryRow(ctx, q, apt.UserID, apt.TrainerID, apt.Start, apt.End).
		Scan(&created.ID, &created.UserID, &created.TrainerID, &created.Start, &created.End)
	if err != nil {
		return AppointmentDL{}, err
	}
	return created, nil
}
