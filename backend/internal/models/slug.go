package models

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// GenerateSlug creates a URL-safe slug from a display name.
// If the slug is already taken, it appends a random 4-char suffix.
// minLength is the minimum random suffix length appended when there's a conflict.
func GenerateSlug(name string) string {
	slug := slugify(name)
	if slug == "" {
		slug = "creator"
	}
	return slug
}

// AppendRandomSuffix adds a short random suffix to resolve slug conflicts.
// Keeps the slug readable while ensuring uniqueness.
func AppendRandomSuffix(slug string) string {
	suffix := randomHex(4)
	// Trim slug to leave room for the suffix
	maxLen := 80
	if len(slug) > maxLen {
		slug = slug[:maxLen]
	}
	// Trim trailing hyphens before adding suffix
	slug = strings.TrimRight(slug, "-")
	return slug + "-" + suffix
}

func slugify(s string) string {
	slug := strings.ToLower(s)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric, non-hyphen characters
	cleaned := make([]byte, 0, len(slug))
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			cleaned = append(cleaned, c)
		}
	}
	slug = string(cleaned)

	// Collapse consecutive hyphens
	result := make([]byte, 0, len(slug))
	lastHyphen := false
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if c == '-' {
			if lastHyphen {
				continue
			}
			lastHyphen = true
		} else {
			lastHyphen = false
		}
		result = append(result, c)
	}
	slug = string(result)
	slug = strings.Trim(slug, "-")
	return slug
}

func randomHex(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}