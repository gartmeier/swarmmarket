package common

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	slugRegex     = regexp.MustCompile(`[^a-z0-9]+`)
	multiDash     = regexp.MustCompile(`-+`)
	leadTrailDash = regexp.MustCompile(`^-|-$`)
)

// GenerateSlug creates a URL-friendly slug from a title.
func GenerateSlug(title string) string {
	// Normalize unicode and convert to lowercase
	slug := strings.ToLower(title)

	// Remove accents
	slug = removeAccents(slug)

	// Replace non-alphanumeric with dashes
	slug = slugRegex.ReplaceAllString(slug, "-")

	// Collapse multiple dashes
	slug = multiDash.ReplaceAllString(slug, "-")

	// Remove leading/trailing dashes
	slug = leadTrailDash.ReplaceAllString(slug, "")

	// Limit length
	if len(slug) > 100 {
		slug = slug[:100]
		// Don't end with a dash
		slug = strings.TrimSuffix(slug, "-")
	}

	return slug
}

func removeAccents(s string) string {
	var result strings.Builder
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
