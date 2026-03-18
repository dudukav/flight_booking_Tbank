package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"flight_booking_Tbank/internal/config"
	"flight_booking_Tbank/internal/flight"

	pb "github.com/dudukav/flight_booking_Tbank/api/flight_gen"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load config:", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("postgres connect:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("sql db:", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("postgres ping failed:", err)
	}
	log.Println("Postgres OK")

	if os.Getenv("ENV") == "development" {
		log.Println("AutoMigrate (dev)")
		db.AutoMigrate(&flight.Flight{}, &flight.SeatReservation{})
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis ping failed:", err)
	}
	log.Println("Redis OK")

	repo := flight.NewRepository(db)
	service := flight.NewService(repo, rdb, db)
	handler := flight.NewFlightHandler(service, cfg.APIKey)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("listen failed:", err)
	}

	srv := grpc.NewServer()
	pb.RegisterFlightServiceServer(srv, handler)
	grpc_health_v1.RegisterHealthServer(srv, &HealthServer{})

	log.Printf("🚀 Flight Service :%s", cfg.GRPCPort)

	ctx, cancel := signal.NotifyContext(context.Background(), 
		os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatal("grpc serve:", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal...")

	srv.GracefulStop()
	log.Println("gRPC stopped")

	sqlDB.Close()
	rdb.Close()
	log.Println("Connections closed")
}

type HealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}
