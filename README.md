# Analytics Service

A simple analytics service that records and retrieves events using PostgreSQL for storage.

## Requirements

- Go 1.22+
- PostgreSQL (or Docker)

## Setup

1. Start PostgreSQL database:

```bash
docker run --name analytics-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=analytics -p 5432:5432 -d postgres:16
```

2. Download dependencies:

```bash
go get github.com/gorilla/mux
go get github.com/lib/pq
```

3. Run the application:

```bash
go run .
```

The application will be available at http://localhost:8080.

## API

### Record an event

```bash
POST /analytics
```

Example:

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "service": "my-app",
  "event": "page-view",
  "path": "/home",
  "referrer": "https://google.com",
  "user_browser": "Chrome",
  "user_device": "Desktop"
}' http://localhost:8080/analytics
```

### Get events

```bash
GET /analytics
```

Supported query parameters:
- `service`: Filter by service name
- `event`: Filter by event type
- `path`: Filter by page path (partial match)
- `referrer`: Filter by referrer
- `browser`: Filter by user browser
- `device`: Filter by user device
- `from`: Filter by timestamp (RFC3339Nano format)
- `to`: Filter by timestamp (RFC3339Nano format)

Example:

```bash
curl "http://localhost:8080/analytics?service=my-app&event=page-view"
```

## Database Configuration

Database connection settings are defined in `database.go`:

```go
const (
    host     = "localhost"
    port     = 5432
    user     = "postgres"
    password = "postgres"
    dbname   = "analytics"
)
```

Modify these constants to match your PostgreSQL configuration if needed.

## Configuration

### CORS Configuration

By default, the server allows requests from all origins. To restrict access to specific origins, set the `ALLOWED_ORIGINS` environment variable:

```
# Allow requests from multiple domains
export ALLOWED_ORIGINS=https://example.com,https://app.example.com

# Or for a single domain
export ALLOWED_ORIGINS=https://example.com
```

When running with Docker, you can set the environment variable in your docker-compose file or when running the container:

```
docker run -e ALLOWED_ORIGINS=https://example.com -p 8080:8080 analytics-server 