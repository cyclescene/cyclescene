package main

import "database/sql"

type RideFromDB struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Address    string  `json:"address"`
	Audience   string  `json:"audience"`
	Cancelled  int     `json:"cancelled"`
	Date       string  `json:"date"`
	StartTime  string  `json:"starttime"`
	SafetyPlan int     `json:"safetyplan"`
	Details    string  `json:"details"`
	Venue      string  `json:"venue"`
	Organizer  string  `json:"organizer"`
	LoopRide   int     `json:"loopride"`
	Shareable  string  `json:"shareable"`
	RideSource string  `json:"ridesource"`

	EndTime       sql.NullString `json:"endtime"`
	Email         sql.NullString `json:"email"`
	EventDuration sql.NullInt32  `json:"eventduration"`
	Image         sql.NullString `json:"image"`
	LocDetails    sql.NullString `json:"locdetails"`
	LocEnd        sql.NullString `json:"locend"`
	NewsFlash     sql.NullString `json:"newsflash"`
	TimeDetails   sql.NullString `json:"timedetails"`
	WebUrl        sql.NullString `json:"weburl"`
	WebName       sql.NullString `json:"webname"`
}

type Ride struct {
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
	LocDetails    string  `json:"locdetails"`
	LocEnd        string  `json:"locend"`
	NewsFlash     string  `json:"newsflash"`
	TimeDetails   string  `json:"timedetails"`
	WebUrl        string  `json:"weburl"`
	WebName       string  `json:"webname"`
}

func (rdb *RideFromDB) ToRide() Ride {
	r := Ride{
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
	r.LocDetails = rdb.LocDetails.String
	r.LocEnd = rdb.LocEnd.String
	r.NewsFlash = rdb.NewsFlash.String
	r.TimeDetails = rdb.TimeDetails.String
	r.WebUrl = rdb.WebUrl.String
	r.WebName = rdb.WebName.String

	return r

}
