package flight

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	pb "flight_booking_Tbank/flight_gen"
)

type Service struct {
	repo *Repository
}

func NewService(rdb *redis.Client) *Service {
	return &Service{
		repo: NewRepository(rdb),
	}
}

func (s *Service) SearchFlights(ctx context.Context, from, to string, departureTime time.Time) (*pb.SearchFlightsResponse, error)
func (s *Service) GetFlight(ctx context.Context, flightID string) (*pb.GetFlightResponse, error)
func (s *Service) ReserveSeats(ctx context.Context, flightID string, seats int32) (*pb.ReserveSeatsResponse, error)
func (s *Service) ReleaseReservation(ctx context.Context, reservationID string) (*pb.ReleaseReservationResponse, error)