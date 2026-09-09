package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	igDefaultBase = "https://www.instagram.com"
	igUA          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
	igDocPost     = "27128499623469141"
	igDocClips    = "27234427476213202"
	igAppID       = "936619743392459"
	igASBDID      = "129477"
)

var (
	igReelsRe = regexp.MustCompile(`instagram\.com/reels?/([A-Za-z0-9_-]+)`)
	igPostRe  = regexp.MustCompile(`instagram\.com/p/([A-Za-z0-9_-]+)`)
	igTvRe    = regexp.MustCompile(`instagram\.com/tv/([A-Za-z0-9_-]+)`)
)

// Instagram implements the anonymous 3-call GraphQL protocol for view counts.
type Instagram struct {
	Client  *http.Client
	BaseURL string // override for testing; defaults to https://www.instagram.com
}

func (ig *Instagram) base() string {
	if ig.BaseURL != "" {
		return ig.BaseURL
	}
	return igDefaultBase
}

// Fetch retrieves view/like/comment counts for an Instagram reel/post.
func (ig *Instagram) Fetch(ctx context.Context, postURL string) (Metrics, error) {
	shortcode, err := extractIGShortcode(postURL)
	if err != nil {
		return Metrics{}, fmt.Errorf("instagram: %w", err)
	}

	client := ig.makeClient()
	jar, _ := cookiejar.New(nil)
	client.Jar = jar

	base := ig.base()

	if err := igBootstrap(ctx, client, base); err != nil {
		return Metrics{}, fmt.Errorf("instagram: bootstrap: %w", err)
	}

	item, err := igPostMetadata(ctx, client, base, shortcode)
	if err != nil {
		return Metrics{}, fmt.Errorf("instagram: post metadata: %w", err)
	}

	user, _ := item["user"].(map[string]interface{})
	if user == nil {
		return Metrics{}, fmt.Errorf("instagram: no owner for shortcode %s", shortcode)
	}

	pkRaw, _ := user["pk"].(float64)
	if pkRaw == 0 {
		return Metrics{}, fmt.Errorf("instagram: no owner pk for shortcode %s", shortcode)
	}
	pkStr := fmt.Sprintf("%.0f", pkRaw)

	for _, pageSize := range []int{12, 50} {
		medias, err := igClipsFeed(ctx, client, base, pkStr, pageSize)
		if err != nil {
			return Metrics{}, fmt.Errorf("instagram: clips feed: %w", err)
		}
		for _, media := range medias {
			code, _ := media["code"].(string)
			if code == shortcode {
				return parseIGMedia(media), nil
			}
		}
	}

	return Metrics{}, fmt.Errorf("instagram: reel %s not found in owner's clips feed", shortcode)
}

func (ig *Instagram) makeClient() *http.Client {
	if ig.Client != nil {
		c := *ig.Client
		c.Jar = nil
		return &c
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func igBootstrap(ctx context.Context, client *http.Client, base string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", igUA)
	req.Header.Set("Accept-Language", "en-US")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("bootstrap GET failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bootstrap returned HTTP %d", resp.StatusCode)
	}

	return nil
}

func igPostMetadata(ctx context.Context, client *http.Client, base, shortcode string) (map[string]interface{}, error) {
	vars := map[string]interface{}{
		"shortcode": shortcode,
		"__relay_internal__pv__PolarisAIGMMediaWebLabelEnabledrelayprovider": false,
	}
	data, err := igGraphQL(ctx, client, base, igDocPost, vars)
	if err != nil {
		return nil, err
	}

	inner, _ := data["data"].(map[string]interface{})
	webInfo, _ := inner["xdt_api__v1__media__shortcode__web_info"].(map[string]interface{})
	items, _ := webInfo["items"].([]interface{})
	if len(items) == 0 {
		return nil, fmt.Errorf("post %s not found (private or deleted)", shortcode)
	}
	item, _ := items[0].(map[string]interface{})
	if item == nil {
		return nil, fmt.Errorf("post %s returned invalid item", shortcode)
	}
	return item, nil
}

func igClipsFeed(ctx context.Context, client *http.Client, base, userPK string, pageSize int) ([]map[string]interface{}, error) {
	vars := map[string]interface{}{
		"data": map[string]interface{}{
			"include_feed_video": true,
			"page_size":          pageSize,
			"target_user_id":     userPK,
		},
	}
	data, err := igGraphQL(ctx, client, base, igDocClips, vars)
	if err != nil {
		return nil, err
	}

	inner, _ := data["data"].(map[string]interface{})
	conn, _ := inner["xdt_api__v1__clips__user__connection_v2"].(map[string]interface{})
	edges, _ := conn["edges"].([]interface{})

	var medias []map[string]interface{}
	for _, edge := range edges {
		e, _ := edge.(map[string]interface{})
		if e == nil {
			continue
		}
		node, _ := e["node"].(map[string]interface{})
		if node == nil {
			continue
		}
		media, _ := node["media"].(map[string]interface{})
		if media != nil {
			medias = append(medias, media)
		}
	}
	return medias, nil
}

func igGraphQL(ctx context.Context, client *http.Client, base, docID string, variables map[string]interface{}) (map[string]interface{}, error) {
	varsJSON, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("marshal variables: %w", err)
	}

	form := url.Values{}
	form.Set("variables", string(varsJSON))
	form.Set("doc_id", docID)
	form.Set("server_timestamps", "true")

	var csrfToken string
	if client.Jar != nil {
		u, _ := url.Parse(base)
		for _, c := range client.Jar.Cookies(u) {
			if c.Name == "csrftoken" {
				csrfToken = c.Value
				break
			}
		}
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/graphql/query", strings.NewReader(form.Encode()))
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("User-Agent", igUA)
		req.Header.Set("Accept-Language", "en-US")
		req.Header.Set("X-CSRFToken", csrfToken)
		req.Header.Set("X-IG-App-ID", igAppID)
		req.Header.Set("X-ASBD-ID", igASBDID)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Origin", base)
		req.Header.Set("Referer", base+"/")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "*/*")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("network error: %w", err)
			if err := waitRetry(ctx, time.Duration(3*(attempt+1))*time.Second); err != nil {
				return nil, err
			}
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		if resp.StatusCode == 429 {
			return nil, fmt.Errorf("instagram: rate limited (HTTP 429)")
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			if err := waitRetry(ctx, time.Duration(3*(attempt+1))*time.Second); err != nil {
				return nil, err
			}
			continue
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("instagram: HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
		}

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("instagram: non-JSON response: %s", truncate(string(body), 200))
		}

		if errs, ok := data["errors"].([]interface{}); ok && len(errs) > 0 {
			errObj, _ := errs[0].(map[string]interface{})
			msg, _ := errObj["message"].(string)
			return nil, fmt.Errorf("instagram: graphql error: %s", msg)
		}

		return data, nil
	}
	return nil, lastErr
}

func waitRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func extractIGShortcode(rawURL string) (string, error) {
	s := strings.TrimSpace(rawURL)
	if shortCodeRe.MatchString(s) {
		return s, nil
	}

	if m := igReelsRe.FindStringSubmatch(s); m != nil {
		return m[1], nil
	}
	if m := igPostRe.FindStringSubmatch(s); m != nil {
		return m[1], nil
	}
	if m := igTvRe.FindStringSubmatch(s); m != nil {
		return m[1], nil
	}

	// Try generic path extraction for /p/XXX or /reel/XXX patterns
	parsed, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("cannot parse Instagram URL: %s", rawURL)
	}
	path := strings.Trim(parsed.Path, "/")
	for _, prefix := range []string{"reels", "reel", "p", "tv"} {
		rel := strings.TrimPrefix(path, prefix+"/")
		if rel != path && shortCodeRe.MatchString(rel) {
			return rel, nil
		}
	}

	return "", fmt.Errorf("cannot extract shortcode from %s", rawURL)
}

func parseIGMedia(media map[string]interface{}) Metrics {
	m := Metrics{}
	if v, ok := media["play_count"].(float64); ok {
		m.Views = int64(v)
	}
	if v, ok := media["like_count"].(float64); ok {
		m.Likes = int64(v)
	}
	if v, ok := media["comment_count"].(float64); ok {
		m.Comments = int64(v)
	}
	return m
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
