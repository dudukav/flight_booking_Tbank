package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"flight_booking_Tbank/internal/flight"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&flight.Flight{},
		&flight.SeatReservation{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
