package flight

import (
	"context"
	"time"

	pb "flight_booking_Tbank/flight_gen"
)

type Service struct {
	repo *Repository
}

func (s *Service) SearchFlights(ctx context.Context, from, to string, departureTime time.Time) (*pb.SearchFlightsResponse, error) {
	flights, err := s.repo.SearchFlights(from, to, departureTime)
	if err != nil {
		return nil, err
	}

	return &pb.SearchFlightsResponse{
		Flights: flights,
	}, nil
}

func (s *Service) GetFlight(ctx context.Context, flightID string) (*pb.GetFlightResponse, error) {
	flight, err := s.repo.GetFlight(flightID)
	if err != nil {
		return nil, err
	}

	return &pb.GetFlightResponse{Flight: flight}, nil
}

func (s *Service) ReserveSeats(ctx context.Context, flightID string, seats int32) (*pb.ReserveSeatsResponse, error) {
	reservationID, reservationStatus, err := s.repo.ReserveSeats(flightID, seats)
	if err != nil {
		return nil, err
	}

	return &pb.ReserveSeatsResponse{
		ReservationId: reservationID, 
		Status: reservationStatus,
		}, nil
}

func (s *Service) ReleaseReservation(ctx context.Context, reservationID string) (*pb.ReleaseReservationResponse, error) {
	status, success, err := s.repo.ReleaseReservation(reservationID)
	if err != nil {
		return nil, err
	}

	return &pb.ReleaseReservationResponse{
		ReservationId: reservationID,
		Status:        status,
		Success:       success,
	}, nil
}