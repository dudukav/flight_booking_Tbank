FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app

COPY api ./api               
COPY booking_service/go.mod booking_service/go.sum ./
RUN go mod download

COPY booking_service/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /booking_service ./cmd/booking/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /booking_service .
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost:8080/health || exit 1
ENTRYPOINT ["./booking_service"]
