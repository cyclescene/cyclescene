package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	// Load .env file if in dev mode
	if os.Getenv("APP_ENV") == "dev" {
		_ = godotenv.Load()
	}
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		generateKey(os.Args[2:])
	case "revoke":
		revokeKey(os.Args[2:])
	case "list":
		listKeys(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Admin API Keys Management

Usage:
  admin-keys generate -name <name>    Generate a new API key
  admin-keys revoke -key <api-key>    Revoke an API key
  admin-keys list                      List all active API keys

Environment Variables:
  TURSO_DB_URL       - Turso database URL
  TURSO_DB_RW_TOKEN  - Turso database auth token

Examples:
  admin-keys generate -name "jd"
  admin-keys revoke -key "abc123..."
  admin-keys list

For local SQLite development, set:
  export TURSO_DB_URL="file:./app.db"
  export TURSO_DB_RW_TOKEN=""
`)
}

func connectToDB() (*sql.DB, error) {
	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_DB_RW_TOKEN")

	if dbURL == "" {
		return nil, fmt.Errorf("TURSO_DB_URL environment variable not set")
	}

	fullURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
	return sql.Open("libsql", fullURL)
}

func generateKey(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	adminName := fs.String("name", "", "Admin name")
	fs.Parse(args)

	if *adminName == "" {
		fmt.Println("Error: -name flag is required")
		os.Exit(1)
	}

	// Generate random API key (raw bytes for hashing)
	rawKey := make([]byte, 32)
	_, err := rand.Read(rawKey)
	if err != nil {
		log.Fatalf("Failed to generate random key: %v", err)
	}

	// Hash the raw key for storage
	hashedKey, err := bcrypt.GenerateFromPassword(rawKey, bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash API key: %v", err)
	}

	// Display plaintext key to user (base64 encoded for readability)
	apiKey := base64.URLEncoding.EncodeToString(rawKey)

	// Connect to database using Turso credentials from environment
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert into admin_api_keys table (store hashed key)
	result, err := db.Exec(`
		INSERT INTO admin_api_keys (api_key, admin_name, created_at)
		VALUES (?, ?, ?)
	`, string(hashedKey), *adminName, time.Now().Format(time.RFC3339))

	if err != nil {
		log.Fatalf("Failed to insert API key: %v", err)
	}

	id, _ := result.LastInsertId()

	fmt.Printf("✓ API Key generated successfully\n")
	fmt.Printf("  ID:         %d\n", id)
	fmt.Printf("  Admin:      %s\n", *adminName)
	fmt.Printf("  API Key:    %s\n", apiKey)
	fmt.Printf("  Created:    %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("\nSave this API key in a secure location. You won't be able to see it again.\n")
}

func revokeKey(args []string) {
	fs := flag.NewFlagSet("revoke", flag.ExitOnError)
	apiKey := fs.String("key", "", "API key to revoke")
	fs.Parse(args)

	if *apiKey == "" {
		fmt.Println("Error: -key flag is required")
		os.Exit(1)
	}

	// Connect to database using Turso credentials from environment
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if key exists
	var id int
	var adminName string
	err = db.QueryRow(`SELECT id, admin_name FROM admin_api_keys WHERE api_key = ?`, *apiKey).Scan(&id, &adminName)
	if err == sql.ErrNoRows {
		fmt.Printf("✗ API key not found\n")
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("Failed to query API key: %v", err)
	}

	// Revoke the key
	_, err = db.Exec(`
		UPDATE admin_api_keys
		SET revoked_at = ?
		WHERE api_key = ?
	`, time.Now().Format(time.RFC3339), *apiKey)

	if err != nil {
		log.Fatalf("Failed to revoke API key: %v", err)
	}

	fmt.Printf("✓ API Key revoked successfully\n")
	fmt.Printf("  ID:    %d\n", id)
	fmt.Printf("  Admin: %s\n", adminName)
}

func listKeys(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	fs.Parse(args)

	// Connect to database using Turso credentials from environment
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Query all keys
	rows, err := db.Query(`
		SELECT id, api_key, admin_name, created_at, revoked_at, last_used_at
		FROM admin_api_keys
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Fatalf("Failed to query API keys: %v", err)
	}
	defer rows.Close()

	fmt.Println("Active API Keys:")
	fmt.Println("───────────────────────────────────────────────────────────────")

	count := 0
	for rows.Next() {
		var id int
		var apiKey, adminName string
		var createdAt, revokedAt, lastUsedAt sql.NullString

		err := rows.Scan(&id, &apiKey, &adminName, &createdAt, &revokedAt, &lastUsedAt)
		if err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		status := "✓ Active"
		if revokedAt.Valid {
			status = "✗ Revoked"
		}

		fmt.Printf("[%d] %s\n", id, status)
		fmt.Printf("    Admin:       %s\n", adminName)
		fmt.Printf("    API Key:     %s...%s\n", apiKey[:8], apiKey[len(apiKey)-8:])
		fmt.Printf("    Created:     %s\n", createdAt.String)
		if lastUsedAt.Valid {
			fmt.Printf("    Last Used:   %s\n", lastUsedAt.String)
		}
		fmt.Println()
		count++
	}

	if count == 0 {
		fmt.Println("No API keys found")
	} else {
		fmt.Printf("Total: %d key(s)\n", count)
	}
}

func generateRandomKey(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("Failed to generate random key: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}
