package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var (
	ytShortsRe = regexp.MustCompile(`(?i)youtube\.com/shorts/([A-Za-z0-9_-]+)`)
	ytWatchRe  = regexp.MustCompile(`(?i)youtube\.com/watch\?.*v=([A-Za-z0-9_-]+)`)
	ytShortRe  = regexp.MustCompile(`(?i)youtu\.be/([A-Za-z0-9_-]+)`)
)

// YouTube fetches metrics via the Data API v3 videos.list endpoint.
type YouTube struct {
	APIKey  string
	Client  *http.Client
	BaseURL string
}

// Fetch retrieves view/like/comment counts for a YouTube video.
func (y *YouTube) Fetch(ctx context.Context, postURL string) (Metrics, error) {
	if y.APIKey == "" {
		return Metrics{}, fmt.Errorf("youtube: API key not configured")
	}

	videoID, err := extractYouTubeID(postURL)
	if err != nil {
		return Metrics{}, fmt.Errorf("youtube: %w", err)
	}

	client := y.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	baseURL := y.BaseURL
	if baseURL == "" {
		baseURL = "https://www.googleapis.com/youtube/v3/videos"
	}
	u := fmt.Sprintf(
		"%s?part=statistics,snippet&id=%s&key=%s",
		baseURL,
		url.QueryEscape(videoID), url.QueryEscape(y.APIKey),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Metrics{}, fmt.Errorf("youtube: build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Metrics{}, fmt.Errorf("youtube: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Metrics{}, fmt.Errorf("youtube: API returned %d: %s", resp.StatusCode, string(body))
	}

	var result youtubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Metrics{}, fmt.Errorf("youtube: decode response: %w", err)
	}

	if len(result.Items) == 0 {
		return Metrics{}, fmt.Errorf("youtube: video %s not found or unavailable", videoID)
	}

	stats := result.Items[0].Statistics
	m := Metrics{}
	m.Views = parseInt64(stats.ViewCount)
	m.Likes = parseInt64(stats.LikeCount)
	m.Comments = parseInt64(stats.CommentCount)
	return m, nil
}

func extractYouTubeID(rawURL string) (string, error) {
	// Regexes are case-insensitive, so match directly on raw URL
	if m := ytShortsRe.FindStringSubmatch(rawURL); m != nil {
		return m[1], nil
	}
	if m := ytWatchRe.FindStringSubmatch(rawURL); m != nil {
		return m[1], nil
	}
	if m := ytShortRe.FindStringSubmatch(rawURL); m != nil {
		return m[1], nil
	}
	return "", fmt.Errorf("cannot parse YouTube URL: %s", rawURL)
}

func parseInt64(s string) int64 {
	var n int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		}
	}
	return n
}

type youtubeResponse struct {
	Items []youtubeItem `json:"items"`
}

type youtubeItem struct {
	Statistics youtubeStats `json:"statistics"`
}

type youtubeStats struct {
	ViewCount    string `json:"viewCount"`
	LikeCount    string `json:"likeCount"`
	CommentCount string `json:"commentCount"`
}
