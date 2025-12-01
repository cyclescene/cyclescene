package scraper

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
	Length        string `json:"length"`
	Timedetails   string `json:"timedetails"`
	Locdetails    string `json:"locdetails"`
	Loopride      bool   `json:"loopride"`
	Locend        string `json:"locend"`
	Eventduration int    `json:"eventduration"`
	Weburl        string `json:"weburl"`
	Webname       string `json:"webname"`
	Image         string `json:"image"`
	Audience      string `json:"audience"`
	Tinytitle     string `json:"tinytitle"`
	Printdescr    string `json:"printdescr"`
	Datestype     string `json:"datestype"`
	Area          string `json:"area"`
	Featured      bool   `json:"featured"`
	Printemail    bool   `json:"printemail"`
	Printphone    bool   `json:"printphone"`
	Printweburl   bool   `json:"printweburl"`
	Printcontact  bool   `json:"printcontact"`
	Published     bool   `json:"published"`
	Safetyplan    bool   `json:"safetyplan"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Contact       string `json:"contact"`
	Date          string `json:"date"`
	CaldailyID    string `json:"caldaily_id"`
	Shareable     string `json:"shareable"`
	Cancelled     bool   `json:"cancelled"`
	Newsflash     string `json:"newsflash"`
	Status        string `json:"status"`
	Endtime       string `json:"endtime"`

	/// Location details
	LocationID int      `json:"-"`
	Location   Location `json:"-"`

	/// Sourced From details
	SourcedFrom string `json:"sourcedfrom"`
	CityCode    string `json:"citycode"`

	/// Route details
	RouteID string `json:"-"`
}

type Shift2BikeEvents struct {
	Events []Shift2BikeEvent `json:"events"`
}

type Location struct {
	ID             int     `json:"-"`
	City           string  `json:"city"`
	Query          string  `json:"query"`
	Address        string  `json:"address"`
	Latitude       float64 `json:"lat"`
	Longitude      float64 `json:"lng"`
	Venue          string  `json:"venue"`
	Details        string  `json:"details"`
	NeedsGeocoding bool    `json:"-"`
}

type GeoCodeCached struct {
	ID        string
	Query     string
	Latitude  float64
	Longitude float64
}

type GoogleGeocodeResponse struct {
	Results []Results `json:"results"`
}
type GeoCodedLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type Low struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type High struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type Viewport struct {
	Low  Low  `json:"low"`
	High High `json:"high"`
}
type Bounds struct {
	Low  Low  `json:"low"`
	High High `json:"high"`
}
type AddressComponents struct {
	LongText     string   `json:"longText"`
	ShortText    string   `json:"shortText"`
	Types        []string `json:"types"`
	LanguageCode string   `json:"languageCode,omitempty"`
}
type Results struct {
	Place             string              `json:"place"`
	PlaceID           string              `json:"placeId"`
	Location          GeoCodedLocation    `json:"location"`
	Granularity       string              `json:"granularity"`
	Viewport          Viewport            `json:"viewport"`
	Bounds            Bounds              `json:"bounds"`
	FormattedAddress  string              `json:"formattedAddress"`
	AddressComponents []AddressComponents `json:"addressComponents"`
	Types             []string            `json:"types"`
}
