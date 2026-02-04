package strava

import (
	"fmt"
	"strings"
	"time"

	"github.com/spacesedan/cyclescene/backend/internal/api/ride"
)

// Supported city codes and their mappings
var cityMappings = map[string]struct {
	Names    []string // City name substrings to match (case-insensitive)
	Timezone string   // IANA timezone
}{
	"pdx": {Names: []string{"portland"}, Timezone: "America/Los_Angeles"},
	"slc": {Names: []string{"salt lake"}, Timezone: "America/Denver"},
}

// Cycling activity types from Strava
var cyclingActivityTypes = map[string]bool{
	"Ride":       true,
	"EBikeRide":  true,
	"VirtualRide": true,
	"Handcycle":  true,
	"Velomobile": true,
}

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
	ID                 int64    `json:"id"`
	ResourceState      int      `json:"resource_state"`
	Name               string   `json:"name"`
	City               string   `json:"city"`
	State              string   `json:"state"`
	Country            string   `json:"country"`
	Private            bool     `json:"private"`
	MemberCount        int      `json:"member_count"`
	SportType          string   `json:"sport_type"`
	LocalizedSportType string   `json:"localized_sport_type"`
	ActivityTypes      []string `json:"activity_types"`
	ActivityTypesIcon  string   `json:"activity_types_icon"`
	Dimensions         []string `json:"dimensions"`
	Profile            string   `json:"profile"`
	ProfileMedium      string   `json:"profile_medium"`
	CoverPhoto         string   `json:"cover_photo"`
	CoverPhotoSmall    string   `json:"cover_photo_small"`
	URL                string   `json:"url"`
	Featured           bool     `json:"featured"`
	Verified           bool     `json:"verified"`
}

// IsCyclingClub returns true if the club is a cycling club
func (c *Club) IsCyclingClub() bool {
	// Check activity_types array first (preferred)
	for _, actType := range c.ActivityTypes {
		if cyclingActivityTypes[actType] {
			return true
		}
	}
	// Fallback to sport_type field (deprecated but still useful)
	return c.SportType == "cycling"
}

// MatchesCity returns true if the club's city matches the given city code
func (c *Club) MatchesCity(cityCode string) bool {
	mapping, ok := cityMappings[cityCode]
	if !ok {
		return false
	}
	clubCity := strings.ToLower(c.City)
	for _, name := range mapping.Names {
		if strings.Contains(clubCity, name) {
			return true
		}
	}
	return false
}

// ClubDetail represents detailed club info from /clubs/{id}
type ClubDetail struct {
	Club
	Admin          bool   `json:"admin"`          // Is current user an admin
	Owner          bool   `json:"owner"`          // Is current user the owner
	Membership     string `json:"membership"`     // "member", "pending", etc.
	ClubType       string `json:"club_type"`      // "company", "casual_club", "racing_team", etc.
	Description    string `json:"description"`
	PostCount      int    `json:"post_count"`
	OwnerID        int64  `json:"owner_id"`
	FollowingCount int    `json:"following_count"`
}

// IsAdminOrOwner returns true if the user is an admin or owner of the club
func (c *ClubDetail) IsAdminOrOwner() bool {
	return c.Admin || c.Owner
}

// GroupEvent represents a Strava club group event
// Field names match the actual Strava API response from /clubs/{id}/group_events
type GroupEvent struct {
	ID            int64  `json:"id"`
	ResourceState int    `json:"resource_state"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	ActivityType  string `json:"activity_type"` // e.g., "Ride"

	// Time - Strava returns upcoming_occurrences array and zone string
	UpcomingOccurrences []string `json:"upcoming_occurrences"` // ISO 8601 UTC timestamps
	Zone                string   `json:"zone"`                 // IANA timezone e.g., "America/Los_Angeles"

	// Location - Strava returns start_latlng as [lat, lng] array
	Address     string    `json:"address"`
	StartLatLng []float64 `json:"start_latlng"` // [latitude, longitude]

	// Route info (can be null)
	RouteID *int64 `json:"route_id"` // Nullable
	Route   *Route `json:"route"`    // Nullable

	// Event details (can be null)
	SkillLevels *string `json:"skill_levels"` // Nullable
	Terrain     *string `json:"terrain"`      // Nullable
	WomenOnly   bool    `json:"women_only"`

	// Privacy
	Private bool `json:"private"`
	Joined  bool `json:"joined"` // Whether current user has joined

	// Club info - embedded object
	ClubID int64          `json:"club_id"`
	Club   *ClubReference `json:"club,omitempty"`

	// Organizer - full athlete object
	OrganizingAthlete *Athlete `json:"organizing_athlete,omitempty"`

	// Meta
	CreatedAt string `json:"-"` // Not always returned
}

// ClubReference is a minimal club object embedded in events
type ClubReference struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	ResourceState int    `json:"resource_state"`
}

// GetLatitude returns the latitude from start_latlng array
func (e *GroupEvent) GetLatitude() float64 {
	if len(e.StartLatLng) >= 2 {
		return e.StartLatLng[0]
	}
	return 0
}

// GetLongitude returns the longitude from start_latlng array
func (e *GroupEvent) GetLongitude() float64 {
	if len(e.StartLatLng) >= 2 {
		return e.StartLatLng[1]
	}
	return 0
}

// HasLocation returns true if the event has valid coordinates
func (e *GroupEvent) HasLocation() bool {
	return len(e.StartLatLng) >= 2 && e.StartLatLng[0] != 0 && e.StartLatLng[1] != 0
}

// GetFirstOccurrence returns the first upcoming occurrence time
func (e *GroupEvent) GetFirstOccurrence() (time.Time, error) {
	if len(e.UpcomingOccurrences) == 0 {
		return time.Time{}, fmt.Errorf("no upcoming occurrences")
	}
	return time.Parse(time.RFC3339, e.UpcomingOccurrences[0])
}

// GetTimezone returns the timezone location for this event
func (e *GroupEvent) GetTimezone() (*time.Location, error) {
	if e.Zone != "" {
		return time.LoadLocation(e.Zone)
	}
	// Fallback to UTC if no zone specified
	return time.UTC, nil
}

// IsUpcoming returns true if the event's first occurrence is in the future
func (e *GroupEvent) IsUpcoming() bool {
	eventTime, err := e.GetFirstOccurrence()
	if err != nil {
		return false
	}
	return eventTime.After(time.Now())
}

// FilterUpcomingEvents returns only events with occurrences in the future
func FilterUpcomingEvents(events []GroupEvent) []GroupEvent {
	var upcoming []GroupEvent
	for _, event := range events {
		if event.IsUpcoming() {
			upcoming = append(upcoming, event)
		}
	}
	return upcoming
}

// ToSubmission converts a Strava GroupEvent to a CycleScene ride.Submission
// cityCode: the target city (e.g., "pdx", "slc")
// organizerEmail: required for magic link editing
// latitude, longitude: pre-resolved coordinates (from event or geocoding)
func (e *GroupEvent) ToSubmission(cityCode, organizerEmail string, latitude, longitude float64) (*ride.Submission, error) {
	// Get the first occurrence time
	eventTime, err := e.GetFirstOccurrence()
	if err != nil {
		return nil, fmt.Errorf("cannot convert event without occurrence time: %w", err)
	}

	// Convert to local timezone
	tz, err := e.GetTimezone()
	if err != nil {
		// Fallback to city timezone
		if mapping, ok := cityMappings[cityCode]; ok {
			tz, _ = time.LoadLocation(mapping.Timezone)
		} else {
			tz = time.UTC
		}
	}
	localTime := eventTime.In(tz)

	// Calculate ride length from route distance if available
	rideLength := ""
	if e.Route != nil && e.Route.Distance > 0 {
		miles := e.Route.Distance * 0.000621371
		rideLength = fmt.Sprintf("%.1f miles", miles)
	}

	// Build Strava event URL for web link
	stravaURL := fmt.Sprintf("https://www.strava.com/clubs/%d/group_events/%d", e.ClubID, e.ID)

	// Determine venue name - use address if available, otherwise use club name as fallback
	venueName := e.Address
	if venueName == "" && e.Club != nil {
		venueName = e.Club.Name
	}

	submission := &ride.Submission{
		Title:       e.Title,
		Description: e.Description,
		City:        cityCode,

		// Location
		Address:    e.Address,
		VenueName:  venueName, // Use address as venue name, fallback to club name
		IsLoopRide: false,

		// Contact - organizer email required for magic link
		OrganizerEmail: organizerEmail,
		OrganizerName:  "",
		OrganizerPhone: "",
		WebURL:         stravaURL,
		WebName:        "View on Strava",
		HideEmail:      true, // Hide by default for imported events
		HidePhone:      true,

		// Event details
		RideLength: rideLength,
		Audience:   "", // Not mapped from Strava
		DateType:   "O", // One-off event

		// Single occurrence
		Occurrences: []ride.Occurrence{
			{
				StartDate:            localTime.Format("2006-01-02"),
				StartTime:            localTime.Format("15:04"),
				EventDurationMinutes: 120, // Default 2 hours
				EventTimeDetails:     "",
			},
		},
	}

	return submission, nil
}

// StravaEventURL returns the URL to the event on Strava
func (e *GroupEvent) StravaEventURL() string {
	return fmt.Sprintf("https://www.strava.com/clubs/%d/group_events/%d", e.ClubID, e.ID)
}

// Route represents a Strava route
type Route struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Distance      float64  `json:"distance"`       // meters
	ElevationGain float64  `json:"elevation_gain"` // meters
	Type          string   `json:"type"`           // "ride", "run"
	SubType       string   `json:"sub_type"`       // "road", "mtb", "gravel"
	Map           *RouteMap `json:"map,omitempty"`
}

// RouteMap contains the route's map/polyline data
type RouteMap struct {
	SummaryPolyline string `json:"summary_polyline"`
}

// DistanceInMiles returns the route distance in miles
func (r *Route) DistanceInMiles() float64 {
	return r.Distance * 0.000621371
}

// Session represents an ephemeral OAuth session
type Session struct {
	SessionID    string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	AthleteID    int64
	AthleteName  string
	CityCode     string // City code from form URL context (e.g., "pdx", "slc")
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
