package warren

// LocationService Repo for datacenter location services
type LocationService struct {
	client *Client
}

// Location Schema for datacenter location definitions
type Location struct {
	// DisplayName is the user-friendly name
	DisplayName string `json:"display_name"`
	// IsDefault is true for the location that is chosen if no location is specified
	IsDefault   bool   `json:"is_default"`
	IsPreferred bool   `json:"is_preferred"`
	Description string `json:"description"`
	OrderNr     int    `json:"order_nr"`
	// Slug is the short string that identifies the location.
	// Set Client.LocationSlug to choose which location to access.
	// If using API directly, slug must be included in the URL.
	Slug        string `json:"slug"`
	CountryCode string `json:"country_code"`
}

// ListLocations List all locations for client datacenter
func (c *LocationService) ListLocations() (*[]Location, error) {
	var locations []Location
	err := c.client.Call(ApiCall{
		method:       "GET",
		path:         "/config/locations",
		responseData: &locations,
	})
  if err != nil {
    return nil, err
  }
	return &locations, err
}
