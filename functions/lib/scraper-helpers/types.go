package scraperhelpers

type Shift2BikeEvent struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Venue         string `json:"venue"`
	Address       string `json:"address"`
	Organizer     string `json:"organizer"`
	Details       string `json:"details"`
	Time          string `json:"time"`
	Hideemail     bool   `json:"hideemail"`
	Hidephone     bool   `json:"hidephone"`
	Hidecontact   bool   `json:"hidecontact"`
	Length        any    `json:"length"`
	Timedetails   any    `json:"timedetails"`
	Locdetails    string `json:"locdetails"`
	Loopride      bool   `json:"loopride"`
	Locend        any    `json:"locend"`
	Eventduration int    `json:"eventduration"`
	Weburl        string `json:"weburl"`
	Webname       string `json:"webname"`
	Image         string `json:"image"`
	Audience      string `json:"audience"`
	Tinytitle     string `json:"tinytitle"`
	Printdescr    any    `json:"printdescr"`
	Datestype     string `json:"datestype"`
	Area          string `json:"area"`
	Featured      bool   `json:"featured"`
	Printemail    bool   `json:"printemail"`
	Printphone    bool   `json:"printphone"`
	Printweburl   bool   `json:"printweburl"`
	Printcontact  bool   `json:"printcontact"`
	Published     bool   `json:"published"`
	Safetyplan    bool   `json:"safetyplan"`
	Email         any    `json:"email"`
	Phone         any    `json:"phone"`
	Contact       string `json:"contact"`
	Date          string `json:"date"`
	CaldailyID    string `json:"caldaily_id"`
	Shareable     string `json:"shareable"`
	Cancelled     bool   `json:"cancelled"`
	Newsflash     any    `json:"newsflash"`
	Status        string `json:"status"`
	Endtime       string `json:"endtime"`

	/// Location details
	LocationID int      `json:"-"`
	Location   Location `json:"-"`

	/// Sourced From details
	SourcedFrom string `json:"sourcedFrom"`
}

type Shift2BikeEvents struct {
	Events []Shift2BikeEvent `json:"events"`
}

type Location struct {
	ID             int     `json:"-"`
	Address        string  `json:"address"`
	Latitude       float64 `json:"lat"`
	Longitude      float64 `json:"lng"`
	Venue          string  `json:"venue"`
	Details        string  `json:"details"`
	NeedsGeocoding bool    `json:"-"`
}
