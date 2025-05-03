package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	err := InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()

	r := mux.NewRouter()

	r.HandleFunc("/api", getEvents).Methods("GET")
	r.HandleFunc("/api", recordEvent).Methods("POST")

	// Get allowed origins from environment variable
	allowedOrigins := getAllowedOrigins()

	// Create a CORS middleware handler
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Use the CORS middleware with the router
	handler := c.Handler(r)

	fmt.Println("Analytics server is running")
	log.Printf("Allowed origins: %v", allowedOrigins)
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// getAllowedOrigins returns a list of allowed origins from the ALLOWED_ORIGINS
// environment variable. If not set, defaults to ["*"] (all origins)
func getAllowedOrigins() []string {
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		log.Println("Warning: ALLOWED_ORIGINS not set, allowing all origins")
		return []string{"*"}
	}
	return strings.Split(origins, ",")
}
