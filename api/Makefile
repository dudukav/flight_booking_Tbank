.PHONY: generate generate-booking generate-flight clean

generate: generate-booking generate-flight

generate-booking:
	rm -rf booking_gen && \
	mkdir -p booking_gen && \
	protoc --proto_path=proto \
		--go_out=booking_gen --go_opt=paths=source_relative \
		--go-grpc_out=booking_gen --go-grpc_opt=paths=source_relative \
		proto/booking.proto

generate-flight:
	rm -rf flight_gen && \
	mkdir -p flight_gen && \
	protoc --proto_path=proto \
		--go_out=flight_gen --go_opt=paths=source_relative \
		--go-grpc_out=flight_gen --go-grpc_opt=paths=source_relative \
		proto/flight.proto

clean:
	rm -rf booking_gen flight_gen
