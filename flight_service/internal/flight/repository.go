package flight

import (
    "context"
    "time"
    
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Repository interface {
    ListByRoute(ctx context.Context, origin, dest string, date time.Time) ([]Flight, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Flight, error)
    
    SumActiveSeatsByFlight(ctx context.Context, flightID uuid.UUID) (int64, error)
    CreateReservation(ctx context.Context, reservation *SeatReservation) error
    GetReservation(ctx context.Context, id uuid.UUID) (*SeatReservation, error)
    UpdateReservationStatus(ctx context.Context, id uuid.UUID, status string) error
}

type repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{db: db}
}

func (r *repository) ListByRoute(ctx context.Context, origin, dest string, date time.Time) ([]Flight, error) {
    var flights []Flight
    return flights, r.db.WithContext(ctx).
        Where("origin_airport = ? AND destination_airport = ? AND departure_date = ?", origin, dest, date).
        Find(&flights).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Flight, error) {
    var flight Flight
    err := r.db.WithContext(ctx).First(&flight, "id = ?", id).Error
    return &flight, err
}

func (r *repository) SumActiveSeatsByFlight(ctx context.Context, flightID uuid.UUID) (int64, error) {
    var sum int64
    err := r.db.WithContext(ctx).
        Model(&SeatReservation{}).
        Where("flight_id = ? AND status = ?", flightID, "ACTIVE").
        Select("COALESCE(SUM(reserved_seats), 0)").Scan(&sum).Error
    return sum, err
}

func (r *repository) CreateReservation(ctx context.Context, reservation *SeatReservation) error {
    return r.db.WithContext(ctx).Create(reservation).Error
}

func (r *repository) GetReservation(ctx context.Context, id uuid.UUID) (*SeatReservation, error) {
    var reservation SeatReservation
    err := r.db.WithContext(ctx).First(&reservation, "id = ?", id).Error
    return &reservation, err
}

func (r *repository) UpdateReservationStatus(ctx context.Context, id uuid.UUID, status string) error {
    return r.db.WithContext(ctx).
        Model(&SeatReservation{}).
        Where("id = ?", id).
        Update("status", status).Error
}
