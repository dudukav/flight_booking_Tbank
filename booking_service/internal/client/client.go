package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/dudukav/flight_booking_Tbank/api/flight_gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FlightClient struct {
    client pb.FlightServiceClient
    conn   *grpc.ClientConn
    circuitBreaker   *CircuitBreaker
}

type RetryPolicy struct {
    MaxAttempts int
    Backoff     []time.Duration
    RetryCodes  []codes.Code
}

func defaultRetryPolicy() *RetryPolicy {
    return &RetryPolicy{
        MaxAttempts: 3,
        Backoff:     []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond},
        RetryCodes:  []codes.Code{codes.Unavailable, codes.DeadlineExceeded},
    }
}

func retryInterceptor(policy *RetryPolicy) grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, 
                cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        
        for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
            err := invoker(ctx, method, req, reply, cc, opts...)
            
            if err == nil || !shouldRetry(status.Code(err), policy.RetryCodes) {
                return err
            }
            
            if attempt < policy.MaxAttempts-1 {
                log.Printf("Retry %d/%d for %s (code: %v): %v", 
                    attempt+1, policy.MaxAttempts, method, status.Code(err), err)
                time.Sleep(policy.Backoff[attempt])
            }
        }
        
        return fmt.Errorf("max retries exceeded")
    }
}

func shouldRetry(code codes.Code, retryCodes []codes.Code) bool {
    for _, rc := range retryCodes {
        if code == rc {
            return true
        }
    }
    return false
}

type apiKeyCreds struct {
    key string
}

func (c *apiKeyCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
    return map[string]string{"x-api-key": c.key}, nil
}

func (c *apiKeyCreds) RequireTransportSecurity() bool { return false }

func NewFlightClient(addr, apiKey string) (*FlightClient, error) {
    creds := &apiKeyCreds{key: apiKey}
    
    maxAttempts := 3
    if attempts := os.Getenv("RETRY_MAX_ATTEMPTS"); attempts != "" {
        fmt.Sscanf(attempts, "%d", &maxAttempts)
    }
    policy := defaultRetryPolicy()
    policy.MaxAttempts = maxAttempts
    
    cbThreshold := 5
    if thresh := os.Getenv("CB_FAILURE_THRESHOLD"); thresh != "" {
        fmt.Sscanf(thresh, "%d", &cbThreshold)
    }
    cbTimeout := 30 * time.Second
    if timeoutStr := os.Getenv("CB_TIMEOUT"); timeoutStr != "" {
        d, err := time.ParseDuration(timeoutStr)
        if err == nil {
            cbTimeout = d
        }
    }
    
    cb := &CircuitBreaker{
        name:      "FlightService",
        state:     CLOSED,
        threshold: cbThreshold,
        timeout:   cbTimeout,
    }
    
    conn, err := grpc.Dial(addr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithPerRPCCredentials(creds),
        grpc.WithChainUnaryInterceptor(
            retryInterceptor(policy),
            circuitBreakerInterceptor(cb),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("dial: %w", err)
    }
    
    return &FlightClient{
        client:         pb.NewFlightServiceClient(conn),
        conn:           conn,
        circuitBreaker: cb,
    }, nil
}

func (f *FlightClient) GetFlight(ctx context.Context, id string) (*pb.Flight, error) {
    resp, err := f.client.GetFlight(ctx, &pb.GetFlightRequest{FlightId: id})
    if err != nil {
        return nil, err
    }
    return resp.Flight, nil
}

func (f *FlightClient) ReserveSeats(ctx context.Context, flightID string, seatCount int32, bookingID string) error {
    _, err := f.client.ReserveSeats(ctx, &pb.ReserveSeatsRequest{
        FlightId:  flightID,
        Seats:     seatCount,
        BookingId: bookingID,
    })
    return err
}

func (f *FlightClient) ReleaseReservation(ctx context.Context, bookingID string) error {
    _, err := f.client.ReleaseReservation(ctx, &pb.ReleaseReservationRequest{
        ReservationId: bookingID,
    })
    return err
}

func (f *FlightClient) SearchFlights(ctx context.Context, origin, destination string, date *timestamppb.Timestamp) ([]*pb.Flight, error) {
    resp, err := f.client.SearchFlights(ctx, &pb.SearchFlightsRequest{
        From:         origin,
        To:           destination,
        DepartureTime: date,
    })
    if err != nil {
        return nil, err
    }
    return resp.Flights, nil
}

func (f *FlightClient) Close() error {
    return f.conn.Close()
}
