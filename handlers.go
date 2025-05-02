package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func getEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filter := EventFilter{}

	query := r.URL.Query()

	filter.Service = query.Get("service")
	filter.Event = query.Get("event")
	filter.Path = query.Get("path")
	filter.Referrer = query.Get("referrer")
	filter.UserBrowser = query.Get("browser")
	filter.UserDevice = query.Get("device")

	filter.FromTime = query.Get("from")
	filter.ToTime = query.Get("to")

	filteredEvents, err := GetEvents(filter)
	if err != nil {
		log.Printf("Error getting events: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(filteredEvents)
}

func recordEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var eventReq EventRequest
	if err := json.NewDecoder(r.Body).Decode(&eventReq); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	event := EventRecord{
		Service:     eventReq.Service,
		Event:       eventReq.Event,
		Path:        eventReq.Path,
		Referrer:    eventReq.Referrer,
		UserBrowser: eventReq.UserBrowser,
		UserDevice:  eventReq.UserDevice,
		Timestamp:   eventReq.Timestamp,
		Metadata:    eventReq.Metadata,
	}

	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	// Extract client IP and location
	event.UserIP = getClientIP(r)
	event.UserLocation = getLocationFromIP(event.UserIP)

	err := SaveEvent(event)
	if err != nil {
		log.Printf("Error saving event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Recorded event: %s on page %s from IP %s\n", event.Event, event.Path, event.UserIP)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}
