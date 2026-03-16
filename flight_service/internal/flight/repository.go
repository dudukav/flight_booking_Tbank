package flight

import (
	"time"

	"github.com/redis/go-redis/v9"
	pb "flight_booking_Tbank/flight_gen"
)

type Repository struct {
	redisClient *redis.Client
}

func NewRepository(rdb *redis.Client) *Repository {
    return &Repository{
        redisClient: rdb,
    }
}

func (r *Repository) SearchFlights(from, to string, departureTime time.Time) ([]*pb.Flight, error)
func (r *Repository) GetFlight(flightID string) (*pb.Flight, error)
func (r *Repository) ReserveSeats(flightID string, seats int32) (reservationID string, status pb.ReservationStatus, err error)
func (r *Repository) ReleaseReservation(reservationID string) (status pb.ReservationStatus, success bool, err error)