package ride

import (
	"database/sql"
	"fmt"
	"strings"
)

// User-submitted rides
type Submission struct {
	// Core content
	Title       string `json:"title"`
	TinyTitle   string `json:"tinytitle"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	ImageSrcSet string `json:"image_srcset,omitempty"`
	ImageUUID   string `json:"image_uuid"`
	Audience    string `json:"audience"`
	RideLength  string `json:"ride_length"`
	Area        string `json:"area"`
	DateType    string `json:"date_type"`

	// Location
	VenueName       string `json:"venue_name"`
	Address         string `json:"address"`
	LocationDetails string `json:"location_details"`
	EndingLocation  string `json:"ending_location"`
	IsLoopRide      bool   `json:"is_loop_ride"`

	// Contact
	OrganizerName   string `json:"organizer_name"`
	OrganizerEmail  string `json:"organizer_email"`
	OrganizerPhone  string `json:"organizer_phone"`
	WebURL          string `json:"web_url"`
	WebName         string `json:"web_name"`
	Newsflash       string `json:"newsflash"`
	HideEmail       bool   `json:"hide_email"`
	HidePhone       bool   `json:"hide_phone"`
	HideContactName bool   `json:"hide_contact_name"`

	// Group
	GroupCode string `json:"group_code"`

	// City
	City string `json:"city"`

	// Occurrences
	Occurrences []Occurrence `json:"occurrences"`
}

type Occurrence struct {
	ID                   int64  `json:"id,omitempty"`
	StartDate            string `json:"start_date"`
	StartTime            string `json:"start_time"`
	EventDurationMinutes int    `json:"event_duration_minutes"`
	EventTimeDetails     string `json:"event_time_details"`
	IsCancelled          bool   `json:"is_cancelled,omitempty"`
}

type SubmissionResponse struct {
	Success   bool   `json:"success"`
	EventID   int64  `json:"event_id,omitempty"`
	EditToken string `json:"edit_token,omitempty"`
	Message   string `json:"message,omitempty"`
}

type EditResponse struct {
	Event       Submission `json:"event"`
	IsPublished bool       `json:"is_published"`
}

type RideForAdmin struct {
	ID                  int64         `json:"id"`
	Title               string        `json:"title"`
	Description         string        `json:"description"`
	City                string        `json:"city"`
	VenueName           string        `json:"venue_name"`
	OrganizerName       string        `json:"organizer_name"`
	OrganizerEmail      string        `json:"organizer_email"`
	ImageURL            string        `json:"image_url"`
	ImageUUID           string        `json:"image_uuid"`
	IsPublished         bool          `json:"is_published"`
	IsLoopRide          bool          `json:"is_loop_ride"`
	CreatedAt           string        `json:"created_at"`
	ModerationNotes     string        `json:"moderation_notes"`
	Occurrences         []Occurrence  `json:"occurrences"`
}

// Scraped rides from Shift2Bikes
type ScrapedRideFromDB struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Lat           float64        `json:"lat"`
	Lng           float64        `json:"lng"`
	Address       string         `json:"address"`
	Audience      string         `json:"audience"`
	Cancelled     int            `json:"cancelled"`
	Date          string         `json:"date"`
	StartTime     string         `json:"starttime"`
	SafetyPlan    int            `json:"safetyplan"`
	Details       string         `json:"details"`
	Venue         string         `json:"venue"`
	Organizer     string         `json:"organizer"`
	LoopRide      int            `json:"loopride"`
	Shareable     string         `json:"shareable"`
	RideSource    string         `json:"ridesource"`
	EndTime       sql.NullString `json:"endtime"`
	Email         sql.NullString `json:"email"`
	EventDuration sql.NullInt32  `json:"eventduration"`
	Image         sql.NullString `json:"image"`
	LocDetails    sql.NullString `json:"locdetails"`
	LocEnd        sql.NullString `json:"locend"`
	NewsFlash     sql.NullString `json:"newsflash"`
	TimeDetails   sql.NullString `json:"timedetails"`
	WebURL        sql.NullString `json:"weburl"`
	WebName       sql.NullString `json:"webname"`
}

type ScrapedRide struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	Address       string  `json:"address"`
	Audience      string  `json:"audience"`
	Cancelled     int     `json:"cancelled"`
	Date          string  `json:"date"`
	StartTime     string  `json:"starttime"`
	SafetyPlan    int     `json:"safetyplan"`
	Details       string  `json:"details"`
	Venue         string  `json:"venue"`
	Organizer     string  `json:"organizer"`
	LoopRide      int     `json:"loopride"`
	Shareable     string  `json:"shareable"`
	RideSource    string  `json:"ridesource"`
	EndTime       string  `json:"endtime"`
	Email         string  `json:"email"`
	EventDuration int32   `json:"eventduration"`
	Image         string  `json:"image"`
	ImageSrcSet   string  `json:"image_srcset"`
	LocDetails    string  `json:"locdetails"`
	LocEnd        string  `json:"locend"`
	NewsFlash     string  `json:"newsflash"`
	TimeDetails   string  `json:"timedetails"`
	WebURL        string  `json:"weburl"`
	WebName       string  `json:"webname"`
}

func (rdb *ScrapedRideFromDB) ToScrapedRide() ScrapedRide {
	r := ScrapedRide{
		ID:         rdb.ID,
		Title:      rdb.Title,
		Lat:        rdb.Lat,
		Lng:        rdb.Lng,
		Address:    rdb.Address,
		Audience:   rdb.Audience,
		Cancelled:  rdb.Cancelled,
		Date:       rdb.Date,
		StartTime:  rdb.StartTime,
		SafetyPlan: rdb.SafetyPlan,
		Details:    rdb.Details,
		Venue:      rdb.Venue,
		Organizer:  rdb.Organizer,
		LoopRide:   rdb.LoopRide,
		Shareable:  rdb.Shareable,
		RideSource: rdb.RideSource,
	}
	r.EndTime = rdb.EndTime.String
	r.Email = rdb.Email.String
	if rdb.EventDuration.Valid {
		r.EventDuration = rdb.EventDuration.Int32
	}
	r.Image = rdb.Image.String

	// Generate SrcSet if image is present and is an optimized WebP
	if r.Image != "" && strings.HasSuffix(r.Image, "_optimized.webp") {
		base := strings.TrimSuffix(r.Image, "_optimized.webp")
		r.ImageSrcSet = fmt.Sprintf("%s_400w.webp 400w, %s_800w.webp 800w, %s_1200w.webp 1200w", base, base, base)
	}

	r.LocDetails = rdb.LocDetails.String
	r.LocEnd = rdb.LocEnd.String
	r.NewsFlash = rdb.NewsFlash.String
	r.TimeDetails = rdb.TimeDetails.String
	r.WebURL = rdb.WebURL.String
	r.WebName = rdb.WebName.String
	return r
}

type ICSContent struct {
	Filename string
	Content  string
}
