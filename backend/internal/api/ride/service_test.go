package ride

import "testing"

func TestNormalizeLinkedRideSubmissionDefaultsDescriptionAndMetadata(t *testing.T) {
	submission := &Submission{
		Title:  " Friday Ride ",
		WebURL: " https://www.strava.com/clubs/example/group_events/123 ",
	}

	normalizeLinkedRideSubmission(submission)

	if submission.Title != "Friday Ride" {
		t.Fatalf("expected trimmed title, got %q", submission.Title)
	}
	if submission.Description != linkedRideDefaultDescription {
		t.Fatalf("expected default linked ride description, got %q", submission.Description)
	}
	if submission.WebURL != "https://www.strava.com/clubs/example/group_events/123" {
		t.Fatalf("expected trimmed web URL, got %q", submission.WebURL)
	}
	if submission.WebName != "Strava" {
		t.Fatalf("expected Strava web name, got %q", submission.WebName)
	}
	if submission.Source != "linked-event" {
		t.Fatalf("expected linked-event source, got %q", submission.Source)
	}
}

func TestNormalizeLinkedRideSubmissionKeepsOrganizerDescriptionAndWebName(t *testing.T) {
	submission := &Submission{
		Description: "Bring lights.",
		WebURL:      "https://example.com/rides/weekly",
		WebName:     "Club site",
		Source:      "club-site",
	}

	normalizeLinkedRideSubmission(submission)

	if submission.Description != "Bring lights." {
		t.Fatalf("expected organizer description to be kept, got %q", submission.Description)
	}
	if submission.WebName != "Club site" {
		t.Fatalf("expected organizer web name to be kept, got %q", submission.WebName)
	}
	if submission.Source != "club-site" {
		t.Fatalf("expected source to be kept, got %q", submission.Source)
	}
}

func TestInferExternalEventLabel(t *testing.T) {
	tests := []struct {
		rawURL string
		want   string
	}{
		{rawURL: "https://strava.app.link/abc123", want: "Strava"},
		{rawURL: "https://www.instagram.com/p/example", want: "Instagram"},
		{rawURL: "https://facebook.com/events/123", want: "Facebook"},
		{rawURL: "https://www.meetup.com/example/events/123", want: "Meetup"},
		{rawURL: "https://club.example/rides", want: "Event link"},
	}

	for _, tt := range tests {
		if got := inferExternalEventLabel(tt.rawURL); got != tt.want {
			t.Fatalf("inferExternalEventLabel(%q) = %q, want %q", tt.rawURL, got, tt.want)
		}
	}
}
