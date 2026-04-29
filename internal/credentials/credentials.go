package credentials

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// File is the shape stored at ~/.insighta/credentials.json
type File struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	Username         string    `json:"username"`
	Role             string    `json:"role"`
	AccessTokenExpAt time.Time `json:"access_token_exp_at"`
}

var ErrNotLoggedIn = errors.New("not logged in — run: insighta login")

func path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".insighta", "credentials.json"), nil
}

func Load() (*File, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotLoggedIn
		}
		return nil, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, ErrNotLoggedIn
	}
	if f.AccessToken == "" {
		return nil, ErrNotLoggedIn
	}
	return &f, nil
}

func Save(f *File) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	// 0600 — owner read/write only
	return os.WriteFile(p, data, 0600)
}

func Clear() error {
	p, err := path()
	if err != nil {
		return err
	}
	err = os.Remove(p)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
