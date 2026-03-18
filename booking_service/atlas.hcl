data "external_schema" "app" {
  program = ["go", "run", "-mod=mod", "./cmd/atlas"]
}

env "docker" {
  src = data.external_schema.app.url
  dev = "postgres://postgres:password@postgres-booking:5432/booking?sslmode=disable"
  url = "postgres://postgres:password@postgres-booking:5432/booking?sslmode=disable"
  
  migration {
    dir = "file://migrations"
  }
}