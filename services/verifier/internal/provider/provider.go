package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Metrics holds the normalized social metrics for a post.
type Metrics struct {
	Views    int64
	Likes    int64
	Comments int64
	Shares   int64
}

// Provider fetches metrics for a given post URL.
type Provider interface {
	Fetch(ctx context.Context, postURL string) (Metrics, error)
}

// Platform identifiers.
const (
	PlatformYouTube   = "youtube"
	PlatformInstagram = "instagram"
	PlatformUnknown   = "unknown"
)

// DetectPlatform determines the platform from a URL.
func DetectPlatform(url string) string {
	parsed, err := netURL(url)
	if err != nil {
		return PlatformUnknown
	}
	host := strings.ToLower(parsed)
	switch {
	case host == "youtube.com" || strings.HasSuffix(host, ".youtube.com") || host == "youtu.be":
		return PlatformYouTube
	case host == "instagram.com" || strings.HasSuffix(host, ".instagram.com"):
		return PlatformInstagram
	default:
		return PlatformUnknown
	}
}

func netURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" {
		return "", fmt.Errorf("invalid URL")
	}
	return parsed.Hostname(), nil
}

// RouteToProvider returns the appropriate provider for a URL.
func RouteToProvider(url string, ytKey string, client *http.Client) Provider {
	switch DetectPlatform(url) {
	case PlatformYouTube:
		return &YouTube{APIKey: ytKey, Client: client}
	case PlatformInstagram:
		return &Instagram{Client: client}
	default:
		return nil
	}
}

// PlatformNotSupported is returned when the platform cannot be determined or supported.
type PlatformNotSupported struct {
	URL string
}

func (e *PlatformNotSupported) Error() string {
	return fmt.Sprintf("unsupported platform for URL: %s", e.URL)
}

// shortCodeRe matches bare Instagram shortcodes.
var shortCodeRe = regexp.MustCompile(`^[A-Za-z0-9_-]{5,}$`)
