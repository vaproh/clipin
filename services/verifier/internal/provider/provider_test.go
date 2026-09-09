package provider

import (
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.youtube.com/shorts/abc123", PlatformYouTube},
		{"https://youtube.com/watch?v=abc123", PlatformYouTube},
		{"https://youtu.be/abc123", PlatformYouTube},
		{"https://www.instagram.com/reels/XYZ12/", PlatformInstagram},
		{"https://www.instagram.com/p/XYZ12/", PlatformInstagram},
		{"https://example.com/video", PlatformUnknown},
		{"https://notinstagram.com/reels/XYZ12/", PlatformUnknown},
		{"https://example.com/?next=https://instagram.com/reels/XYZ12/", PlatformUnknown},
		{"", PlatformUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := DetectPlatform(tt.url)
			if got != tt.expected {
				t.Errorf("DetectPlatform(%q) = %q, want %q", tt.url, got, tt.expected)
			}
		})
	}
}

func TestExtractYouTubeID(t *testing.T) {
	tests := []struct {
		url   string
		want  string
		error bool
	}{
		{"https://www.youtube.com/shorts/abc123", "abc123", false},
		{"https://youtube.com/shorts/abc123", "abc123", false},
		{"https://www.youtube.com/watch?v=abc123&si=xyz", "abc123", false},
		{"https://youtube.com/watch?v=abc123&t=30", "abc123", false},
		{"https://youtu.be/abc123", "abc123", false},
		{"https://youtu.be/abc123?si=xyz", "abc123", false},
		{"https://example.com/video", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got, err := extractYouTubeID(tt.url)
			if (err != nil) != tt.error {
				t.Errorf("extractYouTubeID(%q) error = %v, wantErr %v", tt.url, err, tt.error)
				return
			}
			if got != tt.want {
				t.Errorf("extractYouTubeID(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestParseIGMedia(t *testing.T) {
	media := map[string]interface{}{
		"play_count":    float64(12345),
		"like_count":    float64(678),
		"comment_count": float64(90),
	}
	m := parseIGMedia(media)
	if m.Views != 12345 {
		t.Errorf("Views = %d, want 12345", m.Views)
	}
	if m.Likes != 678 {
		t.Errorf("Likes = %d, want 678", m.Likes)
	}
	if m.Comments != 90 {
		t.Errorf("Comments = %d, want 90", m.Comments)
	}
	if m.Shares != 0 {
		t.Errorf("Shares = %d, want 0", m.Shares)
	}
}
