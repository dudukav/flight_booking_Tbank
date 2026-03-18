package main

import (
	"ariga.io/atlas-provider-gorm/gormschema"
	"flight_booking_Tbank/internal/flight"
	"fmt"
	"io"
	"os"
)

func main() {
	stmts, err := gormschema.New("postgres", gormschema.WithModelPosition(map[any]string{
		&flight.Flight{}:          "/Users/duduka/soa/flight_booking_Tbank/flight_service/internal/flight/models.go:9",
		&flight.SeatReservation{}: "/Users/duduka/soa/flight_booking_Tbank/flight_service/internal/flight/models.go:23",
	})).Load(
		&flight.Flight{},
		&flight.SeatReservation{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
