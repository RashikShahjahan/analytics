package main

import (
	"net/http"
	"strings"
)

func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header first (common behind proxies)
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// The first IP in the list is the client IP
		return strings.Split(forwardedFor, ",")[0]
	}

	// Fall back to RemoteAddr if X-Forwarded-For is not present
	// RemoteAddr includes the port, so we need to strip it
	remoteAddr := r.RemoteAddr
	if strings.Contains(remoteAddr, ":") {
		return strings.Split(remoteAddr, ":")[0]
	}

	return remoteAddr
}

func getLocationFromIP(ip string) string {
	// In a real application, you would integrate with a geolocation API service
	// For this example, we'll return a placeholder
	// TODO: Integrate with a proper geolocation service
	return "Unknown" // Placeholder for actual location lookup
}
