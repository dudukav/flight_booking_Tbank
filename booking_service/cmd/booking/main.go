package main

import (
    "context"
    "log"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "booking_service/internal/booking"
    "booking_service/internal/client"
    "booking_service/internal/config"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    cfg := config.Load()
    log.Printf("Starting booking service (env: %s, version: %s)", cfg.App.Env, cfg.App.Version)

    db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
    if err != nil {
        log.Fatal("failed to connect to database:", err)
    }

    repo := booking.NewRepository(db)

    flightClient, err := client.NewFlightClient(cfg.FlightService.Addr, cfg.FlightService.APIKey)
    if err != nil {
        log.Fatal("failed to create flight client:", err)
    }

    svc := booking.NewService(repo, flightClient)
    handler := booking.NewHandler(svc, flightClient)

    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)

    srv := &http.Server{
        Addr:         ":" + cfg.HTTP.Port,
        Handler:      mux,
        ReadTimeout:  cfg.HTTP.ReadTimeout,
        WriteTimeout: cfg.HTTP.WriteTimeout,
    }

    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    go func() {
        log.Printf("server starting on :%s", cfg.HTTP.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("server failed:", err)
        }
    }()

    <-ctx.Done()
    log.Println("shutting down server...")

    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer shutdownCancel()
    
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Fatal("server shutdown failed:", err)
    }
    log.Println("server stopped")
}
