package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spacesedan/cyclescene/backend/internal/strava"
)

var (
	clientID     string
	clientSecret string
	redirectURI  = "http://localhost:3000/callback"
	state        = "test-state-12345"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	Athlete      struct {
		ID        int64  `json:"id"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
	} `json:"athlete"`
}

type Club struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	State       string `json:"state"`
	MemberCount int    `json:"member_count"`
	Private     bool   `json:"private"`
	Admin       bool   `json:"admin"`  // Only in detailed response
	Owner       bool   `json:"owner"`  // Only in detailed response
}

func main() {
	// Load environment variables
	if os.Getenv("APP_ENV") == "dev" {
		// Try multiple locations for .env file
		// When run from backend dir: .env
		// When run from repo root: backend/.env
		if err := godotenv.Load(".env"); err != nil {
			if err := godotenv.Load("backend/.env"); err != nil {
				if err := godotenv.Load("../../.env"); err != nil {
					log.Println("Warning: Could not load .env file, using environment variables")
				}
			}
		}
	}

	clientID = os.Getenv("STRAVA_CLIENT_ID")
	clientSecret = os.Getenv("STRAVA_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET must be set in environment or backend/.env file")
	}

	fmt.Println("=== Strava OAuth Test Tool ===")
	fmt.Println("Client ID:", clientID)
	fmt.Println("")

	// Initialize M2 Service layer
	config := &strava.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackPath: "/callback",
		Debug:        true,
	}
	client := strava.NewClient(config)
	sessionStore = strava.NewSessionStore()
	service = strava.NewService(client, sessionStore, nil, redirectURI)

	// Set up HTTP server for OAuth callback
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/test", handleTest)
	http.HandleFunc("/all-clubs", handleAllClubs)
	http.HandleFunc("/test-admin", handleTestAdmin)
	http.HandleFunc("/test-admin-check", handleTestAdminCheck)
	http.HandleFunc("/test-members", handleTestMembers)

	// M2 Service layer test endpoints
	http.HandleFunc("/test-service", handleTestService)
	http.HandleFunc("/test-service-clubs", handleTestServiceClubs)
	http.HandleFunc("/test-service-events", handleTestServiceEvents)

	// M3 Import test endpoints
	http.HandleFunc("/test-import-select", handleTestImportSelect)
	http.HandleFunc("/test-convert", handleTestConvert)
	http.HandleFunc("/test-import-preview", handleTestImportPreview)
	http.HandleFunc("/test-websocket", handleTestWebSocket)

	fmt.Println("Starting server on http://localhost:3000")
	fmt.Println("")
	fmt.Println("Step 1: Visit http://localhost:3000 to start OAuth flow")
	fmt.Println("Step 2: After OAuth, visit http://localhost:3000/test to test endpoints")
	fmt.Println("")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf(
		"https://www.strava.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=read,read_all&state=%s",
		clientID,
		redirectURI,
		state,
	)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Strava OAuth Test</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 15px 30px; font-size: 16px; cursor: pointer; border-radius: 5px; }
        button:hover { background: #E34402; }
        .info { background: #f0f0f0; padding: 15px; border-radius: 5px; margin: 20px 0; }
        code { background: #fff; padding: 2px 5px; border: 1px solid #ddd; }
    </style>
</head>
<body>
    <h1>Strava OAuth Test Tool</h1>

    <div class="info">
        <strong>Environment Setup:</strong><br>
        Client ID: <code>%s</code><br>
        Redirect URI: <code>%s</code><br>
        Scopes: <code>read,read_all</code>
    </div>

    <h2>Step 1: Authenticate with Strava</h2>
    <p>Click the button below to start the OAuth flow:</p>
    <button onclick="window.location.href='%s'">Connect to Strava</button>

    <h2>What happens next:</h2>
    <ol>
        <li>You'll be redirected to Strava's authorization page</li>
        <li>Log in and authorize the app</li>
        <li>You'll be redirected back here with an access token</li>
        <li>We'll fetch your clubs and test the group_events endpoint</li>
    </ol>
</body>
</html>
	`, clientID, redirectURI, authURL)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

var accessToken string
var athleteID int64

// M2 Service layer testing
var (
	service      *strava.Service
	sessionStore *strava.SessionStore
	sessionID    string
)

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	returnedState := r.URL.Query().Get("state")

	// Check if this is a legacy or M2 state (without consuming it)
	isLegacyState := (returnedState == state)

	if !isLegacyState {
		// Must be M2 state - validate via service (this will consume it)
		// Let M2 service handle the full callback
		fmt.Printf("\n=== M2 OAuth Callback ===\n")
		var err error
		sessionID, err = service.HandleOAuthCallback(context.Background(), code, returnedState)
		if err != nil {
			http.Error(w, "M2 OAuth callback failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		session, _ := service.GetSession(sessionID)
		athleteID = session.AthleteID
		accessToken = session.AccessToken

		fmt.Printf("✓ M2 Session created: %s...\n", sessionID[:16])
		fmt.Printf("  - City Code: %s\n", session.CityCode)
		fmt.Printf("  - Athlete: %s (ID: %d)\n", session.AthleteName, session.AthleteID)
		fmt.Printf("  - Expires: %s\n", session.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("========================\n\n")

		// Redirect to M2 test page
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>M2 OAuth Success</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .success { background: #e8f5e9; color: #2e7d32; padding: 20px; border-radius: 5px; margin: 20px 0; }
        button { background: #FC4C02; color: white; border: none; padding: 15px 30px; font-size: 16px; cursor: pointer; border-radius: 5px; margin: 10px; }
        button:hover { background: #E34402; }
        code { background: #f0f0f0; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>✓ M2 OAuth Successful!</h1>
    <div class="success">
        <strong>Authenticated as:</strong> %s (ID: %d)<br>
        <strong>Session ID:</strong> <code>%s...</code><br>
        <strong>City Context:</strong> %s<br>
        <strong>Expires:</strong> %s
    </div>
    <h2>Next Steps</h2>
    <p>Test the M2 Service Layer:</p>
    <button onclick="window.location.href='/test-service-clubs'">Test GetAdminClubs()</button>
    <button onclick="window.location.href='/test-service'">Back to Test Menu</button>
</body>
</html>
		`, session.AthleteName, session.AthleteID, sessionID[:16], session.CityCode,
		   session.ExpiresAt.Format(time.RFC3339))

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
		return
	}

	// Legacy OAuth flow
	fmt.Printf("\n=== Legacy OAuth Callback ===\n")

	// Exchange code for token
	tokenResp, err := http.PostForm("https://www.strava.com/oauth/token", map[string][]string{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tokenResp.Body.Close()

	var tokens TokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokens); err != nil {
		http.Error(w, "Failed to parse token response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store token globally for testing
	accessToken = tokens.AccessToken
	athleteID = tokens.Athlete.ID

	fmt.Printf("✓ Legacy OAuth successful!\n")
	fmt.Printf("Athlete: %s %s (ID: %d)\n", tokens.Athlete.FirstName, tokens.Athlete.LastName, tokens.Athlete.ID)
	fmt.Printf("Access Token: %s...\n", accessToken[:20])
	fmt.Printf("Expires At: %s\n", time.Unix(tokens.ExpiresAt, 0).Format(time.RFC3339))
	fmt.Printf("=============================\n\n")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>OAuth Success</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .success { background: #4CAF50; color: white; padding: 15px; border-radius: 5px; margin: 20px 0; }
        button { background: #FC4C02; color: white; border: none; padding: 15px 30px; font-size: 16px; cursor: pointer; border-radius: 5px; margin: 10px 5px; }
        button:hover { background: #E34402; }
        code { background: #f0f0f0; padding: 2px 5px; border: 1px solid #ddd; }
        pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>✓ OAuth Successful!</h1>

    <div class="success">
        <strong>Authenticated as:</strong> %s %s (ID: %d)<br>
        <strong>Access Token:</strong> <code>%s...</code><br>
        <strong>Expires:</strong> %s
    </div>

    <h2>Step 2: Test Endpoints</h2>
    <p><strong>M1 Client Tests (Raw API):</strong></p>
    <button onclick="window.location.href='/test'">Test Admin/Owner Clubs</button>
    <button onclick="window.location.href='/all-clubs'">View All Clubs</button>
    <button onclick="window.location.href='/test-admin-check'">Test Admin Check</button>

    <p><strong>M2 Service Layer Tests:</strong></p>
    <button onclick="window.location.href='/test-service'">🧪 M2 Service Tests</button>
    <button onclick="window.location.href='/test-service-clubs'">Get Admin Clubs (M2)</button>

    <h3>Or test manually:</h3>
    <ul>
        <li><a href="/test">GET /api/v3/athlete/clubs (Admin/Owner only)</a></li>
        <li><a href="/all-clubs">GET /api/v3/athlete/clubs (All clubs)</a></li>
        <li><a href="/test-admin-check">Test /clubs/{id}/admins for all clubs</a></li>
        <li><a href="/test?club_id=YOUR_CLUB_ID">GET /api/v3/clubs/{id}/group_events</a></li>
    </ul>
</body>
</html>
	`, tokens.Athlete.FirstName, tokens.Athlete.LastName, tokens.Athlete.ID, accessToken[:20], time.Unix(tokens.ExpiresAt, 0).Format(time.RFC3339))

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	if accessToken == "" {
		http.Error(w, "Not authenticated. Please go to http://localhost:3000 first", http.StatusUnauthorized)
		return
	}

	clubID := r.URL.Query().Get("club_id")

	var htmlOutput string

	// Fetch clubs
	clubs, clubsJSON, err := fetchClubs()
	if err != nil {
		http.Error(w, "Failed to fetch clubs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	htmlOutput += fmt.Sprintf("<h2>Your Clubs (Admin/Owner only)</h2>")
	htmlOutput += fmt.Sprintf("<p>Found %d clubs where you have admin/owner access:</p>", len(clubs))

	for _, club := range clubs {
		role := "Member"
		if club.Owner {
			role = "Owner"
		} else if club.Admin {
			role = "Admin"
		}

		htmlOutput += fmt.Sprintf(`
			<div class="club-card">
				<h3>%s (ID: %d)</h3>
				<p><strong>Role:</strong> %s</p>
				<p><strong>Location:</strong> %s, %s</p>
				<p><strong>Members:</strong> %d</p>
				<button onclick="window.location.href='/test?endpoint=club_events&club_id=%d'">
					Fetch Group Events
				</button>
			</div>
		`, club.Name, club.ID, role, club.City, club.State, club.MemberCount, club.ID)
	}

	htmlOutput += `<h3>Raw Clubs API Response:</h3><pre>` + clubsJSON + `</pre>`

	// If club_id provided, fetch group events
	if clubID != "" {
		events, eventsJSON, err := fetchGroupEvents(clubID)
		if err != nil {
			htmlOutput += fmt.Sprintf(`<div class="error">Failed to fetch group events: %s</div>`, err.Error())
		} else {
			htmlOutput += fmt.Sprintf(`
				<h2>Group Events for Club %s</h2>
				<p>Found %d group events</p>
				<h3>Raw group_events API Response:</h3>
				<pre>%s</pre>
			`, clubID, events, eventsJSON)
		}
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Strava API Test Results</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        .club-card { background: #f9f9f9; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #FC4C02; }
        .error { background: #ffebee; color: #c62828; padding: 15px; border-radius: 5px; margin: 10px 0; }
        pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; font-size: 12px; }
        h2 { color: #FC4C02; }
    </style>
</head>
<body>
    <h1>Strava API Test Results</h1>
    %s
</body>
</html>
	`, htmlOutput)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func fetchClubs() ([]Club, string, error) {
	req, _ := http.NewRequest("GET", "https://www.strava.com/api/v3/athlete/clubs", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// LOG ALL RESPONSE HEADERS
	fmt.Printf("\n=== RESPONSE HEADERS FOR /athlete/clubs ===\n")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Printf("===========================================\n\n")

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyJSON := string(bodyBytes)

	// Pretty print JSON
	var prettyJSON interface{}
	json.Unmarshal(bodyBytes, &prettyJSON)
	prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")
	bodyJSON = string(prettyBytes)

	if resp.StatusCode != 200 {
		return nil, bodyJSON, fmt.Errorf("status %d: %s", resp.StatusCode, bodyJSON)
	}

	var clubs []Club
	json.Unmarshal(bodyBytes, &clubs)

	// Filter for admin/owner only
	adminClubs := []Club{}
	for _, club := range clubs {
		if club.Admin || club.Owner {
			adminClubs = append(adminClubs, club)
		}
	}

	return adminClubs, bodyJSON, nil
}

func isClubAdmin(clubID int64) (bool, error) {
	// Use detailed club endpoint which includes admin/owner flags
	url := fmt.Sprintf("https://www.strava.com/api/v3/clubs/%d", clubID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error fetching club details for %d: %v\n", clubID, err)
		return false, err
	}
	defer resp.Body.Close()

	// LOG ALL RESPONSE HEADERS
	fmt.Printf("\n=== RESPONSE HEADERS FOR /clubs/%d ===\n", clubID)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Printf("=========================================\n\n")

	bodyBytes, _ := io.ReadAll(resp.Body)

	fmt.Printf("\n=== ADMIN CHECK (via /clubs/%d) ===\n", clubID)
	fmt.Printf("Status: %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		fmt.Printf("✗ Failed to fetch club details\n")
		fmt.Printf("===================================\n\n")
		return false, nil
	}

	var clubDetail struct {
		ID    int64 `json:"id"`
		Name  string `json:"name"`
		Admin bool  `json:"admin"`
		Owner bool  `json:"owner"`
	}

	if err := json.Unmarshal(bodyBytes, &clubDetail); err != nil {
		fmt.Printf("Error unmarshaling club details: %v\n", err)
		return false, err
	}

	isAdmin := clubDetail.Admin || clubDetail.Owner

	fmt.Printf("Club: %s\n", clubDetail.Name)
	fmt.Printf("Admin: %t\n", clubDetail.Admin)
	fmt.Printf("Owner: %t\n", clubDetail.Owner)
	fmt.Printf("Result: %t\n", isAdmin)
	fmt.Printf("===================================\n\n")

	return isAdmin, nil
}

func fetchGroupEvents(clubID string) (int, string, error) {
	url := fmt.Sprintf("https://www.strava.com/api/v3/clubs/%s/group_events", clubID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	// LOG ALL RESPONSE HEADERS
	fmt.Printf("\n=== RESPONSE HEADERS FOR /clubs/%s/group_events ===\n", clubID)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Printf("====================================================\n\n")

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyJSON := string(bodyBytes)

	// Pretty print JSON
	var prettyJSON interface{}
	json.Unmarshal(bodyBytes, &prettyJSON)
	prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")
	bodyJSON = string(prettyBytes)

	fmt.Printf("\n=== GROUP EVENTS API RESPONSE ===\n")
	fmt.Printf("Club ID: %s\n", clubID)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response:\n%s\n", bodyJSON)
	fmt.Printf("=================================\n\n")

	if resp.StatusCode != 200 {
		return 0, bodyJSON, fmt.Errorf("status %d: %s", resp.StatusCode, bodyJSON)
	}

	var events []interface{}
	json.Unmarshal(bodyBytes, &events)

	return len(events), bodyJSON, nil
}

func handleAllClubs(w http.ResponseWriter, r *http.Request) {
	if accessToken == "" {
		http.Error(w, "Not authenticated. Please go to http://localhost:3000 first", http.StatusUnauthorized)
		return
	}

	clubID := r.URL.Query().Get("club_id")

	// Fetch ALL clubs (no filtering)
	req, _ := http.NewRequest("GET", "https://www.strava.com/api/v3/athlete/clubs", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch clubs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// Pretty print JSON
	var prettyJSON interface{}
	json.Unmarshal(bodyBytes, &prettyJSON)
	prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")
	clubsJSON := string(prettyBytes)

	if resp.StatusCode != 200 {
		http.Error(w, "Failed to fetch clubs: "+clubsJSON, http.StatusInternalServerError)
		return
	}

	var allClubs []Club
	json.Unmarshal(bodyBytes, &allClubs)

	var htmlOutput string
	htmlOutput += fmt.Sprintf("<h2>All Your Clubs</h2>")
	htmlOutput += fmt.Sprintf("<p>Found %d total clubs:</p>", len(allClubs))

	for _, club := range allClubs {
		// Check if user is admin/owner via /clubs/{id}/admins endpoint
		isAdmin, err := isClubAdmin(club.ID)

		role := "Member"
		if err == nil && isAdmin {
			role = "Admin/Owner ⭐"
		}

		location := "Unknown"
		if club.City != "" && club.State != "" {
			location = fmt.Sprintf("%s, %s", club.City, club.State)
		} else if club.City != "" {
			location = club.City
		} else if club.State != "" {
			location = club.State
		}

		htmlOutput += fmt.Sprintf(`
			<div class="club-card">
				<h3>%s (ID: %d)</h3>
				<p><strong>Role:</strong> %s</p>
				<p><strong>Location:</strong> %s</p>
				<p><strong>Members:</strong> %d</p>
				<p><strong>Private:</strong> %t</p>
				<button onclick="window.location.href='/all-clubs?club_id=%d'">
					Fetch Group Events
				</button>
			</div>
		`, club.Name, club.ID, role, location, club.MemberCount, club.Private, club.ID)
	}

	htmlOutput += `<h3>Raw Clubs API Response:</h3><pre>` + clubsJSON + `</pre>`

	// If club_id provided, fetch group events
	if clubID != "" {
		events, eventsJSON, err := fetchGroupEvents(clubID)
		if err != nil {
			htmlOutput += fmt.Sprintf(`<div class="error">Failed to fetch group events: %s</div>`, err.Error())
		} else {
			htmlOutput += fmt.Sprintf(`
				<h2>Group Events for Club %s</h2>
				<p>Found %d group events</p>
				<h3>Raw group_events API Response:</h3>
				<pre>%s</pre>
			`, clubID, events, eventsJSON)
		}
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>All Clubs - Strava API Test</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        .club-card { background: #f9f9f9; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #FC4C02; }
        .error { background: #ffebee; color: #c62828; padding: 15px; border-radius: 5px; margin: 10px 0; }
        pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; font-size: 12px; }
        h2 { color: #FC4C02; }
        .nav { background: #f0f0f0; padding: 10px; border-radius: 5px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="nav">
        <button onclick="window.location.href='/test'">Admin/Owner Clubs Only</button>
        <button onclick="window.location.href='/all-clubs'">All Clubs</button>
    </div>
    <h1>All Clubs - Strava API Test</h1>
    %s
</body>
</html>
	`, htmlOutput)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleTestAdmin(w http.ResponseWriter, r *http.Request) {
	if accessToken == "" {
		http.Error(w, "Not authenticated. Please go to http://localhost:3000 first", http.StatusUnauthorized)
		return
	}

	clubID := r.URL.Query().Get("club_id")
	if clubID == "" {
		clubID = "1947941" // Default to your test club
	}

	// Test the admins endpoint
	url := fmt.Sprintf("https://www.strava.com/api/v3/clubs/%s/admins", clubID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// Pretty print
	var prettyJSON interface{}
	json.Unmarshal(bodyBytes, &prettyJSON)
	prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Test Admin Endpoint</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1000px; margin: 50px auto; padding: 20px; }
        pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; }
        .info { background: #e3f2fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .error { background: #ffebee; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .success { background: #e8f5e9; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <h1>Test /clubs/{id}/admins Endpoint</h1>

    <div class="info">
        <strong>Testing Club ID:</strong> %s<br>
        <strong>Your Athlete ID:</strong> %d<br>
        <strong>Endpoint:</strong> <code>GET %s</code>
    </div>

    <h2>Response</h2>
    <div class="%s">
        <strong>Status Code:</strong> %d
    </div>

    <h3>Raw JSON Response:</h3>
    <pre>%s</pre>

    <h3>Interpretation:</h3>
    <ul>
        <li><strong>Status 200:</strong> Successfully retrieved admins list</li>
        <li><strong>Status 403:</strong> Forbidden - You don't have permission to view admins</li>
        <li><strong>Status 404:</strong> Club not found or endpoint doesn't exist</li>
    </ul>

    <form>
        <label>Test different club: <input name="club_id" placeholder="Club ID" value="%s"></label>
        <button type="submit">Test</button>
    </form>
</body>
</html>
	`, clubID, athleteID, url,
	   getStatusClass(resp.StatusCode), resp.StatusCode,
	   string(prettyBytes), clubID)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getStatusClass(status int) string {
	if status == 200 {
		return "success"
	}
	return "error"
}

func handleTestAdminCheck(w http.ResponseWriter, r *http.Request) {
	if accessToken == "" {
		http.Error(w, "Not authenticated. Please go to http://localhost:3000 first", http.StatusUnauthorized)
		return
	}

	selectedClubID := r.URL.Query().Get("club_id")

	// Fetch all clubs
	req, _ := http.NewRequest("GET", "https://www.strava.com/api/v3/athlete/clubs", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch clubs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var allClubs []Club
	json.Unmarshal(bodyBytes, &allClubs)

	var htmlOutput string
	htmlOutput += fmt.Sprintf("<h2>Test Admin Check for All Clubs</h2>")
	htmlOutput += fmt.Sprintf("<p>Your Athlete ID: <strong>%d</strong></p>", athleteID)
	htmlOutput += fmt.Sprintf("<p>Found %d total clubs. Click 'Test Admin' to check admin status:</p>", len(allClubs))

	for _, club := range allClubs {
		location := "Unknown"
		if club.City != "" && club.State != "" {
			location = fmt.Sprintf("%s, %s", club.City, club.State)
		} else if club.City != "" {
			location = club.City
		} else if club.State != "" {
			location = club.State
		}

		htmlOutput += fmt.Sprintf(`
			<div class="club-card">
				<h3>%s (ID: %d)</h3>
				<p><strong>Location:</strong> %s</p>
				<p><strong>Members:</strong> %d</p>
				<p><strong>Private:</strong> %t</p>
				<button onclick="window.location.href='/test-admin-check?club_id=%d'">
					Test Admin Status
				</button>
			</div>
		`, club.Name, club.ID, location, club.MemberCount, club.Private, club.ID)
	}

	// If club_id provided, test admin status
	if selectedClubID != "" {
		htmlOutput += fmt.Sprintf(`<hr><h2>Admin Check Result for Club %s</h2>`, selectedClubID)

		adminURL := fmt.Sprintf("https://www.strava.com/api/v3/clubs/%s/admins", selectedClubID)
		adminReq, _ := http.NewRequest("GET", adminURL, nil)
		adminReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		adminResp, err := http.DefaultClient.Do(adminReq)
		if err != nil {
			htmlOutput += fmt.Sprintf(`<div class="error">Error: %s</div>`, err.Error())
		} else {
			defer adminResp.Body.Close()
			adminBodyBytes, _ := io.ReadAll(adminResp.Body)

			// Pretty print
			var prettyJSON interface{}
			json.Unmarshal(adminBodyBytes, &prettyJSON)
			prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")

			statusClass := "error"
			if adminResp.StatusCode == 200 {
				statusClass = "success"
			}

			htmlOutput += fmt.Sprintf(`
				<div class="%s">
					<strong>HTTP Status:</strong> %d<br>
					<strong>Endpoint:</strong> <code>GET %s</code>
				</div>
			`, statusClass, adminResp.StatusCode, adminURL)

			htmlOutput += fmt.Sprintf(`<h3>Raw Response:</h3><pre>%s</pre>`, string(prettyBytes))

			// Check if user is in admins list
			if adminResp.StatusCode == 200 {
				var admins []struct {
					ID        int64  `json:"id"`
					FirstName string `json:"firstname"`
					LastName  string `json:"lastname"`
				}
				json.Unmarshal(adminBodyBytes, &admins)

				isAdmin := false
				for _, admin := range admins {
					if admin.ID == athleteID {
						isAdmin = true
						break
					}
				}

				if isAdmin {
					htmlOutput += `<div class="success"><h3>✓ You ARE an admin of this club!</h3></div>`
				} else {
					htmlOutput += `<div class="error"><h3>✗ You are NOT an admin of this club</h3></div>`
					htmlOutput += fmt.Sprintf(`<p>Your athlete ID (%d) was not found in the admins list above.</p>`, athleteID)
				}
			} else if adminResp.StatusCode == 403 {
				htmlOutput += `<div class="error"><h3>403 Forbidden</h3><p>You don't have permission to view this club's admins list.</p></div>`
			} else if adminResp.StatusCode == 404 {
				htmlOutput += `<div class="error"><h3>404 Not Found</h3><p>This endpoint doesn't exist or the club was not found.</p></div>`
			}
		}
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Test Admin Check - Strava API</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        .club-card { background: #f9f9f9; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #FC4C02; }
        .error { background: #ffebee; color: #c62828; padding: 15px; border-radius: 5px; margin: 10px 0; }
        .success { background: #e8f5e9; color: #2e7d32; padding: 15px; border-radius: 5px; margin: 10px 0; }
        pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; font-size: 12px; }
        h2 { color: #FC4C02; }
        .nav { background: #f0f0f0; padding: 10px; border-radius: 5px; margin-bottom: 20px; }
        hr { margin: 30px 0; border: none; border-top: 2px solid #ddd; }
    </style>
</head>
<body>
    <div class="nav">
        <button onclick="window.location.href='/test'">Admin/Owner Clubs Only</button>
        <button onclick="window.location.href='/all-clubs'">All Clubs</button>
        <button onclick="window.location.href='/test-admin-check'">Test Admin Check</button>
    </div>
    <h1>Test Admin Status Check</h1>
    %s
</body>
</html>
	`, htmlOutput)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleTestMembers(w http.ResponseWriter, r *http.Request) {
	if accessToken == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	clubID := r.URL.Query().Get("club_id")
	if clubID == "" {
		clubID = "1947941"
	}

	var htmlOutput string
	htmlOutput += fmt.Sprintf("<h2>Testing Different Endpoints for Club %s</h2>", clubID)
	htmlOutput += fmt.Sprintf("<p>Your Athlete ID: <strong>%d</strong></p>", athleteID)

	// Test 1: /clubs/{id}/admins
	htmlOutput += "<h3>1. GET /clubs/{id}/admins</h3>"
	adminURL := fmt.Sprintf("https://www.strava.com/api/v3/clubs/%s/admins", clubID)
	adminResp, adminBody := fetchStravaEndpoint(adminURL)
	htmlOutput += fmt.Sprintf("<p>Status: %d</p><pre>%s</pre>", adminResp, adminBody)

	// Test 2: /clubs/{id}/members  
	htmlOutput += "<h3>2. GET /clubs/{id}/members</h3>"
	membersURL := fmt.Sprintf("https://www.strava.com/api/v3/clubs/%s/members", clubID)
	membersResp, membersBody := fetchStravaEndpoint(membersURL)
	htmlOutput += fmt.Sprintf("<p>Status: %d</p><pre>%s</pre>", membersResp, membersBody)

	// Test 3: /clubs/{id} (detailed club info)
	htmlOutput += "<h3>3. GET /clubs/{id} (Detailed Club Info)</h3>"
	clubURL := fmt.Sprintf("https://www.strava.com/api/v3/clubs/%s", clubID)
	clubResp, clubBody := fetchStravaEndpoint(clubURL)
	htmlOutput += fmt.Sprintf("<p>Status: %d</p><pre>%s</pre>", clubResp, clubBody)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Test Multiple Endpoints</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; font-size: 12px; }
        h2 { color: #FC4C02; }
        h3 { color: #666; margin-top: 30px; }
    </style>
</head>
<body>
    <h1>Compare Strava Club Endpoints</h1>
    %s
</body>
</html>
	`, htmlOutput)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func fetchStravaEndpoint(url string) (int, string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Sprintf("Error: %s", err.Error())
	}
	defer resp.Body.Close()

	// LOG ALL RESPONSE HEADERS
	fmt.Printf("\n=== RESPONSE HEADERS FOR %s ===\n", url)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Printf("=========================================\n\n")

	bodyBytes, _ := io.ReadAll(resp.Body)

	var prettyJSON interface{}
	json.Unmarshal(bodyBytes, &prettyJSON)
	prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")

	return resp.StatusCode, string(prettyBytes)
}

// M2 Service Layer Test Endpoints

func handleTestService(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>M2 Service Layer Test</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 15px 30px; font-size: 16px; cursor: pointer; border-radius: 5px; margin: 10px; }
        button:hover { background: #E34402; }
        .info { background: #e3f2fd; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .success { background: #e8f5e9; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .feature { background: #f9f9f9; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #FC4C02; }
        code { background: #f0f0f0; padding: 2px 5px; border-radius: 3px; }
        h1 { color: #FC4C02; }
        h2 { color: #666; }
    </style>
</head>
<body>
    <h1>🧪 M2 Service Layer Test</h1>

    <div class="info">
        <h2>What This Tests</h2>
        <p>This page tests the <strong>M2 Service Layer</strong> implementation, which includes:</p>
        <ul>
            <li>OAuth session management with city context</li>
            <li>City-first filtering (83% API reduction)</li>
            <li>Admin club detection</li>
            <li>Event fetching and conversion</li>
            <li>Upcoming events filter</li>
            <li>Geocoding fallback</li>
        </ul>
    </div>

    <h2>Test Flow</h2>
    <div class="feature">
        <h3>Step 1: OAuth with City Context</h3>
        <p>Tests <code>Service.InitiateOAuth()</code> and <code>HandleOAuthCallback()</code></p>
        <p>City code "pdx" will be stored with the session.</p>
        <button onclick="window.location.href='/test-service?action=oauth&city=pdx'">Start OAuth (Portland)</button>
        <button onclick="window.location.href='/test-service?action=oauth&city=slc'">Start OAuth (Salt Lake)</button>
    </div>

    <div class="feature">
        <h3>Step 2: Get Admin Clubs (with City Filtering)</h3>
        <p>Tests <code>Service.GetAdminClubs()</code></p>
        <p>Only returns Portland cycling clubs where you're admin/owner.</p>
        <button onclick="window.location.href='/test-service-clubs'">Test GetAdminClubs()</button>
    </div>

    <div class="feature">
        <h3>Step 3: Get Club Events (with Upcoming Filter)</h3>
        <p>Tests <code>Service.GetClubEvents()</code> and <code>FilterUpcomingEvents()</code></p>
        <button onclick="window.location.href='/test-service-events'">Test GetClubEvents()</button>
    </div>

    <div class="feature">
        <h3>Step 4: Convert Event to Submission</h3>
        <p>Tests <code>Service.ConvertEventToSubmission()</code> with geocoding fallback</p>
        <p><em>Available after fetching events</em></p>
    </div>

    <h2>M3 Import Tests</h2>
    <div class="feature">
        <h3>Import Preview</h3>
        <p>Tests event conversion and preview without database</p>
        <button onclick="window.location.href='/test-import-select'">Test Import (Preview)</button>
    </div>

    <div class="feature">
        <h3>WebSocket Import (Full)</h3>
        <p>Tests real WebSocket import via the main API (port 8081)</p>
        <p>Requires OAuth on port 8081 first.</p>
        <button onclick="window.location.href='/test-websocket'">WebSocket Test Tool</button>
    </div>
</body>
</html>
	`

	action := r.URL.Query().Get("action")
	city := r.URL.Query().Get("city")

	if action == "oauth" && city != "" {
		// Use M2 Service to initiate OAuth
		authURL, err := service.InitiateOAuth(context.Background(), city)
		if err != nil {
			http.Error(w, "Failed to initiate OAuth: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleTestServiceClubs(w http.ResponseWriter, r *http.Request) {
	if sessionID == "" {
		http.Error(w, "No session. Please run OAuth first: http://localhost:3000/test-service", http.StatusUnauthorized)
		return
	}

	// Test M2 Service.GetAdminClubs()
	ctx := context.Background()
	adminClubs, err := service.GetAdminClubs(ctx, sessionID)
	if err != nil {
		http.Error(w, "Failed to get admin clubs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get session to show city code
	session, _ := service.GetSession(sessionID)

	var htmlOutput string
	htmlOutput += fmt.Sprintf("<h2>✅ M2 Service.GetAdminClubs() Test</h2>")
	htmlOutput += fmt.Sprintf("<div class='info'>")
	htmlOutput += fmt.Sprintf("<strong>Session ID:</strong> %s...<br>", sessionID[:16])
	htmlOutput += fmt.Sprintf("<strong>City Filter:</strong> %s<br>", session.CityCode)
	htmlOutput += fmt.Sprintf("<strong>Athlete ID:</strong> %d", session.AthleteID)
	htmlOutput += fmt.Sprintf("</div>")

	htmlOutput += fmt.Sprintf("<div class='success'>")
	htmlOutput += fmt.Sprintf("<h3>Found %d admin clubs</h3>", len(adminClubs))
	htmlOutput += fmt.Sprintf("<p>These are Portland cycling clubs where you're admin/owner:</p>")
	htmlOutput += fmt.Sprintf("</div>")

	for _, club := range adminClubs {
		role := "Admin"
		if club.Owner {
			role = "Owner"
		}
		htmlOutput += fmt.Sprintf(`
			<div class="club-card">
				<h3>%s (ID: %d)</h3>
				<p><strong>Role:</strong> %s</p>
				<p><strong>Location:</strong> %s, %s</p>
				<p><strong>Members:</strong> %d</p>
				<p><strong>Cycling Club:</strong> %t</p>
				<button onclick="window.location.href='/test-service-events?club_id=%d'">
					Get Events (M2 Service)
				</button>
			</div>
		`, club.Name, club.ID, role, club.City, club.State, club.MemberCount, club.IsCyclingClub(), club.ID)
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>M2 GetAdminClubs Test</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        .club-card { background: #f9f9f9; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #FC4C02; }
        .info { background: #e3f2fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .success { background: #e8f5e9; padding: 15px; border-radius: 5px; margin: 20px 0; }
        h1 { color: #FC4C02; }
        h2 { color: #666; }
    </style>
</head>
<body>
    <h1>M2 Service Layer Test Results</h1>
    <button onclick="window.location.href='/test-service'">Back to Test Menu</button>
    %s
</body>
</html>
	`, htmlOutput)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleTestServiceEvents(w http.ResponseWriter, r *http.Request) {
	if sessionID == "" {
		http.Error(w, "No session. Please run OAuth first: http://localhost:3000/test-service", http.StatusUnauthorized)
		return
	}

	clubIDStr := r.URL.Query().Get("club_id")
	if clubIDStr == "" {
		http.Error(w, "club_id parameter required", http.StatusBadRequest)
		return
	}

	var clubID int64
	fmt.Sscanf(clubIDStr, "%d", &clubID)

	// Test M2 Service.GetClubEvents()
	ctx := context.Background()
	events, err := service.GetClubEvents(ctx, sessionID, clubID)
	if err != nil {
		http.Error(w, "Failed to get club events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter upcoming events (M2 feature)
	upcomingEvents := strava.FilterUpcomingEvents(events)

	var htmlOutput string
	htmlOutput += fmt.Sprintf("<h2>✅ M2 Service.GetClubEvents() Test</h2>")
	htmlOutput += fmt.Sprintf("<div class='info'>")
	htmlOutput += fmt.Sprintf("<strong>Club ID:</strong> %d<br>", clubID)
	htmlOutput += fmt.Sprintf("<strong>Total Events:</strong> %d<br>", len(events))
	htmlOutput += fmt.Sprintf("<strong>Upcoming Events:</strong> %d", len(upcomingEvents))
	htmlOutput += fmt.Sprintf("</div>")

	if len(upcomingEvents) == 0 {
		htmlOutput += `<p><em>No upcoming events for this club.</em></p>`
	}

	for _, event := range upcomingEvents {
		occurrence, _ := event.GetFirstOccurrence()
		tz, _ := event.GetTimezone()
		localTime := occurrence.In(tz)

		htmlOutput += fmt.Sprintf(`
			<div class="event-card">
				<h3>%s (ID: %d)</h3>
				<p><strong>When:</strong> %s</p>
				<p><strong>Timezone:</strong> %s</p>
				<p><strong>Address:</strong> %s</p>
				<p><strong>Has Location:</strong> %t (lat: %.4f, lng: %.4f)</p>
				<p><strong>Route:</strong> %v</p>
				<p><strong>Activity Type:</strong> %s</p>
				<p><strong>Description:</strong> %s</p>
			</div>
		`, event.Title, event.ID, localTime.Format("Mon Jan 2, 2006 at 3:04 PM MST"),
			event.Zone, event.Address, event.HasLocation(), event.GetLatitude(), event.GetLongitude(),
			event.Route != nil, event.ActivityType, event.Description)
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>M2 GetClubEvents Test</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        .event-card { background: #f9f9f9; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #4CAF50; }
        .info { background: #e3f2fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        h1 { color: #FC4C02; }
        h2 { color: #666; }
    </style>
</head>
<body>
    <h1>M2 Service Layer Test Results</h1>
    <button onclick="window.location.href='/test-service-clubs'">Back to Clubs</button>
    <button onclick="window.location.href='/test-service'">Back to Test Menu</button>
    <button onclick="window.location.href='/test-import-select?club_id=%d'">🚀 Test M3 Import</button>
    %s
</body>
</html>
	`, clubID, htmlOutput)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// ============================================================================
// M3 IMPORT TEST ENDPOINTS
// ============================================================================

func handleTestImportSelect(w http.ResponseWriter, r *http.Request) {
	if sessionID == "" {
		http.Error(w, "No session. Please run OAuth first: http://localhost:3000/test-service", http.StatusUnauthorized)
		return
	}

	clubIDStr := r.URL.Query().Get("club_id")
	if clubIDStr == "" {
		// Show club selection first
		ctx := context.Background()
		adminClubs, err := service.GetAdminClubs(ctx, sessionID)
		if err != nil {
			http.Error(w, "Failed to get admin clubs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var htmlOutput string
		htmlOutput += "<h2>Select a Club for Import Test</h2>"
		htmlOutput += "<p>Choose a club to test importing events from:</p>"

		for _, club := range adminClubs {
			htmlOutput += fmt.Sprintf(`
				<div class="club-card">
					<h3>%s (ID: %d)</h3>
					<p><strong>Location:</strong> %s, %s</p>
					<p><strong>Members:</strong> %d</p>
					<button onclick="window.location.href='/test-import-select?club_id=%d'">
						Select This Club
					</button>
				</div>
			`, club.Name, club.ID, club.City, club.State, club.MemberCount, club.ID)
		}

		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>M3 Import Test - Select Club</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        .club-card { background: #f9f9f9; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #FC4C02; }
        h1 { color: #FC4C02; }
    </style>
</head>
<body>
    <h1>🚀 M3 Import Test</h1>
    <button onclick="window.location.href='/test-service'">Back to Test Menu</button>
    %s
</body>
</html>
		`, htmlOutput)

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
		return
	}

	// Show events for the selected club
	var clubID int64
	fmt.Sscanf(clubIDStr, "%d", &clubID)

	ctx := context.Background()
	events, err := service.GetClubEvents(ctx, sessionID, clubID)
	if err != nil {
		http.Error(w, "Failed to get club events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	upcomingEvents := strava.FilterUpcomingEvents(events)

	var htmlOutput string
	htmlOutput += fmt.Sprintf("<h2>Select Events to Import (Club %d)</h2>", clubID)
	htmlOutput += fmt.Sprintf("<p>Found %d upcoming events:</p>", len(upcomingEvents))

	if len(upcomingEvents) == 0 {
		htmlOutput += `<p><em>No upcoming events to import.</em></p>`
	} else {
		htmlOutput += `<form action="/test-import-preview" method="get">`
		htmlOutput += fmt.Sprintf(`<input type="hidden" name="club_id" value="%d">`, clubID)

		for _, event := range upcomingEvents {
			occurrence, _ := event.GetFirstOccurrence()
			tz, _ := event.GetTimezone()
			localTime := occurrence.In(tz)

			hasLocation := "❌ No"
			if event.HasLocation() {
				hasLocation = fmt.Sprintf("✅ Yes (%.4f, %.4f)", event.GetLatitude(), event.GetLongitude())
			}

			hasRoute := "❌ No"
			if event.Route != nil {
				hasRoute = fmt.Sprintf("✅ %s (%.1f km)", event.Route.Name, event.Route.Distance/1000)
			}

			htmlOutput += fmt.Sprintf(`
				<div class="event-card">
					<label style="display: flex; align-items: flex-start; gap: 15px;">
						<input type="checkbox" name="event_id" value="%d" style="margin-top: 5px; width: 20px; height: 20px;">
						<div>
							<h3 style="margin: 0;">%s</h3>
							<p><strong>When:</strong> %s</p>
							<p><strong>Location:</strong> %s</p>
							<p><strong>Has Coordinates:</strong> %s</p>
							<p><strong>Has Route:</strong> %s</p>
							<p style="font-size: 12px; color: #666;">%s</p>
						</div>
					</label>
					<button type="button" onclick="window.location.href='/test-convert?club_id=%d&event_id=%d'" style="margin-top: 10px;">
						Test Convert Only
					</button>
				</div>
			`, event.ID, event.Title, localTime.Format("Mon Jan 2, 2006 at 3:04 PM MST"),
				event.Address, hasLocation, hasRoute, event.Description, clubID, event.ID)
		}

		htmlOutput += `
			<div style="margin-top: 20px; padding: 20px; background: #e8f5e9; border-radius: 5px;">
				<h3>Import Settings</h3>
				<label style="display: block; margin: 10px 0;">
					<strong>Organizer Email:</strong><br>
					<input type="email" name="email" placeholder="your@email.com" style="padding: 10px; width: 300px; margin-top: 5px;" required>
				</label>
				<button type="submit" style="background: #4CAF50; margin-top: 10px;">
					Preview Import
				</button>
			</div>
		</form>`
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>M3 Import Test - Select Events</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        .event-card { background: #f9f9f9; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #4CAF50; }
        h1 { color: #FC4C02; }
        input[type="checkbox"] { cursor: pointer; }
    </style>
</head>
<body>
    <h1>🚀 M3 Import Test - Select Events</h1>
    <button onclick="window.location.href='/test-import-select'">Back to Club Selection</button>
    <button onclick="window.location.href='/test-service'">Back to Test Menu</button>
    %s
</body>
</html>
	`, htmlOutput)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleTestConvert(w http.ResponseWriter, r *http.Request) {
	if sessionID == "" {
		http.Error(w, "No session. Please run OAuth first", http.StatusUnauthorized)
		return
	}

	clubIDStr := r.URL.Query().Get("club_id")
	eventIDStr := r.URL.Query().Get("event_id")

	if clubIDStr == "" || eventIDStr == "" {
		http.Error(w, "club_id and event_id required", http.StatusBadRequest)
		return
	}

	var clubID, eventID int64
	fmt.Sscanf(clubIDStr, "%d", &clubID)
	fmt.Sscanf(eventIDStr, "%d", &eventID)

	ctx := context.Background()

	// Fetch events
	events, err := service.GetClubEvents(ctx, sessionID, clubID)
	if err != nil {
		http.Error(w, "Failed to get events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Find the specific event
	var targetEvent *strava.GroupEvent
	for i := range events {
		if events[i].ID == eventID {
			targetEvent = &events[i]
			break
		}
	}

	if targetEvent == nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Test the conversion
	submission, lat, lng, err := service.ConvertEventToSubmission(ctx, sessionID, targetEvent, "test@example.com")

	var htmlOutput string
	htmlOutput += fmt.Sprintf("<h2>✅ M3 ConvertEventToSubmission Test</h2>")
	htmlOutput += fmt.Sprintf("<div class='info'>")
	htmlOutput += fmt.Sprintf("<strong>Strava Event ID:</strong> %d<br>", eventID)
	htmlOutput += fmt.Sprintf("<strong>Club ID:</strong> %d", clubID)
	htmlOutput += fmt.Sprintf("</div>")

	if err != nil {
		htmlOutput += fmt.Sprintf(`<div class="error"><h3>❌ Conversion Failed</h3><p>%s</p></div>`, err.Error())
	} else {
		htmlOutput += `<div class="success"><h3>✅ Conversion Successful</h3></div>`
		htmlOutput += fmt.Sprintf(`
			<h3>Generated CycleScene Submission:</h3>
			<table class="data-table">
				<tr><td><strong>Title</strong></td><td>%s</td></tr>
				<tr><td><strong>TinyTitle</strong></td><td>%s</td></tr>
				<tr><td><strong>Description</strong></td><td>%s</td></tr>
				<tr><td><strong>City</strong></td><td>%s</td></tr>
				<tr><td><strong>VenueName</strong></td><td>%s</td></tr>
				<tr><td><strong>Address</strong></td><td>%s</td></tr>
				<tr><td><strong>Latitude</strong></td><td>%.6f</td></tr>
				<tr><td><strong>Longitude</strong></td><td>%.6f</td></tr>
				<tr><td><strong>Audience</strong></td><td>%s</td></tr>
				<tr><td><strong>DateType</strong></td><td>%s</td></tr>
				<tr><td><strong>OrganizerEmail</strong></td><td>%s</td></tr>
				<tr><td><strong>WebURL</strong></td><td><a href="%s" target="_blank">%s</a></td></tr>
			</table>
		`, submission.Title, submission.TinyTitle, submission.Description, submission.City,
			submission.VenueName, submission.Address, lat, lng, submission.Audience,
			submission.DateType, submission.OrganizerEmail, submission.WebURL, submission.WebURL)

		if len(submission.Occurrences) > 0 {
			htmlOutput += "<h3>Occurrences:</h3>"
			for i, occ := range submission.Occurrences {
				htmlOutput += fmt.Sprintf(`
					<div class="occurrence">
						<strong>Occurrence %d:</strong><br>
						Date: %s<br>
						Time: %s<br>
						Duration: %d minutes
					</div>
				`, i+1, occ.StartDate, occ.StartTime, occ.EventDurationMinutes)
			}
		}
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>M3 Convert Test</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        .info { background: #e3f2fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .success { background: #e8f5e9; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .error { background: #ffebee; color: #c62828; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .data-table { width: 100%%; border-collapse: collapse; margin: 15px 0; }
        .data-table td { padding: 10px; border: 1px solid #ddd; }
        .data-table tr:nth-child(even) { background: #f9f9f9; }
        .occurrence { background: #fff3e0; padding: 10px; margin: 10px 0; border-radius: 5px; }
        h1 { color: #FC4C02; }
    </style>
</head>
<body>
    <h1>🔄 M3 Event Conversion Test</h1>
    <button onclick="window.location.href='/test-import-select?club_id=%s'">Back to Event Selection</button>
    <button onclick="window.location.href='/test-service'">Back to Test Menu</button>
    %s
</body>
</html>
	`, clubIDStr, htmlOutput)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleTestImportPreview(w http.ResponseWriter, r *http.Request) {
	if sessionID == "" {
		http.Error(w, "No session. Please run OAuth first", http.StatusUnauthorized)
		return
	}

	clubIDStr := r.URL.Query().Get("club_id")
	eventIDs := r.URL.Query()["event_id"]
	email := r.URL.Query().Get("email")

	if clubIDStr == "" || len(eventIDs) == 0 {
		http.Error(w, "club_id and at least one event_id required", http.StatusBadRequest)
		return
	}

	if email == "" {
		email = "test@example.com"
	}

	var clubID int64
	fmt.Sscanf(clubIDStr, "%d", &clubID)

	ctx := context.Background()

	// Fetch events
	events, err := service.GetClubEvents(ctx, sessionID, clubID)
	if err != nil {
		http.Error(w, "Failed to get events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build event ID map for quick lookup
	eventIDMap := make(map[int64]bool)
	for _, idStr := range eventIDs {
		var id int64
		fmt.Sscanf(idStr, "%d", &id)
		eventIDMap[id] = true
	}

	var htmlOutput string
	htmlOutput += fmt.Sprintf("<h2>🚀 M3 Import Preview</h2>")
	htmlOutput += fmt.Sprintf("<div class='info'>")
	htmlOutput += fmt.Sprintf("<strong>Selected Events:</strong> %d<br>", len(eventIDs))
	htmlOutput += fmt.Sprintf("<strong>Organizer Email:</strong> %s<br>", email)
	htmlOutput += fmt.Sprintf("<strong>Club ID:</strong> %d", clubID)
	htmlOutput += fmt.Sprintf("</div>")

	htmlOutput += "<h3>Events to Import:</h3>"

	successCount := 0
	failCount := 0

	for _, event := range events {
		if !eventIDMap[event.ID] {
			continue
		}

		// Test conversion
		submission, lat, lng, err := service.ConvertEventToSubmission(ctx, sessionID, &event, email)

		status := "✅ Ready"
		statusClass := "success"
		details := ""

		if err != nil {
			status = "❌ Error"
			statusClass = "error"
			details = err.Error()
			failCount++
		} else {
			successCount++
			details = fmt.Sprintf(`
				<strong>Title:</strong> %s<br>
				<strong>City:</strong> %s<br>
				<strong>Address:</strong> %s<br>
				<strong>Coordinates:</strong> (%.6f, %.6f)<br>
				<strong>Date:</strong> %s at %s<br>
				<strong>Source:</strong> strava<br>
				<strong>SourceID:</strong> %d_%s
			`, submission.Title, submission.City, submission.Address, lat, lng,
				submission.Occurrences[0].StartDate, submission.Occurrences[0].StartTime, event.ID, submission.Occurrences[0].StartDate)
		}

		htmlOutput += fmt.Sprintf(`
			<div class="import-item %s">
				<h4>%s - %s</h4>
				<p class="status">%s</p>
				<div class="details">%s</div>
			</div>
		`, statusClass, event.Title, status, status, details)
	}

	htmlOutput += fmt.Sprintf(`
		<div class="summary">
			<h3>Summary</h3>
			<p><strong>Ready to Import:</strong> %d events</p>
			<p><strong>Will Fail:</strong> %d events</p>
		</div>
	`, successCount, failCount)

	if successCount > 0 {
		htmlOutput += `
			<div class="action-box">
				<h3>⚠️ Note: This is a Preview Only</h3>
				<p>To actually import events, you would need to:</p>
				<ol>
					<li>Run the full API server with database connection</li>
					<li>Use the WebSocket endpoint at <code>ws://localhost:8080/v1/strava/import</code></li>
					<li>Send a JSON message with session_id, organizer_email, and events array</li>
				</ol>
				<h4>Example WebSocket Message:</h4>
				<pre>{
  "session_id": "` + sessionID[:16] + `...",
  "organizer_email": "` + email + `",
  "events": [`
		for i, idStr := range eventIDs {
			if i > 0 {
				htmlOutput += `,`
			}
			htmlOutput += fmt.Sprintf(`
    {
      "strava_event_id": %s,
      "club_id": %d
    }`, idStr, clubID)
		}
		htmlOutput += `
  ]
}</pre>
			</div>
		`
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>M3 Import Preview</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        .info { background: #e3f2fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .import-item { padding: 15px; margin: 10px 0; border-radius: 5px; }
        .import-item.success { background: #e8f5e9; border-left: 4px solid #4CAF50; }
        .import-item.error { background: #ffebee; border-left: 4px solid #f44336; }
        .import-item h4 { margin: 0 0 10px 0; }
        .details { font-size: 14px; color: #666; }
        .summary { background: #fff3e0; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .action-box { background: #f5f5f5; padding: 20px; border-radius: 5px; margin: 20px 0; }
        pre { background: #263238; color: #aed581; padding: 15px; border-radius: 5px; overflow-x: auto; }
        h1 { color: #FC4C02; }
    </style>
</head>
<body>
    <h1>🚀 M3 Import Preview</h1>
    <button onclick="window.location.href='/test-import-select?club_id=%s'">Back to Event Selection</button>
    <button onclick="window.location.href='/test-service'">Back to Test Menu</button>
    %s
</body>
</html>
	`, clubIDStr, htmlOutput)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleTestWebSocket(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Import Test</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1000px; margin: 50px auto; padding: 20px; }
        button { background: #FC4C02; color: white; border: none; padding: 10px 20px; font-size: 14px; cursor: pointer; border-radius: 5px; margin: 5px; }
        button:hover { background: #E34402; }
        button:disabled { background: #ccc; }
        .info { background: #e3f2fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        #log { background: #263238; color: #aed581; padding: 15px; border-radius: 5px; height: 400px; overflow-y: auto; font-family: monospace; font-size: 12px; }
        .log-entry { margin: 5px 0; padding: 5px; border-bottom: 1px solid #37474f; }
        .log-send { color: #81d4fa; }
        .log-recv { color: #aed581; }
        .log-error { color: #ef9a9a; }
        .log-info { color: #fff59d; }
        textarea { width: 100%; height: 150px; font-family: monospace; font-size: 12px; }
        input { padding: 8px; margin: 5px 0; }
        h1 { color: #FC4C02; }
    </style>
</head>
<body>
    <h1>🔌 WebSocket Import Test</h1>

    <div class="info">
        <strong>Instructions:</strong>
        <ol>
            <li>First authenticate via <a href="http://localhost:8081/v1/strava/auth/initiate?city=pdx" target="_blank">OAuth on port 8081</a></li>
            <li>Copy your session ID from the cookie (browser dev tools → Application → Cookies → strava_session_id)</li>
            <li>Enter the session ID below and connect</li>
        </ol>
    </div>

    <h2>Connection</h2>
    <div>
        <label>WebSocket URL:</label><br>
        <input type="text" id="wsUrl" value="ws://localhost:8081/v1/strava/import" style="width: 400px;">
    </div>
    <div>
        <label>Session ID (from strava_session_id cookie):</label><br>
        <input type="text" id="sessionId" placeholder="paste your session ID here" style="width: 400px;">
    </div>
    <div style="margin-top: 10px;">
        <button id="connectBtn" onclick="connect()">Connect</button>
        <button id="disconnectBtn" onclick="disconnect()" disabled>Disconnect</button>
        <span id="status" style="margin-left: 10px;">Disconnected</span>
    </div>

    <h2>Import Request</h2>
    <div>
        <label>Organizer Email:</label><br>
        <input type="email" id="email" placeholder="your@email.com" style="width: 300px;">
    </div>
    <div>
        <label>Events JSON (IDs must be strings to avoid precision loss):</label><br>
        <textarea id="eventsJson">[
  {
    "strava_event_id": "3452409311648303182",
    "club_id": "1947941"
  }
]</textarea>
    </div>
    <div>
        <button id="sendBtn" onclick="sendImport()" disabled>Send Import Request</button>
    </div>

    <h2>Log</h2>
    <div id="log"></div>
    <button onclick="document.getElementById('log').innerHTML=''">Clear Log</button>

    <script>
        let ws = null;

        function log(message, type) {
            type = type || 'info';
            const logDiv = document.getElementById('log');
            const entry = document.createElement('div');
            entry.className = 'log-entry log-' + type;
            const time = new Date().toLocaleTimeString();
            entry.textContent = '[' + time + '] ' + message;
            logDiv.appendChild(entry);
            logDiv.scrollTop = logDiv.scrollHeight;
        }

        function connect() {
            const url = document.getElementById('wsUrl').value;
            log('Connecting to ' + url + '...', 'info');

            try {
                ws = new WebSocket(url);

                ws.onopen = function() {
                    log('Connected!', 'info');
                    document.getElementById('status').textContent = 'Connected';
                    document.getElementById('status').style.color = 'green';
                    document.getElementById('connectBtn').disabled = true;
                    document.getElementById('disconnectBtn').disabled = false;
                    document.getElementById('sendBtn').disabled = false;
                };

                ws.onmessage = function(event) {
                    log('← ' + event.data, 'recv');
                    try {
                        const msg = JSON.parse(event.data);
                        if (msg.type === 'error') {
                            log('ERROR: ' + msg.message, 'error');
                        } else if (msg.type === 'complete' && msg.success) {
                            log('✅ Event imported: ' + msg.edit_url, 'info');
                        } else if (msg.type === 'done') {
                            log('🎉 Import complete! ' + (msg.total_imported || 0) + ' imported, ' + (msg.total_failed || 0) + ' failed', 'info');
                        }
                    } catch (e) {}
                };

                ws.onerror = function(error) {
                    log('WebSocket error', 'error');
                };

                ws.onclose = function() {
                    log('Disconnected', 'info');
                    document.getElementById('status').textContent = 'Disconnected';
                    document.getElementById('status').style.color = 'red';
                    document.getElementById('connectBtn').disabled = false;
                    document.getElementById('disconnectBtn').disabled = true;
                    document.getElementById('sendBtn').disabled = true;
                    ws = null;
                };
            } catch (e) {
                log('Failed to connect: ' + e, 'error');
            }
        }

        function disconnect() {
            if (ws) {
                ws.close();
            }
        }

        function sendImport() {
            if (!ws) {
                log('Not connected!', 'error');
                return;
            }

            const sessionId = document.getElementById('sessionId').value;
            const email = document.getElementById('email').value;
            const eventsJson = document.getElementById('eventsJson').value;

            if (!sessionId) {
                log('Session ID is required!', 'error');
                return;
            }
            if (!email) {
                log('Email is required!', 'error');
                return;
            }

            let events;
            try {
                events = JSON.parse(eventsJson);
            } catch (e) {
                log('Invalid events JSON: ' + e, 'error');
                return;
            }

            const request = {
                session_id: sessionId,
                organizer_email: email,
                events: events
            };

            const json = JSON.stringify(request, null, 2);
            log('→ ' + json, 'send');
            ws.send(json);
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
