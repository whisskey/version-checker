package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"
)

var (
	ErrEmptyInput      = errors.New("package name and version are required")
	ErrPackageNotFound = errors.New("package not found")
	ErrNoVersions      = errors.New("package has no version information")
)

type CrateService interface {
	CheckVersion(name, version string) (*Result, error)
}

type Service struct {
	client *http.Client
	apiURL string
}

type Result struct {
	Latest    string
	HasUpdate bool
}

type versionsResponse struct {
	Versions []struct {
		Num string `json:"num"`
	} `json:"versions"`
	Meta struct {
		Total    int `json:"total"`
		NextPage any `json:"next_page"`
	} `json:"meta"`
}

type ServiceOption func(*Service)

func WithHTTPClient(client *http.Client) ServiceOption {
	return func(s *Service) {
		s.client = client
	}
}

func WithAPIURL(url string) ServiceOption {
	return func(s *Service) {
		s.apiURL = url
	}
}

func New(opts ...ServiceOption) CrateService {
	s := &Service{
		client: &http.Client{},
		apiURL: "https://crates.io/api/v1/crates",
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Service) CheckVersion(name, version string) (*Result, error) {
	if name == "" || version == "" {
		return nil, ErrEmptyInput
	}

	currentVer, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/%s/versions", s.apiURL, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPackageNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result versionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Versions) == 0 {
		return nil, ErrNoVersions
	}

	latestVerStr := result.Versions[0].Num
	latestVer, err := semver.NewVersion(latestVerStr)
	if err != nil {
		return nil, fmt.Errorf("invalid latest version format: %w", err)
	}

	return &Result{
		Latest:    latestVerStr,
		HasUpdate: latestVer.GreaterThan(currentVer),
	}, nil
}
