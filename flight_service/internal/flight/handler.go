package flight

import ( 
	"context"
	pb "flight_booking_Tbank/flight_gen"
)



type Handler struct {
    pb.UnimplementedFlightServiceServer
    service *Service
}

func (h *Handler) ReserveSeats(ctx context.Context, req *pb.ReserveSeatsRequest) (*pb.ReserveSeatsResponse, error)