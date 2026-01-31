package strava

import "time"

// OAuth Token Response from Strava
type TokenResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresAt    int64   `json:"expires_at"`
	ExpiresIn    int     `json:"expires_in"`
	TokenType    string  `json:"token_type"`
	Athlete      Athlete `json:"athlete"`
}

// Athlete represents a Strava user
type Athlete struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Profile   string `json:"profile"`
	City      string `json:"city"`
	State     string `json:"state"`
	Country   string `json:"country"`
}

// Club represents a Strava club (basic info from /athlete/clubs)
type Club struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	Private     bool   `json:"private"`
	MemberCount int    `json:"member_count"`
	SportType   string `json:"sport_type"`
	Profile     string `json:"profile"`
	ProfileMed  string `json:"profile_medium"`
	CoverPhoto  string `json:"cover_photo"`
	URL         string `json:"url"`
}

// ClubDetail represents detailed club info from /clubs/{id}
type ClubDetail struct {
	Club
	Admin      bool   `json:"admin"`       // Is current user an admin
	Owner      bool   `json:"owner"`       // Is current user the owner
	Membership string `json:"membership"`  // "member", "pending", etc.
	ClubType   string `json:"club_type"`   // "casual_club", "racing_team", etc.
	Description string `json:"description"`
	Featured   bool   `json:"featured"`
	Verified   bool   `json:"verified"`
}

// GroupEvent represents a Strava club group event
type GroupEvent struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	EventTime   string  `json:"event_time"`    // ISO 8601 timestamp

	// Location
	Address         string  `json:"address"`
	LocationCity    string  `json:"location_city"`
	LocationState   string  `json:"location_state"`
	LocationCountry string  `json:"location_country"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`

	// Route info
	RouteID int64  `json:"route_id"`
	Route   *Route `json:"route,omitempty"`

	// Event details
	SkillLevels     string `json:"skill_levels"`
	Terrain         string `json:"terrain"`
	Visibility      string `json:"visibility"`
	WomenOnly       bool   `json:"women_only"`

	// Stats
	AttendingCount  int `json:"attending_count"`
	InterestedCount int `json:"interested_count"`

	// Club/Organizer
	ClubID      int64 `json:"club_id"`
	OrganizerID int64 `json:"organizer_id"`

	// Meta
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Route represents a Strava route
type Route struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Distance      float64 `json:"distance"`       // meters
	ElevationGain float64 `json:"elevation_gain"` // meters
	Type          string  `json:"type"`           // "ride", "run"
	SubType       string  `json:"sub_type"`       // "road", "mtb", "gravel"
}

// Session represents an ephemeral OAuth session
type Session struct {
	SessionID    string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	AthleteID    int64
	AthleteName  string
	CreatedAt    time.Time
}

// ImportRequest is the request body for importing a Strava event
type ImportRequest struct {
	StravaEventID int64  `json:"strava_event_id"`
	GroupCode     string `json:"group_code"`
	City          string `json:"city"`
}

// ImportResponse is the response after successfully importing an event
type ImportResponse struct {
	Success   bool   `json:"success"`
	EventID   int    `json:"event_id"`
	EditToken string `json:"edit_token"`
	Message   string `json:"message,omitempty"`
}
