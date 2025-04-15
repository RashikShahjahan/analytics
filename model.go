package main

type EventBase struct {
	Service     string `json:"service"`
	Event       string `json:"event"`
	Path        string `json:"path"`
	Referrer    string `json:"referrer"`
	UserBrowser string `json:"user_browser"`
	UserDevice  string `json:"user_device"`
}

type EventRequest struct {
	EventBase
	Timestamp string `json:"timestamp"`
}

type EventFilter struct {
	EventBase
	FromTime string `json:"from,omitempty"`
	ToTime   string `json:"to,omitempty"`
}

type EventRecord struct {
	EventBase
	Timestamp    string `json:"timestamp"`
	UserIP       string `json:"user_ip"`
	UserLocation string `json:"user_location"`
}

type QueryBuilder struct {
	baseQuery  string
	conditions []string
	args       []interface{}
}
