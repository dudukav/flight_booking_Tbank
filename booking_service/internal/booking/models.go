package booking

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         string    `gorm:"not null"`
	FlightID       uuid.UUID `gorm:"not null"`
	PassengerName  string    `gorm:"not null"`
	PassengerEmail string    `gorm:"not null"`
	SeatCount      int32     `gorm:"check:seat_count > 0"`
	TotalPrice     float32   `gorm:"check:total_price > 0"`
	Status         string    `gorm:"check:status IN ('CONFIRMED', 'CANCELLED')"`
	CreatedAt      time.Time
}
