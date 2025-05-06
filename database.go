package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

var (
	defaultHost     = "localhost"
	defaultPort     = 5432
	defaultUser     = "postgres"
	defaultPassword = "postgres"
	defaultDBName   = "analytics"
)

var db *sql.DB

func InitDB() error {
	host := getEnv("PGHOST", defaultHost)
	portStr := getEnv("PGPORT", strconv.Itoa(defaultPort))
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid PGPORT value, using default: %d", defaultPort)
		port = defaultPort
	}
	user := getEnv("PGUSER", defaultUser)
	password := getEnv("PGPASSWORD", defaultPassword)
	dbname := getEnv("POSTGRES_DB", defaultDBName)

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	err = createTablesIfNotExist()
	if err != nil {
		return err
	}

	log.Println("Successfully connected to the database")
	return nil
}

func createTablesIfNotExist() error {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id SERIAL PRIMARY KEY,
		service VARCHAR(100) NOT NULL,
		event VARCHAR(100) NOT NULL,
		path TEXT NOT NULL,
		referrer TEXT,
		user_browser VARCHAR(200),
		user_device VARCHAR(200),
		timestamp TIMESTAMPTZ NOT NULL,
		user_ip VARCHAR(45),
		user_location VARCHAR(200),
		metadata JSONB
	);
	CREATE INDEX IF NOT EXISTS idx_events_service ON events(service);
	CREATE INDEX IF NOT EXISTS idx_events_event ON events(event);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	`

	_, err := db.Exec(query)
	return err
}

func SaveEvent(event EventRecord) error {
	var query string
	var args []interface{}

	// Base params without metadata
	args = append(args,
		event.Service,
		event.Event,
		event.Path,
		event.Referrer,
		event.UserBrowser,
		event.UserDevice,
		event.Timestamp,
		event.UserIP,
		event.UserLocation)

	// Different query format based on metadata presence
	if event.Metadata != nil && len(event.Metadata) > 0 {
		// Use a text value that will be cast to JSONB by PostgreSQL
		jsonData, err := json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %v", err)
		}

		query = `
		INSERT INTO events (service, event, path, referrer, user_browser, user_device, timestamp, user_ip, user_location, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
		RETURNING id
		`
		args = append(args, string(jsonData))
	} else {
		query = `
		INSERT INTO events (service, event, path, referrer, user_browser, user_device, timestamp, user_ip, user_location, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL)
		RETURNING id
		`
	}

	var id int
	err := db.QueryRow(query, args...).Scan(&id)
	if err != nil {
		return fmt.Errorf("database insert error: %v", err)
	}

	return nil
}

func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery:  baseQuery,
		conditions: []string{},
		args:       []interface{}{},
	}
}

func (qb *QueryBuilder) AddWhere(field string, value interface{}) *QueryBuilder {
	return qb.AddCondition(field, "=", value)
}

func (qb *QueryBuilder) AddCondition(field, operator string, value interface{}) *QueryBuilder {
	if value == nil {
		return qb
	}
	if strVal, ok := value.(string); ok && strVal == "" {
		return qb
	}

	paramNum := len(qb.args) + 1
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s %s $%d", field, operator, paramNum))
	qb.args = append(qb.args, value)
	return qb
}

func (qb *QueryBuilder) AddFilters(filter EventFilter) *QueryBuilder {
	qb.AddWhere("service", filter.Service)
	qb.AddWhere("event", filter.Event)
	qb.AddWhere("referrer", filter.Referrer)
	qb.AddWhere("user_browser", filter.UserBrowser)
	qb.AddWhere("user_device", filter.UserDevice)

	if filter.Path != "" {
		qb.AddCondition("path", "LIKE", "%"+filter.Path+"%")
	}

	if filter.FromTime != "" {
		t, err := time.Parse(time.RFC3339Nano, filter.FromTime)
		if err == nil {
			qb.AddCondition("timestamp", ">=", t)
		}
	}

	if filter.ToTime != "" {
		t, err := time.Parse(time.RFC3339Nano, filter.ToTime)
		if err == nil {
			qb.AddCondition("timestamp", "<=", t)
		}
	}

	return qb
}

func (qb *QueryBuilder) Build(orderBy string, limit int, offset int) (string, []interface{}) {
	query := qb.baseQuery

	for _, condition := range qb.conditions {
		query += " AND " + condition
	}

	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return query, qb.args
}

func GetEvents(filter EventFilter) ([]EventRecord, error) {
	qb := NewQueryBuilder(`
		SELECT service, event, path, referrer, user_browser, user_device, timestamp, user_ip, user_location, metadata
		FROM events
		WHERE 1=1
	`)

	qb.AddFilters(filter)

	// Get total count first to see if we might need pagination
	countQuery := "SELECT COUNT(*) FROM events WHERE 1=1"
	for _, condition := range qb.conditions {
		countQuery += " AND " + condition
	}

	var totalCount int
	err := db.QueryRow(countQuery, qb.args...).Scan(&totalCount)
	if err != nil {
		return nil, err
	}

	// The maximum number of records to return in one query
	const maxLimit = 1000

	// If we have a time range query and more records than our limit
	if totalCount > maxLimit && filter.FromTime != "" && filter.ToTime != "" {
		// We need to fetch data in chunks across the time range
		return getEventsWithTimePagination(filter, totalCount, maxLimit)
	}

	// Regular query with a limit
	limit := maxLimit
	if totalCount > limit {
		log.Printf("WARNING: Query would return %d records, but we're limited to %d. Consider using a more specific time range.",
			totalCount, limit)
	}

	query, args := qb.Build("timestamp DESC", limit, 0)

	return executeEventQuery(query, args)
}

// getEventsWithTimePagination fetches events in time-based chunks when the total count exceeds our limit
func getEventsWithTimePagination(filter EventFilter, totalCount, maxLimit int) ([]EventRecord, error) {
	fromTime, err := time.Parse(time.RFC3339Nano, filter.FromTime)
	if err != nil {
		return nil, fmt.Errorf("invalid from time format: %v", err)
	}

	toTime, err := time.Parse(time.RFC3339Nano, filter.ToTime)
	if err != nil {
		return nil, fmt.Errorf("invalid to time format: %v", err)
	}

	// Calculate how many chunks we need
	chunks := (totalCount + maxLimit - 1) / maxLimit
	if chunks > 10 {
		// Limit to a reasonable number to avoid excessive database load
		chunks = 10
		log.Printf("WARNING: Large result set (%d records). Limiting to %d chunks of %d records each.",
			totalCount, chunks, maxLimit)
	}

	// Time range in seconds
	timeRange := toTime.Sub(fromTime).Seconds()

	// Size of each chunk in seconds
	chunkSize := timeRange / float64(chunks)

	var allEvents []EventRecord

	// Fetch each chunk
	for i := 0; i < chunks; i++ {
		// Calculate time boundaries for this chunk
		chunkStart := fromTime.Add(time.Duration(float64(i) * chunkSize * float64(time.Second)))
		chunkEnd := fromTime.Add(time.Duration(float64(i+1) * chunkSize * float64(time.Second)))

		// Make sure we don't exceed the original time range
		if chunkEnd.After(toTime) {
			chunkEnd = toTime
		}

		// Create a filter for this chunk
		chunkFilter := filter
		chunkFilter.FromTime = chunkStart.Format(time.RFC3339Nano)
		chunkFilter.ToTime = chunkEnd.Format(time.RFC3339Nano)

		qb := NewQueryBuilder(`
			SELECT service, event, path, referrer, user_browser, user_device, timestamp, user_ip, user_location, metadata
			FROM events
			WHERE 1=1
		`)

		qb.AddFilters(chunkFilter)

		query, args := qb.Build("timestamp DESC", maxLimit, 0)

		events, err := executeEventQuery(query, args)
		if err != nil {
			return nil, err
		}

		allEvents = append(allEvents, events...)

		// Log progress
		log.Printf("Fetched chunk %d/%d: %d events from %s to %s",
			i+1, chunks, len(events), chunkStart.Format(time.RFC3339), chunkEnd.Format(time.RFC3339))
	}

	// Sort the combined results
	// We're using timestamp DESC but we've combined results from different queries
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Timestamp > allEvents[j].Timestamp
	})

	// Ensure we're still respecting the maxLimit
	if len(allEvents) > maxLimit {
		log.Printf("Truncating results from %d to %d records", len(allEvents), maxLimit)
		allEvents = allEvents[:maxLimit]
	}

	return allEvents, nil
}

// executeEventQuery executes a query and processes the rows into EventRecord structs
func executeEventQuery(query string, args []interface{}) ([]EventRecord, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventRecord
	for rows.Next() {
		var e EventRecord
		var timestamp time.Time
		var metadataJSON []byte

		// Initialize empty metadata map
		e.Metadata = make(map[string]interface{})

		err := rows.Scan(
			&e.Service,
			&e.Event,
			&e.Path,
			&e.Referrer,
			&e.UserBrowser,
			&e.UserDevice,
			&timestamp,
			&e.UserIP,
			&e.UserLocation,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		e.Timestamp = timestamp.Format(time.RFC3339Nano)

		// Parse metadata if it exists and is not null
		if metadataJSON != nil && len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &e.Metadata); err != nil {
				log.Printf("Error unmarshaling metadata for event %s: %v", e.Event, err)
				// Continue with the event even if metadata parsing fails
				e.Metadata = map[string]interface{}{"error": "Failed to parse metadata"}
			}
		}

		events = append(events, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// Helper function to get environment variables with fallback to default values
func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}
