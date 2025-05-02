package main

import (
	"encoding/json"
	"fmt"
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
	// Look up location via ip-api.com
	if ip == "" {
		return "Unknown"
	}
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,regionName,city", ip))
	if err != nil {
		return "Unknown"
	}
	defer resp.Body.Close()

	var result struct {
		Status     string `json:"status"`
		Message    string `json:"message"`
		Country    string `json:"country"`
		RegionName string `json:"regionName"`
		City       string `json:"city"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Status != "success" {
		return "Unknown"
	}

	if result.City != "" {
		return fmt.Sprintf("%s, %s, %s", result.City, result.RegionName, result.Country)
	}
	if result.RegionName != "" {
		return fmt.Sprintf("%s, %s", result.RegionName, result.Country)
	}
	return result.Country
}
