package flight

import (
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	FlightNumber     string    `gorm:"not null;index:idx_flight_number_date,unique"`
	DepartureDate    time.Time `gorm:"not null;index:idx_flight_number_date,unique"`
	Airline          string    `gorm:"not null"`
	OriginAirport    string    `gorm:"type:char(3);not null"`
	DestinationAirport string  `gorm:"type:char(3);not null"`
	DepartureTime    time.Time `gorm:"not null"`
	ArrivalTime      time.Time `gorm:"not null"`
	TotalSeats       int64       `gorm:"not null;check:total_seats > 0"`
	Price            float64   `gorm:"not null;check:price > 0"`
	Status           string    `gorm:"not null;check:status IN ('SCHEDULED','DEPARTED','CANCELLED','COMPLETED')"`
}

type SeatReservation struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	FlightID  uuid.UUID `gorm:"type:uuid;not null;index"`
	BookingID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`

	ReservedSeats int64       `gorm:"not null;check:reserved_seats >= 0"`
	Status        string    `gorm:"not null;check:status IN ('ACTIVE','RELEASED','EXPIRED')"`

	CreatedAt time.Time
	UpdatedAt time.Time
}