package group

type Registration struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	City        string `json:"city"`
	WebURL      string `json:"web_url"`
	MarkerColor string `json:"marker_color"`
	ImageUUID   string `json:"image_uuid"`
}

type Response struct {
	Success   bool   `json:"success"`
	Code      string `json:"code,omitempty"`
	ID        string `json:"id,omitempty"`
	PublicID  string `json:"public_id,omitempty"`
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
