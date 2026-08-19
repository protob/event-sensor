package ticketmaster

// Response is the top-level Ticketmaster Discovery API response.
type Response struct {
	Embedded *EmbeddedEvents `json:"_embedded"`
	Page     Page            `json:"page"`
}

// EmbeddedEvents contains the list of events from the API response.
type EmbeddedEvents struct {
	Events []Event `json:"events"`
}

// Page contains pagination information.
type Page struct {
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
	Number        int `json:"number"`
}

// Event represents a single event from the Ticketmaster API.
type Event struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	URL             string           `json:"url"`
	Dates           Dates            `json:"dates"`
	Embedded        *EventEmbedded   `json:"_embedded"`
	Classifications []Classification `json:"classifications"`
	Info            string           `json:"info"`
	Images          []Image          `json:"images"`
}

// Dates contains start and end date information.
type Dates struct {
	Start DateInfo  `json:"start"`
	End   *DateInfo `json:"end"`
}

// DateInfo contains date/time details.
type DateInfo struct {
	LocalDate string `json:"localDate"`
	LocalTime string `json:"localTime"`
	DateTime  string `json:"dateTime"`
}

// EventEmbedded contains embedded venue and attraction data.
type EventEmbedded struct {
	Venues      []Venue      `json:"venues"`
	Attractions []Attraction `json:"attractions"`
}

// Venue represents a venue from the Ticketmaster API.
type Venue struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	City     City     `json:"city"`
	Country  Country  `json:"country"`
	Location Location `json:"location"`
}

// City contains city name.
type City struct {
	Name string `json:"name"`
}

// Country contains country name and code.
type Country struct {
	Name        string `json:"name"`
	CountryCode string `json:"countryCode"`
}

// Location contains geographic coordinates.
type Location struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// Attraction represents an artist/attraction from the Ticketmaster API.
type Attraction struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Classification contains event classification (segment, genre, type, subType).
type Classification struct {
	Primary bool  `json:"primary"`
	Segment Genre `json:"segment"`
	Genre   Genre `json:"genre"`
	Type    Genre `json:"type"`
	SubType Genre `json:"subType"`
}

// Genre represents a genre or segment classification.
type Genre struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Image represents an event image.
type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
