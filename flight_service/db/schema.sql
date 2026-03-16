-- Это все сгенерировано по коду в dbdiagram.io

CREATE TABLE "flights" (
  "id" uuid PRIMARY KEY,
  "flight_number" varchar,
  "departure_date" date,
  "airline" varchar,
  "origin_airport" char(3),
  "destination_airport" char(3),
  "departure_time" timestamp,
  "arrival_time" timestamp,
  "total_seats" int NOT NULL,
  "price" decimal(10,2) NOT NULL,
  "status" varchar NOT NULL
);

CREATE TABLE "seat_reservations" (
  "id" uuid PRIMARY KEY,
  "flight_id" uuid NOT NULL,
  "booking_id" uuid NOT NULL,
  "reserved_seats" int NOT NULL,
  "status" varchar NOT NULL
);

CREATE UNIQUE INDEX ON "flights" ("flight_number", "departure_date");

CREATE UNIQUE INDEX ON "seat_reservations" ("booking_id");

COMMENT ON TABLE "flights" IS 'Flight information. total_seats and price must be positive.';

COMMENT ON COLUMN "flights"."total_seats" IS 'must be > 0';

COMMENT ON COLUMN "flights"."price" IS 'must be > 0';

COMMENT ON COLUMN "flights"."status" IS 'SCHEDULED, DEPARTED, CANCELLED, COMPLETED';

COMMENT ON TABLE "seat_reservations" IS 'One reservation per booking. booking_id refers to external Booking Service.';

COMMENT ON COLUMN "seat_reservations"."reserved_seats" IS 'must be >= 0';

COMMENT ON COLUMN "seat_reservations"."status" IS 'ACTIVE, RELEASED, EXPIRED';

ALTER TABLE "seat_reservations" ADD FOREIGN KEY ("flight_id") REFERENCES "flights" ("id") DEFERRABLE INITIALLY IMMEDIATE;
