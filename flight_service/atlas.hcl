data "external_schema" "app" {
  program = ["go", "run", "-mod=mod", "./cmd/atlas"]
}

env "docker" {
  src = data.external_schema.app.url
  dev = "postgres://postgres:password@postgres-flight:5432/flight?sslmode=disable"
  url = "postgres://postgres:password@postgres-flight:5432/flight?sslmode=disable"
  
  migration {
    dir = "file://migrations"
  }
}