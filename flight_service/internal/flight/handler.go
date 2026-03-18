package flight

import (
	"context"
	pb "github.com/dudukav/flight_booking_Tbank/api/flight_gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Handler struct {
    pb.UnimplementedFlightServiceServer
    service *Service
	apiKey string
}

func NewFlightHandler(service *Service, apiKey string) *Handler {
	return &Handler{service: service, apiKey: apiKey}
}

func (h *Handler) auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}
	keys := md.Get("x-api-key")
	if len(keys) == 0 || keys[0] != h.apiKey {
		return status.Error(codes.Unauthenticated, "invalid API key")
	}
	return nil
}

func (h *Handler) SearchFlights(ctx context.Context, req *pb.SearchFlightsRequest) (*pb.SearchFlightsResponse, error) {
	if err := h.auth(ctx); err != nil {
		return nil, err
	}

	return h.service.SearchFlights(ctx, req.From, req.To, req.DepartureTime.AsTime())
}

func (h *Handler) GetFlight(ctx context.Context, req *pb.GetFlightRequest) (*pb.GetFlightResponse, error) {
	if err := h.auth(ctx); err != nil {
		return nil, err
	}

	return h.service.GetFlight(ctx, req.FlightId)
}

func (h *Handler) ReserveSeats(ctx context.Context, req *pb.ReserveSeatsRequest) (*pb.ReserveSeatsResponse, error) {
	if err := h.auth(ctx); err != nil {
		return nil, err
	}

	return h.service.ReserveSeats(ctx, req.FlightId, req.Seats, req.BookingId)
}

func (h *Handler) ReleaseReservation(ctx context.Context, req *pb.ReleaseReservationRequest) (*pb.ReleaseReservationResponse, error) {
	if err := h.auth(ctx); err != nil {
		return nil, err
	}

	return h.service.ReleaseReservation(ctx, req.ReservationId)
}