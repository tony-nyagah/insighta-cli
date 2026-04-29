package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"insighta-cli/internal/credentials"
)

// ErrUnauthorized is returned when the server responds with 401 and
// the refresh attempt also fails — caller should prompt re-login.
var ErrUnauthorized = errors.New("session expired — run: insighta login")

// Client wraps http.Client with automatic token refresh and credential
// injection.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New() *Client {
	base := os.Getenv("INSIGHTA_API_URL")
	if base == "" {
		base = "https://api.insighta.app" // overridden per-user via env
	}
	return &Client{
		baseURL:    base,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Do performs an authenticated request, transparently refreshing the access
// token if it is expired. Returns the decoded JSON body (map or slice).
func (c *Client) Do(method, path string, body interface{}) ([]byte, int, error) {
	creds, err := credentials.Load()
	if err != nil {
		return nil, 0, err
	}

	// Proactively refresh if the access token is within 10 s of expiry
	if time.Until(creds.AccessTokenExpAt) < 10*time.Second {
		if refreshErr := c.refresh(creds); refreshErr != nil {
			return nil, 0, ErrUnauthorized
		}
	}

	return c.doWithCreds(method, path, body, creds.AccessToken)
}

// doWithCreds sends the request; if a 401 comes back it tries one refresh
// then retries.
func (c *Client) doWithCreds(method, path string, body interface{}, token string) ([]byte, int, error) {
	resp, raw, err := c.send(method, path, body, token)
	if err != nil {
		return nil, 0, err
	}
	if resp == http.StatusUnauthorized {
		// One retry after refresh
		creds, lerr := credentials.Load()
		if lerr != nil {
			return nil, resp, ErrUnauthorized
		}
		if rerr := c.refresh(creds); rerr != nil {
			return nil, resp, ErrUnauthorized
		}
		creds, _ = credentials.Load()
		resp, raw, err = c.send(method, path, body, creds.AccessToken)
		if err != nil {
			return nil, resp, err
		}
		if resp == http.StatusUnauthorized {
			return nil, resp, ErrUnauthorized
		}
	}
	return raw, resp, nil
}

func (c *Client) send(method, path string, body interface{}, token string) (int, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-API-Version", "1")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return res.StatusCode, raw, nil
}

// refresh calls /auth/refresh, updates credentials on disk.
func (c *Client) refresh(creds *credentials.File) error {
	payload := map[string]string{"refresh_token": creds.RefreshToken}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/auth/refresh", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ErrUnauthorized
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}

	creds.AccessToken = result.AccessToken
	creds.RefreshToken = result.RefreshToken
	creds.AccessTokenExpAt = time.Now().Add(3 * time.Minute)
	return credentials.Save(creds)
}

// GetRaw performs an authenticated GET and returns raw bytes (used for CSV download).
func (c *Client) GetRaw(path string) ([]byte, string, error) {
	creds, err := credentials.Load()
	if err != nil {
		return nil, "", err
	}
	if time.Until(creds.AccessTokenExpAt) < 10*time.Second {
		if err := c.refresh(creds); err != nil {
			return nil, "", ErrUnauthorized
		}
		creds, _ = credentials.Load()
	}

	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+creds.AccessToken)
	req.Header.Set("X-API-Version", "1")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(res.Body)
		return nil, "", fmt.Errorf("server returned %d: %s", res.StatusCode, string(raw))
	}

	raw, err := io.ReadAll(res.Body)
	contentDisposition := res.Header.Get("Content-Disposition")
	return raw, contentDisposition, err
}

// PostNoAuth sends a POST without auth headers (used for OAuth callback exchange).
func PostNoAuth(url string, payload interface{}) ([]byte, int, error) {
	b, _ := json.Marshal(payload)
	res, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return raw, res.StatusCode, nil
}
