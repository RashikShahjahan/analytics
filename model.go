package main

type EventBase struct {
	Service     string                 `json:"service"`
	Event       string                 `json:"event"`
	Path        string                 `json:"path"`
	Referrer    string                 `json:"referrer"`
	UserBrowser string                 `json:"user_browser"`
	UserDevice  string                 `json:"user_device"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type EventRequest struct {
	Service     string                 `json:"service"`
	Event       string                 `json:"event"`
	Path        string                 `json:"path"`
	Referrer    string                 `json:"referrer"`
	UserBrowser string                 `json:"user_browser"`
	UserDevice  string                 `json:"user_device"`
	Timestamp   string                 `json:"timestamp,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type EventFilter struct {
	Service      string `json:"service,omitempty"`
	Event        string `json:"event,omitempty"`
	Path         string `json:"path,omitempty"`
	Referrer     string `json:"referrer,omitempty"`
	UserBrowser  string `json:"user_browser,omitempty"`
	UserDevice   string `json:"user_device,omitempty"`
	FromTime     string `json:"from,omitempty"`
	ToTime       string `json:"to,omitempty"`
	UserLocation string `json:"user_location,omitempty"`
}

type EventRecord struct {
	Service      string                 `json:"service"`
	Event        string                 `json:"event"`
	Path         string                 `json:"path"`
	Referrer     string                 `json:"referrer"`
	UserBrowser  string                 `json:"user_browser"`
	UserDevice   string                 `json:"user_device"`
	Timestamp    string                 `json:"timestamp"`
	UserIP       string                 `json:"user_ip"`
	UserLocation string                 `json:"user_location"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type QueryBuilder struct {
	baseQuery  string
	conditions []string
	args       []interface{}
}
