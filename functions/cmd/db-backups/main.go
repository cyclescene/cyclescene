package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// DB Vars
	if os.Getenv("TURSO_DB_HTTP_URL") == "" || os.Getenv("TURSO_DB_RW_TOKEN") == "" {
		log.Fatal("FATAL: Turso env variable not set properly")
	}

	if os.Getenv("BACKUP_BUCKET") == "" {
		log.Fatalf("FATAL: Backup Bucket not set up properly")
	}

	// // set up logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))

	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		slog.Error("Failed to create GCP Storage Client", "error", err)
		log.Fatalf("Failed to create GCP Storage Client: %v", err)
	}

	// 2. Fetch SQL Dump from Turso (HTTP GET)
	dumpURL := fmt.Sprintf("%s/dump", os.Getenv("TURSO_DB_HTTP_URL"))
	slog.Info("Attempting to get Turso dump", "url", dumpURL)
	req, err := http.NewRequestWithContext(ctx, "GET", dumpURL, nil)
	if err != nil {
		slog.Error("Failed to create dump request")
		log.Fatalf("Failed to create dump request: %v", err)
	}
	// Authenticate the request using the Turso Auth Token
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TURSO_DB_RW_TOKEN")))

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error("failed to get dump from Turso", "error", err)
		log.Fatalf("failed to get dump from Turso: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("turso dump failed with status: %d", resp.StatusCode)
	}

	// 3. Define the GCS object name
	objectName := fmt.Sprintf("main-db/dump-%s.sql", time.Now().Format("20060102-150405"))

	// 4. Stream the Response Body (SQL Dump) directly to GCS
	// Create a GCS Writer
	wc := gcsClient.Bucket(os.Getenv("BACKUP_BUCKET")).Object(objectName).NewWriter(ctx)
	wc.ContentType = "text/plain" // SQL dump is plain text

	// 5. io.Copy and Close with Combined Error Check
	copyErr := func() error {
		if _, err := io.Copy(wc, resp.Body); err != nil {
			// Error during stream transfer
			return fmt.Errorf("failed to copy dump to GCS: %w", err)
		}
		// Finalize the upload
		if err := wc.Close(); err != nil {
			// Error during finalization (often due to the copy error)
			return fmt.Errorf("failed to finalize GCS writer: %w", err)
		}
		return nil
	}()

	if copyErr != nil {
		slog.Error("Backup failed", "error", copyErr)
		log.Fatalf("Backup process failed: %v", copyErr)
	}

	slog.Info("Backup successfull", "bucket", os.Getenv("BACKUP_BUCKET"), "object", objectName)

}
