package group

type Registration struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	City        string `json:"city"`
	IconURL     string `json:"icon_url"`
	WebURL      string `json:"web_url"`
}

type Response struct {
	Success   bool   `json:"success"`
	Code      string `json:"code,omitempty"`
	EditToken string `json:"edit_token,omitempty"`
	Message   string `json:"message,omitempty"`
}

type ValidationResponse struct {
	Valid bool   `json:"valid"`
	Name  string `json:"name,omitempty"`
}

type AvailabilityResponse struct {
	Available bool   `json:"available"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}
