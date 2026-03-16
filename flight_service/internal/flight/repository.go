package flight

import (
	"context"
	"fmt"
	"time"

	pb "flight_booking_Tbank/flight_gen"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	redisClient  *redis.Client
	db			 *gorm.DB
}

func NewRepository(db *gorm.DB, rdb *redis.Client) *Repository {
    return &Repository{
		db: db,
        redisClient: rdb,
    }
}

func (r *Repository) SearchFlights(from, to string, departureTime time.Time) ([]*pb.Flight, error) {
	var result []*pb.Flight

	var flights []Flight
	if err := r.db.Where("origin_airport = ? AND destination_airport = ? AND departure_date = ?", from, to, departureTime).Find(&flights).Error; err != nil {
		return nil, err
	}

	for _, flight := range flights {
		var reservedSeats int64
		row := r.db.Model(&SeatReservation{}).
			Where("flight_id = ? AND status = ?", flight.ID, "ACTIVE").
			Select("COALESCE(SUM(reserved_seats), 0)").Row()
		if err := row.Scan(&reservedSeats); err != nil {
			return nil, err
		}

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

	return result, nil
}

func (r *Repository) GetFlight(flightID string) (*pb.Flight, error) {
	var flight Flight
	if err := r.db.First(&flight, "id = ?", flightID).Error; err != nil {
		return nil, err
	}

	var reservedSeats int64
	row := r.db.Model(&SeatReservation{}).
		Where("flight_id = ? AND status = ?", flight.ID, "ACTIVE").
		Select("COALESCE(SUM(reserved_seats), 0)").Row()
	if err := row.Scan(&reservedSeats); err != nil {
		return nil, err
	}

	availableSeats := int32(flight.TotalSeats - reservedSeats)

	result := &pb.Flight{
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

	return result, nil
}

func (r *Repository) ReserveSeats(flightID string, seats int32) (reservationID string, status pb.ReservationStatus, err error) {
	ctx := context.Background()

	tx := r.db.Begin()
	if tx.Error != nil {
		return "", pb.ReservationStatus_RESERVATION_UNKNOWN, tx.Error
	}

	var flight Flight
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&flight, "id = ?", flightID).Error; err != nil {
		tx.Rollback()
		return "", pb.ReservationStatus_RESERVATION_UNKNOWN, err
	}

	var reservedSeats int64
	row := tx.Model(&SeatReservation{}).
		Where("flight_id = ? AND status = ?", flightID, "ACTIVE").
		Select("COALESCE(SUM(reserved_seats),0)").Row()
	if err := row.Scan(&reservedSeats); err != nil {
		tx.Rollback()
		return "", pb.ReservationStatus_RESERVATION_UNKNOWN, err
	}

	availableSeats := int32(flight.TotalSeats - reservedSeats)
	if seats > availableSeats {
		tx.Rollback()
		return "", pb.ReservationStatus_RESERVATION_UNKNOWN, fmt.Errorf("not enough seats")
	}

	reservationID = uuid.New().String()
	reservation := SeatReservation{
		ID:            uuid.MustParse(reservationID),
		FlightID:      uuid.MustParse(flightID),
		BookingID:     uuid.New(),
		ReservedSeats: int64(seats),
		Status:        "ACTIVE",
	}

	if err := tx.Create(&reservation).Error; err != nil {
		tx.Rollback()
		return "", pb.ReservationStatus_RESERVATION_UNKNOWN, err
	}

	if err := tx.Commit().Error; err != nil {
		return "", pb.ReservationStatus_RESERVATION_UNKNOWN, err
	}

	key := fmt.Sprintf("flight:%s:available_seats", flightID)
	_, _ = r.redisClient.DecrBy(ctx, key, int64(seats)).Result()

	return reservationID, pb.ReservationStatus_RESERVED, nil
}

func (r *Repository) ReleaseReservation(reservationID string) (status pb.ReservationStatus, success bool, err error) {
	ctx := context.Background()
	tx := r.db.Begin()
	if tx.Error != nil {
		return pb.ReservationStatus_RESERVATION_UNKNOWN, false, tx.Error
	}

	var reservation SeatReservation
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&reservation, "id = ?", reservationID).Error; err != nil {
		tx.Rollback()
		return pb.ReservationStatus_RESERVATION_UNKNOWN, false, err
	}

	if reservation.Status != "ACTIVE" {
		tx.Rollback()
		return pb.ReservationStatus_RESERVATION_UNKNOWN, false, fmt.Errorf("reservation not active")
	}

	reservation.Status = "RELEASED"
	if err := tx.Save(&reservation).Error; err != nil {
		tx.Rollback()
		return pb.ReservationStatus_RESERVATION_UNKNOWN, false, err
	}

	if err := tx.Commit().Error; err != nil {
		return pb.ReservationStatus_RESERVATION_UNKNOWN, false, err
	}

	key := fmt.Sprintf("flight:%s:available_seats", reservation.FlightID)
	_, _ = r.redisClient.IncrBy(ctx, key, int64(reservation.ReservedSeats)).Result()

	return pb.ReservationStatus_RESERVATION_CANCELLED, true, nil
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