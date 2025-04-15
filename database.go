package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
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
	host := getEnv("DB_HOST", defaultHost)
	portStr := getEnv("DB_PORT", strconv.Itoa(defaultPort))
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid DB_PORT value, using default: %d", defaultPort)
		port = defaultPort
	}
	user := getEnv("DB_USER", defaultUser)
	password := getEnv("DB_PASSWORD", defaultPassword)
	dbname := getEnv("DB_NAME", defaultDBName)

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
		user_location VARCHAR(200)
	);
	CREATE INDEX IF NOT EXISTS idx_events_service ON events(service);
	CREATE INDEX IF NOT EXISTS idx_events_event ON events(event);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	`

	_, err := db.Exec(query)
	return err
}

func SaveEvent(event EventRecord) error {
	query := `
	INSERT INTO events (service, event, path, referrer, user_browser, user_device, timestamp, user_ip, user_location)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id
	`

	var id int
	err := db.QueryRow(
		query,
		event.Service,
		event.Event,
		event.Path,
		event.Referrer,
		event.UserBrowser,
		event.UserDevice,
		event.Timestamp,
		event.UserIP,
		event.UserLocation,
	).Scan(&id)

	return err
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

func (qb *QueryBuilder) Build(orderBy string, limit int) (string, []interface{}) {
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

	return query, qb.args
}

func GetEvents(filter EventFilter) ([]EventRecord, error) {
	qb := NewQueryBuilder(`
		SELECT service, event, path, referrer, user_browser, user_device, timestamp, user_ip, user_location
		FROM events
		WHERE 1=1
	`)

	qb.AddFilters(filter)

	query, args := qb.Build("timestamp DESC", 1000)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventRecord
	for rows.Next() {
		var e EventRecord
		var timestamp time.Time

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
		)
		if err != nil {
			return nil, err
		}

		e.Timestamp = timestamp.Format(time.RFC3339Nano)
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
