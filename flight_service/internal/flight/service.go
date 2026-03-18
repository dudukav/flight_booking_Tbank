package flight

import (
    "context"
    "fmt"
    "log"
    "time"
    
    pb "github.com/dudukav/flight_booking_Tbank/api/flight_gen"
    
    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type Service struct {
    repo        Repository
    redisClient *redis.Client
    db          *gorm.DB
}

func NewService(repo Repository, rdb *redis.Client, db *gorm.DB) *Service {
    return &Service{
        repo:        repo,
        redisClient: rdb,
        db:          db,
    }
}

func (s *Service) SearchFlights(ctx context.Context, from, to string, departureTime time.Time) (*pb.SearchFlightsResponse, error) {
    flights, err := s.repo.ListByRoute(ctx, from, to, departureTime)
    if err != nil {
        return nil, err
    }
    
    var result []*pb.Flight
    for _, flight := range flights {
        reservedSeats, _ := s.repo.SumActiveSeatsByFlight(ctx, flight.ID)
        availableSeats := int32(flight.TotalSeats - reservedSeats)
        
        result = append(result, &pb.Flight{
            Id:                 flight.ID.String(),
            FlightNumber:       flight.FlightNumber,
            Airline:            flight.Airline,
            OriginAirport:      flight.OriginAirport,
            DestinationAirport: flight.DestinationAirport,
            DepartureTime:      timestamppb.New(flight.DepartureTime),
            ArrivalTime:        timestamppb.New(flight.ArrivalTime),
            TotalSeats:         int32(flight.TotalSeats),
            AvailableSeats:     availableSeats,
            Price:              float32(flight.Price),
            Status:             toProtoFlightStatus(flight.Status),
        })
    }
    
    return &pb.SearchFlightsResponse{Flights: result}, nil
}

func (s *Service) GetFlight(ctx context.Context, flightID string) (*pb.GetFlightResponse, error) {
    id := uuid.MustParse(flightID)
    flight, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, status.Error(codes.NotFound, "flight not found")
    }
    
    reservedSeats, _ := s.repo.SumActiveSeatsByFlight(ctx, id)
    availableSeats := int32(flight.TotalSeats - reservedSeats)
    
    pbFlight := &pb.Flight{
        Id:                 flight.ID.String(),
        FlightNumber:       flight.FlightNumber,
        Airline:            flight.Airline,
        OriginAirport:      flight.OriginAirport,
        DestinationAirport: flight.DestinationAirport,
        DepartureTime:      timestamppb.New(flight.DepartureTime),
        ArrivalTime:        timestamppb.New(flight.ArrivalTime),
        TotalSeats:         int32(flight.TotalSeats),
        AvailableSeats:     availableSeats,
        Price:              float32(flight.Price),
        Status:             toProtoFlightStatus(flight.Status),
    }
    
    return &pb.GetFlightResponse{Flight: pbFlight}, nil
}

func (s *Service) ReserveSeats(ctx context.Context, flightIDStr string, seats int32, bookingIDStr string) (*pb.ReserveSeatsResponse, error) {
    flightID := uuid.MustParse(flightIDStr)
    
    err := s.db.Transaction(func(tx *gorm.DB) error {
        var flight Flight
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&flight, "id = ?", flightID).Error; err != nil {
            return status.Error(codes.NotFound, "flight not found")
        }
      
        reservedSeats, err := s.repo.SumActiveSeatsByFlight(tx.Statement.Context, flightID)
        if err != nil {
            return err
        }
        
        availableSeats := int32(flight.TotalSeats - reservedSeats)
        if seats > availableSeats {
            return status.Error(codes.ResourceExhausted, "not enough seats")
        }
        
        reservationID := uuid.New().String()
        reservation := SeatReservation{
            ID:            uuid.MustParse(reservationID),
            FlightID:      flightID,
            BookingID:     uuid.MustParse(bookingIDStr),
            ReservedSeats: int64(seats),
            Status:        "ACTIVE",
        }
        
        return s.repo.CreateReservation(tx.Statement.Context, &reservation)
    })
    
    if err != nil {
        return nil, err
    }
    
    ctxRedis := context.Background()
    key := fmt.Sprintf("flight:%s", flightIDStr)
    s.redisClient.Del(ctxRedis, key)
    
    return &pb.ReserveSeatsResponse{
        ReservationId: uuid.New().String(),
        Status:        pb.ReservationStatus_RESERVED,
    }, nil
}


func (s *Service) ReleaseReservation(ctx context.Context, reservationIDStr string) (*pb.ReleaseReservationResponse, error) {
    id := uuid.MustParse(reservationIDStr)
    
    err := s.db.Transaction(func(tx *gorm.DB) error {
        var reservation SeatReservation
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&reservation, "id = ?", id).Error; err != nil {
            return status.Error(codes.NotFound, "reservation not found")
        }
        
        if reservation.Status != "ACTIVE" {
            return status.Error(codes.FailedPrecondition, "reservation not active")
        }

        return s.repo.UpdateReservationStatus(tx.Statement.Context, id, "RELEASED")
    })
    
    if err != nil {
        return nil, err
    }

    ctxRedis := context.Background()
    key := fmt.Sprintf("flight:%s", reservationIDStr)
    s.redisClient.Del(ctxRedis, key)
    
    log.Printf("Released reservation %s", reservationIDStr)
    
    return &pb.ReleaseReservationResponse{
        ReservationId: reservationIDStr,
        Status:        pb.ReservationStatus_RESERVATION_CANCELLED,
        Success:       true,
    }, nil
}



func toProtoFlightStatus(status string) pb.FlightStatus {
    switch status {
    case "SCHEDULED":
        return pb.FlightStatus_SCHEDULED
    case "DEPARTED":
        return pb.FlightStatus_DEPARTED
    case "CANCELLED":
        return pb.FlightStatus_FLIGHT_CANCELLED
    case "COMPLETED":
        return pb.FlightStatus_FLIGHT_COMPLETED
    default:
        return pb.FlightStatus_FLIGHT_UNKNOWN
    }
}
