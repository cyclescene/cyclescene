package synclog

import "time"

// SyncLog represents a recorded sync event from the client
type SyncLog struct {
	ID        int64     `json:"id"`
	ClientID  string    `json:"client_id"`    // Device/browser identifier (UUID)
	SyncType  string    `json:"sync_type"`    // "periodic", "manual", "foreground"
	Status    string    `json:"status"`       // "success", "error"
	ErrorMsg  string    `json:"error_msg"`    // Error details if failed
	RideCount int       `json:"ride_count"`   // Number of rides synced
	Duration  int       `json:"duration"`     // Sync duration in milliseconds
	Timestamp time.Time `json:"timestamp"`
	CityCode  string    `json:"city_code"`
	OS        string    `json:"os"`           // Operating system (iOS, Android, Linux, Windows, macOS, etc.)
}
