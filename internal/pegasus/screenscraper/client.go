package screenscraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a minimal ScreenScraper API client.
//
// NOTE: ScreenScraper's API requires credentials (at least `devid` + `devpassword`).
// This client is implemented to be mockable in tests and can be evolved to cover
// more endpoints.
type Client struct {
	BaseURL     string
	DevID       string
	DevPassword string
	SoftName    string
	User        string
	Password    string
	HTTPClient  *http.Client
}

func (c Client) http() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (c Client) base() string {
	b := strings.TrimSpace(c.BaseURL)
	if b == "" {
		b = "https://www.screenscraper.fr/api2"
	}
	return strings.TrimRight(b, "/")
}

// ByHashResponse is the (partial) response shape we need.
// It matches the public JSON structure used by ScreenScraper.
// Fields are kept as strings because some APIs return numbers as strings.
type ByHashResponse struct {
	Jeu *struct {
		ID     string `json:"id"`
		Medias struct {
			Media []struct {
				Type string `json:"type"`
				URL  string `json:"url"`
			} `json:"media"`
		} `json:"medias"`
	} `json:"jeu"`
}

// GetGameByRomHash queries ScreenScraper using a ROM hash.
//
// hashType is typically "crc" or "md5" (depending on what you compute).
// This method uses the endpoint name as commonly documented: jeuInfos.php.
func (c Client) GetGameByRomHash(ctx context.Context, hashType, hash string) (*ByHashResponse, error) {
	base := c.base()
	endpoint := base + "/jeuInfos.php"

	q := url.Values{}
	q.Set("output", "json")
	q.Set("devid", c.DevID)
	q.Set("devpassword", c.DevPassword)
	if c.SoftName != "" {
		q.Set("softname", c.SoftName)
	}
	if c.User != "" {
		q.Set("ssid", c.User)
	}
	if c.Password != "" {
		q.Set("sspassword", c.Password)
	}

	// Hash query. ScreenScraper uses romhash/romhashmd5 keys in some docs.
	// We support both common forms by setting only one based on hashType.
	switch strings.ToLower(strings.TrimSpace(hashType)) {
	case "md5":
		q.Set("romhashmd5", hash)
	default:
		q.Set("romhash", hash)
	}

	u := endpoint + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return nil, fmt.Errorf("screenscraper http %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	var out ByHashResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
