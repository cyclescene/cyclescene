package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func NilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

var cityTimeZones = map[string]string{
	"pdx": "America/Los_Angeles",
	"slc": "America/Denver",
}

func GetTimeZone(cityCode string) string {
	if tz, ok := cityTimeZones[strings.ToLower(cityCode)]; ok {
		return tz
	}
	return "America/Los_Angeles"
}
