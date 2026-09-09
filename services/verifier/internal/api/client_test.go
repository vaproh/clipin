package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListPending(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(405)
			return
		}
		if r.URL.Path != "/internal/snapshots/list" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		if r.Header.Get("X-Verifier-Key") != "test-key" {
			t.Errorf("missing or wrong API key: %s", r.Header.Get("X-Verifier-Key"))
			w.WriteHeader(401)
			return
		}

		var body struct {
			Limit int `json:"limit"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.Limit != 10 {
			t.Errorf("limit = %d, want 10", body.Limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(listResponse{
			Submissions: []Submission{
				{SubmissionID: "abc123", PostURL: "https://youtube.com/shorts/xyz", Platform: "youtube"},
				{SubmissionID: "def456", PostURL: "https://instagram.com/reels/abc/", Platform: "instagram"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 5*time.Second)
	subs, err := client.ListPending(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 submissions, got %d", len(subs))
	}
	if subs[0].SubmissionID != "abc123" {
		t.Errorf("first submission ID = %s, want abc123", subs[0].SubmissionID)
	}
	if subs[1].PostURL != "https://instagram.com/reels/abc/" {
		t.Errorf("second post URL = %s", subs[1].PostURL)
	}
}

func TestListPendingNoAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Verifier-Key") != "" {
			w.WriteHeader(401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(listResponse{Submissions: []Submission{}})
	}))
	defer server.Close()

	// No API key
	client := NewClient(server.URL, "", 5*time.Second)
	subs, err := client.ListPending(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("expected 0 submissions, got %d", len(subs))
	}
}

func TestListPendingNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "key", 5*time.Second)
	_, err := client.ListPending(context.Background(), 10)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestRecordSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(405)
			return
		}
		if r.URL.Path != "/internal/snapshots" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		if r.Header.Get("X-Verifier-Key") != "test-key" {
			t.Errorf("missing or wrong API key")
			w.WriteHeader(401)
			return
		}

		var snap Snapshot
		if err := json.NewDecoder(r.Body).Decode(&snap); err != nil {
			t.Errorf("decode body: %v", err)
			w.WriteHeader(400)
			return
		}
		if snap.SubmissionID != "abc123" {
			t.Errorf("submission_id = %s, want abc123", snap.SubmissionID)
		}
		if snap.Views != 99999 {
			t.Errorf("views = %d, want 99999", snap.Views)
		}
		if snap.CapturedAt == "" {
			t.Error("captured_at should not be empty")
		}

		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", 5*time.Second)
	snap := Snapshot{
		SubmissionID: "abc123",
		Platform:     "youtube",
		Views:        99999,
		Likes:        5000,
		Comments:     300,
		Shares:       50,
		CapturedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if err := client.RecordSnapshot(context.Background(), snap); err != nil {
		t.Fatalf("RecordSnapshot: %v", err)
	}
}

func TestRecordSnapshotNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte("unprocessable"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "key", 5*time.Second)
	snap := Snapshot{
		SubmissionID: "abc123",
		Platform:     "youtube",
		Views:        100,
		CapturedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	err := client.RecordSnapshot(context.Background(), snap)
	if err == nil {
		t.Error("expected error for 422 response")
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8080", "key", 10*time.Second)
	if c.BaseURL != "http://localhost:8080" {
		t.Errorf("BaseURL = %s", c.BaseURL)
	}
	if c.APIKey != "key" {
		t.Errorf("APIKey = %s", c.APIKey)
	}
}
