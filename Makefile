.PHONY: proto migrate up dev down clean

proto:
	docker run --rm -v $(PWD):/workspace -w /workspace/api \
		golang:1.24-alpine sh -c "\
			apk add --no-cache protobuf-dev protoc git make && \
			go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
			go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
			make generate \
		"

migrate:
	make proto && \
	docker-compose up postgres-booking postgres-flight && \
	docker-compose run --rm migrate-booking && \
	docker-compose run --rm migrate-flight

up:
	make migrate && docker-compose up booking-service flight-service -d

dev:
	make proto && docker-compose up -d

down:
	docker-compose down -v
