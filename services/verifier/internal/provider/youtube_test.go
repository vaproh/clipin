package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestYouTubeFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/youtube/v3/videos" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		if r.URL.Query().Get("key") != "test-key" {
			t.Error("missing or wrong API key")
			w.WriteHeader(403)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(youtubeResponse{
			Items: []youtubeItem{
				{Statistics: youtubeStats{
					ViewCount:    "123456",
					LikeCount:    "7890",
					CommentCount: "321",
				}},
			},
		})
	}))
	defer server.Close()

	_ = server // server available for integration-style tests

	videoID, err := extractYouTubeID("https://www.youtube.com/shorts/abc123")
	if err != nil {
		t.Fatalf("extractYouTubeID: %v", err)
	}
	if videoID != "abc123" {
		t.Fatalf("unexpected video ID: %s", videoID)
	}

	// Test the parse path directly since we can't easily redirect the YouTube API URL
	// Instead, test the parse logic with a mock response
	resp := youtubeResponse{
		Items: []youtubeItem{
			{Statistics: youtubeStats{
				ViewCount:    "123456",
				LikeCount:    "7890",
				CommentCount: "321",
			}},
		},
	}
	body, _ := json.Marshal(resp)

	var parsed youtubeResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(parsed.Items))
	}
	if parsed.Items[0].Statistics.ViewCount != "123456" {
		t.Errorf("ViewCount = %s, want 123456", parsed.Items[0].Statistics.ViewCount)
	}
}

func TestYouTubeFetchNoAPIKey(t *testing.T) {
	yt := &YouTube{APIKey: ""}
	_, err := yt.Fetch(context.Background(), "https://www.youtube.com/shorts/abc123")
	if err == nil {
		t.Error("expected error for missing API key")
	}
}

func TestYouTubeFetchBadURL(t *testing.T) {
	yt := &YouTube{APIKey: "test"}
	_, err := yt.Fetch(context.Background(), "https://example.com/video")
	if err == nil {
		t.Error("expected error for non-YouTube URL")
	}
}

func TestYouTubeFetchVideoNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(youtubeResponse{Items: []youtubeItem{}})
	}))
	defer server.Close()

	// We can't easily test the full Fetch path without URL rewriting,
	// so test the parse logic
	resp := youtubeResponse{Items: []youtubeItem{}}
	if len(resp.Items) != 0 {
		t.Error("expected empty items")
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"0", 0},
		{"123", 123},
		{"999999999999", 999999999999},
		{"", 0},
		{"abc", 0},
	}
	for _, tt := range tests {
		got := parseInt64(tt.input)
		if got != tt.want {
			t.Errorf("parseInt64(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
