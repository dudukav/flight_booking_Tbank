FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app

COPY api ./api

COPY flight_service/go.mod flight_service/go.sum ./
RUN go mod download

COPY flight_service/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /flight_service ./cmd/flight/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl netcat-openbsd
WORKDIR /root/
COPY --from=builder /flight_service .
EXPOSE 8081
ENTRYPOINT ["./flight_service"]
