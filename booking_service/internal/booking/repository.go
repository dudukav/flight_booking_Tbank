package booking

import (
    "context"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Repository interface {
    Create(ctx context.Context, booking *Booking) error
    GetByID(ctx context.Context, id uuid.UUID) (*Booking, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
    ListByUserID(ctx context.Context, userID string) ([]*Booking, error)
}

type repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, booking *Booking) error {
    return r.db.WithContext(ctx).Create(booking).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Booking, error) {
    var booking Booking
    err := r.db.WithContext(ctx).
        Where("id = ?", id).
        First(&booking).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, err
    }
    return &booking, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
    return r.db.WithContext(ctx).
        Model(&Booking{}).
        Where("id = ?", id).
        Where("status = ?", "CONFIRMED").
        Update("status", status).Error
}

func (r *repository) ListByUserID(ctx context.Context, userID string) ([]*Booking, error) {
    var bookings []*Booking
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Order("created_at desc").
        Find(&bookings).Error
    return bookings, err
}
