package ticketmaster

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DefaultBaseURL is the Ticketmaster Discovery events endpoint.
const DefaultBaseURL = "https://app.ticketmaster.com/discovery/v2/events.json"

// Client is an HTTP client for the Ticketmaster Discovery API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a client against the Ticketmaster API.
func NewClient(apiKey string) *Client {
	return NewClientWithBase(apiKey, DefaultBaseURL)
}

// NewClientWithBase creates a client against baseURL, which can be a local server
// returning recorded responses. An empty baseURL means DefaultBaseURL.
func NewClientWithBase(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SearchEvents searches for events by keyword.
func (c *Client) SearchEvents(keyword string) (*Response, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("keyword", keyword)
	params.Set("locale", "*")
	params.Set("size", "200")
	// Intentionally NOT constrained to the Music segment. Many tributes / "music of X"
	// shows are filed under Arts & Theatre, and those can be of interest. Cross-segment
	// noise (the comedy "Les Bonobos", sports, etc.) is filtered downstream by matching the
	// billed attraction - see ticketmaster.ClassifyArtistMatch.

	resp, err := c.httpClient.Get(fmt.Sprintf("%s?%s", c.baseURL, params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("ticketmaster request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ticketmaster returned status %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ticketmaster response: %w", err)
	}
	return &result, nil
}
