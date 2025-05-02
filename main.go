package main

import (
	"fmt"
	"log"
	"net/http"

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

	// Create a CORS middleware handler
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Use the CORS middleware with the router
	handler := c.Handler(r)

	fmt.Println("Analytics server is running")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
