package booking

import (
	"booking_service/internal/client"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	bookingPB "github.com/dudukav/flight_booking_Tbank/api/booking_gen"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
    service     Service
    flightClient *client.FlightClient
}

func NewHandler(service Service, flightClient *client.FlightClient) *Handler {
    return &Handler{
        service:     service,
        flightClient: flightClient,
    }
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /flights", h.SearchFlights)
    mux.HandleFunc("/flights/", h.GetFlight)
    mux.HandleFunc("POST /bookings", h.CreateBooking)
    mux.HandleFunc("/bookings/", h.GetBooking)
    mux.HandleFunc("/bookings/cancel", h.CancelBooking)
    mux.HandleFunc("GET /bookings", h.ListBookings)
}

func (h *Handler) SearchFlights(w http.ResponseWriter, r *http.Request) {
    origin := r.URL.Query().Get("origin")
    destination := r.URL.Query().Get("destination")
    
    if origin == "" || destination == "" {
        http.Error(w, "origin and destination required", http.StatusBadRequest)
        return
    }

    var date *timestamppb.Timestamp
    dateStr := r.URL.Query().Get("date")
    if dateStr != "" {
        t, err := time.Parse(time.DateOnly, dateStr)
        if err != nil {
            http.Error(w, "date must be YYYY-MM-DD", http.StatusBadRequest)
            return
        }
        date = timestamppb.New(t)
    }

    flights, err := h.flightClient.SearchFlights(r.Context(), origin, destination, date)
    if err != nil {
        http.Error(w, fmt.Sprintf("search flights failed: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "flights": flights,
    })
}

func (h *Handler) GetFlight(w http.ResponseWriter, r *http.Request) {
    id := strings.TrimPrefix(r.URL.Path, "/flights/")
    if id == "" {
        http.Error(w, "flight ID required", http.StatusBadRequest)
        return
    }

    flight, err := h.flightClient.GetFlight(r.Context(), id)
    if err != nil {
        http.Error(w, fmt.Sprintf("get flight failed: %v", err), http.StatusInternalServerError)
        return
    }
    if flight == nil {
        http.Error(w, "flight not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "flight": flight,
    })
}

func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
    var req bookingPB.CreateBookingRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    if req.SeatCount <= 0 {
        http.Error(w, "seat_count must be > 0", http.StatusBadRequest)
        return
    }

    bookingModel, err := h.service.CreateBooking(r.Context(), &req)
    if err != nil {
        if st, ok := status.FromError(err); ok {
            switch st.Code() {
            case codes.NotFound:
                http.Error(w, st.Message(), http.StatusNotFound)
            case codes.ResourceExhausted:
                http.Error(w, st.Message(), http.StatusBadRequest)
            default:
                http.Error(w, err.Error(), http.StatusBadRequest)
            }
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "booking_id":  bookingModel.ID.String(),
        "total_price": bookingModel.TotalPrice,
    })
}

func (h *Handler) GetBooking(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    idStr := strings.TrimPrefix(r.URL.Path, "/bookings/")
    id, err := uuid.Parse(idStr)
    if err != nil {
        http.Error(w, "invalid booking ID", http.StatusBadRequest)
        return
    }

    bookingModel, err := h.service.GetBooking(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if bookingModel == nil {
        http.Error(w, "booking not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bookingModel)
}

func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    pathParts := strings.Split(r.URL.Path, "/")
    if len(pathParts) < 3 || pathParts[len(pathParts)-2] != "bookings" || pathParts[len(pathParts)-1] != "cancel" {
        http.Error(w, "path must be /bookings/{id}/cancel", http.StatusBadRequest)
        return
    }

    idStr := pathParts[len(pathParts)-3]
    id, err := uuid.Parse(idStr)
    if err != nil {
        http.Error(w, "invalid booking ID", http.StatusBadRequest)
        return
    }

    if err := h.service.CancelBooking(r.Context(), id); err != nil {
        if st, ok := status.FromError(err); ok {
            switch st.Code() {
            case codes.NotFound:
                http.Error(w, st.Message(), http.StatusNotFound)
            case codes.FailedPrecondition:
                http.Error(w, st.Message(), http.StatusBadRequest)
            default:
                http.Error(w, err.Error(), http.StatusInternalServerError)
            }
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *Handler) ListBookings(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        http.Error(w, "user_id required", http.StatusBadRequest)
        return
    }

    bookings, err := h.service.ListBookings(r.Context(), userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "bookings": bookings,
    })
}
