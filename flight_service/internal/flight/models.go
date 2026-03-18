package flight

import (
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	ID                 uuid.UUID 	`gorm:"type:uuid;primaryKey;comment:Primary key"`
	FlightNumber       string    	`gorm:"not null;index:idx_flight_number_date,unique;comment:Flight number"`
	DepartureDate      time.Time 	`gorm:"not null;index:idx_flight_number_date,unique;comment:Departure date"`
	Airline            string    	`gorm:"not null;comment:Airline"`
	OriginAirport      string    	`gorm:"type:char(3);not null;comment:Origin airport code"`
	DestinationAirport string    	`gorm:"type:char(3);not null;comment:Destination airport code"`
	DepartureTime      time.Time 	`gorm:"not null;comment:Scheduled departure time"`
	ArrivalTime        time.Time 	`gorm:"not null;comment:Scheduled arrival time"`
	TotalSeats         int64     	`gorm:"not null;check:total_seats > 0;comment:Total seats, must be > 0"`
	Price              float64   	`gorm:"not null;check:price > 0;comment:Ticket price, must be > 0"`
	Status             string    	`gorm:"not null;check:status IN ('SCHEDULED','DEPARTED','CANCELLED','COMPLETED');comment:Flight status"`
}

type SeatReservation struct {
	ID           	   uuid.UUID 	`gorm:"type:uuid;primaryKey;comment:Primary key"`
	FlightID      	   uuid.UUID 	`gorm:"type:uuid;not null;index;comment:Foreign key to flights table"`
	BookingID          uuid.UUID 	`gorm:"type:uuid;not null;uniqueIndex;comment:Booking identifier"`
	ReservedSeats      int64     	`gorm:"not null;check:reserved_seats >= 0;comment:Number of reserved seats, must be >= 0"`
	Status        	   string    	`gorm:"not null;check:status IN ('ACTIVE','RELEASED','EXPIRED');comment:Reservation status"`
	CreatedAt          time.Time
	UpdatedAt     	   time.Time
}