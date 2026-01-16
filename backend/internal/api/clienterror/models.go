package clienterror

import "time"

type ClientError struct {
	ID         int64     `json:"id"`
	ClientID   string    `json:"client_id"`
	ErrorType  string    `json:"error_type"`
	ErrorMsg   string    `json:"error_msg"`
	StackTrace string    `json:"stack_trace,omitempty"`
	Component  string    `json:"component,omitempty"`
	Action     string    `json:"action,omitempty"`
	URL        string    `json:"url,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	OS         string    `json:"os"`
	CityCode   string    `json:"city_code,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}
