package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

func TestInstagramFetch(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("bootstrap: expected GET, got %s", r.Method)
			w.WriteHeader(405)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "csrftoken", Value: "test-csrf-token", Path: "/"})
		w.WriteHeader(200)
		w.Write([]byte("<html></html>"))
	})

	callCount := 0
	mux.HandleFunc("/graphql/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("graphql: expected POST, got %s", r.Method)
			w.WriteHeader(405)
			return
		}
		csrf := r.Header.Get("X-CSRFToken")
		if csrf != "test-csrf-token" {
			t.Errorf("graphql: wrong CSRF token: %s", csrf)
		}
		appID := r.Header.Get("X-IG-App-ID")
		if appID != igAppID {
			t.Errorf("graphql: wrong app ID: %s", appID)
		}

		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"xdt_api__v1__media__shortcode__web_info": map[string]interface{}{
						"items": []interface{}{
							map[string]interface{}{
								"code": "ABC12",
								"user": map[string]interface{}{
									"pk":       float64(12345678),
									"username": "testuser",
								},
								"taken_at": float64(1700000000),
							},
						},
					},
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"xdt_api__v1__clips__user__connection_v2": map[string]interface{}{
						"edges": []interface{}{
							map[string]interface{}{
								"node": map[string]interface{}{
									"media": map[string]interface{}{
										"code":          "ABC12",
										"play_count":    float64(99999),
										"like_count":    float64(5000),
										"comment_count": float64(250),
									},
								},
							},
							map[string]interface{}{
								"node": map[string]interface{}{
									"media": map[string]interface{}{
										"code":          "OTHER1",
										"play_count":    float64(100),
										"like_count":    float64(10),
										"comment_count": float64(5),
									},
								},
							},
						},
					},
				},
			})
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	client := server.Client()
	client.Jar = jar
	base := server.URL

	if err := igBootstrap(context.Background(), client, base); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	item, err := igPostMetadata(context.Background(), client, base, "ABC12")
	if err != nil {
		t.Fatalf("post metadata: %v", err)
	}
	user, _ := item["user"].(map[string]interface{})
	if user == nil {
		t.Fatal("no user in item")
	}
	if user["username"] != "testuser" {
		t.Errorf("username = %v, want testuser", user["username"])
	}

	medias, err := igClipsFeed(context.Background(), client, base, "12345678", 12)
	if err != nil {
		t.Fatalf("clips feed: %v", err)
	}
	if len(medias) != 2 {
		t.Fatalf("expected 2 medias, got %d", len(medias))
	}

	found := false
	for _, m := range medias {
		code, _ := m["code"].(string)
		if code == "ABC12" {
			found = true
			metrics := parseIGMedia(m)
			if metrics.Views != 99999 {
				t.Errorf("Views = %d, want 99999", metrics.Views)
			}
			if metrics.Likes != 5000 {
				t.Errorf("Likes = %d, want 5000", metrics.Likes)
			}
			if metrics.Comments != 250 {
				t.Errorf("Comments = %d, want 250", metrics.Comments)
			}
		}
	}
	if !found {
		t.Error("target shortcode ABC12 not found in clips feed")
	}
}

func TestInstagramGraphQL429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.SetCookie(w, &http.Cookie{Name: "csrftoken", Value: "tok", Path: "/"})
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(429)
	}))
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	client := server.Client()
	client.Jar = jar

	_ = igBootstrap(context.Background(), client, server.URL)

	_, err := igGraphQL(context.Background(), client, server.URL, "123", map[string]interface{}{"shortcode": "test"})
	if err == nil {
		t.Error("expected error for 429")
	}
}

func TestInstagramGraphQLServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.SetCookie(w, &http.Cookie{Name: "csrftoken", Value: "tok", Path: "/"})
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(500)
	}))
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	client := server.Client()
	client.Jar = jar

	_ = igBootstrap(context.Background(), client, server.URL)

	_, err := igGraphQL(context.Background(), client, server.URL, "123", map[string]interface{}{"shortcode": "test"})
	if err == nil {
		t.Error("expected error for 500")
	}
}

func TestInstagramPostNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.SetCookie(w, &http.Cookie{Name: "csrftoken", Value: "tok", Path: "/"})
			w.WriteHeader(200)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"xdt_api__v1__media__shortcode__web_info": map[string]interface{}{
					"items": []interface{}{},
				},
			},
		})
	}))
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	client := server.Client()
	client.Jar = jar

	_ = igBootstrap(context.Background(), client, server.URL)

	_, err := igPostMetadata(context.Background(), client, server.URL, "NONEXISTENT")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestInstagramGraphQLGraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.SetCookie(w, &http.Cookie{Name: "csrftoken", Value: "tok", Path: "/"})
			w.WriteHeader(200)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []interface{}{
				map[string]interface{}{
					"message": "execution error",
				},
			},
		})
	}))
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	client := server.Client()
	client.Jar = jar

	_ = igBootstrap(context.Background(), client, server.URL)

	_, err := igGraphQL(context.Background(), client, server.URL, "123", map[string]interface{}{"shortcode": "test"})
	if err == nil {
		t.Error("expected error for graphql error")
	}
}

func TestInstagramShortcodeExtraction(t *testing.T) {
	tests := []struct {
		url   string
		want  string
		error bool
	}{
		{"ABC12", "ABC12", false},
		{"https://www.instagram.com/reels/ABC12/", "ABC12", false},
		{"https://www.instagram.com/reel/ABC12/", "ABC12", false},
		{"https://www.instagram.com/p/ABC12/", "ABC12", false},
		{"https://www.instagram.com/tv/ABC12/", "ABC12", false},
		{"https://www.instagram.com/reels/ABC12/?igshid=xyz", "ABC12", false},
		{"abc12", "abc12", false},
		{"ab", "", true},
		{"https://example.com", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got, err := extractIGShortcode(tt.url)
			if (err != nil) != tt.error {
				t.Errorf("extractIGShortcode(%q) error = %v, wantErr %v", tt.url, err, tt.error)
			}
			if got != tt.want {
				t.Errorf("extractIGShortcode(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestInstagramBootstrapFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	err := igBootstrap(context.Background(), server.Client(), server.URL)
	if err == nil {
		t.Error("expected error for bootstrap failure")
	}
}

func TestInstagramNoUserPK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.SetCookie(w, &http.Cookie{Name: "csrftoken", Value: "tok", Path: "/"})
			w.WriteHeader(200)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"xdt_api__v1__media__shortcode__web_info": map[string]interface{}{
					"items": []interface{}{
						map[string]interface{}{
							"code": "ABC12",
							"user": map[string]interface{}{
								"pk": nil,
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	jar, _ := cookiejar.New(nil)
	client := server.Client()
	client.Jar = jar

	_ = igBootstrap(context.Background(), client, server.URL)

	item, err := igPostMetadata(context.Background(), client, server.URL, "ABC12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, _ := item["user"].(map[string]interface{})
	if user != nil {
		pk, _ := user["pk"].(float64)
		if pk != 0 {
			t.Errorf("expected zero pk, got %v", pk)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("truncate('hello', 10) = %q, want 'hello'", got)
	}
	if got := truncate("hello world", 5); got != "hello" {
		t.Errorf("truncate('hello world', 5) = %q, want 'hello'", got)
	}
}

func TestInstagramFetchIntegration(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "csrftoken", Value: "test-csrf", Path: "/"})
		w.WriteHeader(200)
		w.Write([]byte("<html></html>"))
	})

	callCount := 0
	mux.HandleFunc("/graphql/query", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"xdt_api__v1__media__shortcode__web_info": map[string]interface{}{
						"items": []interface{}{
							map[string]interface{}{
								"code": "TEST1",
								"user": map[string]interface{}{
									"pk":       float64(99999),
									"username": "integration",
								},
							},
						},
					},
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"xdt_api__v1__clips__user__connection_v2": map[string]interface{}{
						"edges": []interface{}{
							map[string]interface{}{
								"node": map[string]interface{}{
									"media": map[string]interface{}{
										"code":          "TEST1",
										"play_count":    float64(54321),
										"like_count":    float64(1234),
										"comment_count": float64(56),
									},
								},
							},
						},
					},
				},
			})
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ig := &Instagram{BaseURL: server.URL}
	metrics, err := ig.Fetch(context.Background(), "https://www.instagram.com/reels/TEST1/")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if metrics.Views != 54321 {
		t.Errorf("Views = %d, want 54321", metrics.Views)
	}
	if metrics.Likes != 1234 {
		t.Errorf("Likes = %d, want 1234", metrics.Likes)
	}
	if metrics.Comments != 56 {
		t.Errorf("Comments = %d, want 56", metrics.Comments)
	}
}
