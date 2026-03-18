package booking

import (
	"booking_service/internal/client"
	"context"
	"fmt"

	bookingPB "github.com/dudukav/flight_booking_Tbank/api/booking_gen"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
    CreateBooking(ctx context.Context, req *bookingPB.CreateBookingRequest) (*Booking, error)
    GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error)
    CancelBooking(ctx context.Context, id uuid.UUID) error
    ListBookings(ctx context.Context, userID string) ([]*Booking, error)
}

type service struct {
    repo         Repository
    flightClient *client.FlightClient
}

func NewService(repo Repository, flightClient *client.FlightClient) Service {
    return &service{
        repo:         repo,
        flightClient: flightClient,
    }
}

func (s *service) CreateBooking(ctx context.Context, req *bookingPB.CreateBookingRequest) (*Booking, error) {
    flightID, err := uuid.Parse(req.FlightId)
    if err != nil {
        return nil, fmt.Errorf("invalid flight_id: %w", err)
    }

    flight, err := s.flightClient.GetFlight(ctx, req.FlightId)
    if err != nil {
        return nil, fmt.Errorf("get flight: %w", err)
    }
    if flight == nil {
        return nil, status.Error(codes.NotFound, "flight not found")
    }

    bookingID := uuid.New().String()
    err = s.flightClient.ReserveSeats(ctx, req.FlightId, req.SeatCount, bookingID)
    if err != nil {
        return nil, fmt.Errorf("reserve seats: %w", err)
    }

    totalPrice := flight.Price * float32(req.SeatCount)
    booking := &Booking{
        ID:             uuid.MustParse(bookingID),
        UserID:         req.UserId,
        FlightID:       flightID,
        PassengerName:  req.PassengerName,
        PassengerEmail: req.PassengerEmail,
        SeatCount:      req.SeatCount,
        TotalPrice:     totalPrice,
        Status:         "CONFIRMED",
    }

    if err := s.repo.Create(ctx, booking); err != nil {
        s.flightClient.ReleaseReservation(ctx, bookingID)
        return nil, fmt.Errorf("create booking: %w", err)
    }

    return booking, nil
}


func (s *service) GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error) {
    return s.repo.GetByID(ctx, id)
}

func (s *service) CancelBooking(ctx context.Context, id uuid.UUID) error {
    bookingID := id.String()
    
    booking, err := s.repo.GetByID(ctx, id)
    if err != nil || booking == nil {
        return status.Error(codes.NotFound, "booking not found")
    }
    if booking.Status != "CONFIRMED" {
        return status.Error(codes.FailedPrecondition, "only CONFIRMED bookings can be cancelled")
    }

    err = s.flightClient.ReleaseReservation(ctx, bookingID)
    if err != nil {
        return fmt.Errorf("release reservation: %w", err)
    }

    return s.repo.UpdateStatus(ctx, id, "CANCELLED")
}

func (s *service) ListBookings(ctx context.Context, userID string) ([]*Booking, error) {
    return s.repo.ListByUserID(ctx, userID)
}