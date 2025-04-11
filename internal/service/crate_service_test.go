package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of http.Client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []ServiceOption
		wantURL string
	}{
		{
			name:    "default configuration",
			opts:    []ServiceOption{},
			wantURL: "https://crates.io/api/v1/crates",
		},
		{
			name: "custom API URL",
			opts: []ServiceOption{
				WithAPIURL("https://custom-api.example.com"),
			},
			wantURL: "https://custom-api.example.com",
		},
		{
			name: "custom HTTP client",
			opts: []ServiceOption{
				WithHTTPClient(&http.Client{
					Timeout: 5 * time.Second,
				}),
			},
			wantURL: "https://crates.io/api/v1/crates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.opts...)
			assert.NotNil(t, svc)
			assert.Implements(t, (*CrateService)(nil), svc)
		})
	}
}

func TestService_CheckVersion(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		version     string
		setupMock   func() *httptest.Server
		wantResult  *Result
		wantErr     error
	}{
		{
			name:        "valid package with update available",
			packageName: "serde",
			version:     "1.0.0",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"versions": []map[string]string{
							{"num": "1.0.1"},
							{"num": "1.0.0"},
						},
					})
				}))
			},
			wantResult: &Result{
				Latest:    "1.0.1",
				HasUpdate: true,
			},
			wantErr: nil,
		},
		{
			name:        "valid package with no update available",
			packageName: "serde",
			version:     "1.0.1",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"versions": []map[string]string{
							{"num": "1.0.1"},
							{"num": "1.0.0"},
						},
					})
				}))
			},
			wantResult: &Result{
				Latest:    "1.0.1",
				HasUpdate: false,
			},
			wantErr: nil,
		},
		{
			name:        "package not found",
			packageName: "non-existent-package",
			version:     "1.0.0",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			wantResult: nil,
			wantErr:    ErrPackageNotFound,
		},
		{
			name:        "empty package name",
			packageName: "",
			version:     "1.0.0",
			setupMock:   nil,
			wantResult:  nil,
			wantErr:     ErrEmptyInput,
		},
		{
			name:        "empty version",
			packageName: "serde",
			version:     "",
			setupMock:   nil,
			wantResult:  nil,
			wantErr:     ErrEmptyInput,
		},
		{
			name:        "invalid version format",
			packageName: "serde",
			version:     "invalid-version",
			setupMock:   nil,
			wantResult:  nil,
			wantErr:     errors.New("invalid version format"),
		},
		{
			name:        "API error",
			packageName: "serde",
			version:     "1.0.0",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			wantResult: nil,
			wantErr:    errors.New("API error: status 500"),
		},
		{
			name:        "no versions available",
			packageName: "serde",
			version:     "1.0.0",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"versions": []map[string]string{},
					})
				}))
			},
			wantResult: nil,
			wantErr:    ErrNoVersions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupMock != nil {
				server = tt.setupMock()
				defer server.Close()
			}

			var svc CrateService
			if server != nil {
				svc = New(WithAPIURL(server.URL))
			} else {
				svc = New()
			}

			result, err := svc.CheckVersion(tt.packageName, tt.version)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if tt.wantErr == ErrEmptyInput || tt.wantErr == ErrPackageNotFound || tt.wantErr == ErrNoVersions {
					assert.Equal(t, tt.wantErr, err)
				} else {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult.Latest, result.Latest)
			assert.Equal(t, tt.wantResult.HasUpdate, result.HasUpdate)
		})
	}
}
