package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database
	err := InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()

	r := mux.NewRouter()

	r.HandleFunc("/analytics", getEvents).Methods("GET")
	r.HandleFunc("/analytics", recordEvent).Methods("POST")

	fmt.Println("Analytics server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
