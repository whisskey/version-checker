package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	testVersions = []string{"1.0.0", "1.0.1", "1.0.2"}
)

func createVersionResponse(versions []string) []byte {
	response := struct {
		Versions []struct {
			Num string `json:"num"`
		} `json:"versions"`
	}{
		Versions: make([]struct {
			Num string `json:"num"`
		}, len(versions)),
	}

	for i := range versions {
		response.Versions[i].Num = versions[len(versions)-1-i]
	}

	data, _ := json.Marshal(response)
	return data
}

type CrateServiceSuite struct {
	suite.Suite
	svc    CrateService
	server *httptest.Server
	mu     sync.Mutex
}

func (s *CrateServiceSuite) SetupTest() {
	s.updateServer(s.createSuccessHandler(testVersions))
}

func (s *CrateServiceSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *CrateServiceSuite) updateServer(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		s.server.Close()
	}
	s.server = httptest.NewServer(handler)
	s.svc = New(WithAPIURL(s.server.URL))
}

func (s *CrateServiceSuite) createSuccessHandler(versions []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(createVersionResponse(versions))
	})
}

func (s *CrateServiceSuite) createErrorHandler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
}

func (s *CrateServiceSuite) createTimeoutHandler(delay time.Duration, versions []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Write(createVersionResponse(versions))
	})
}

func (s *CrateServiceSuite) withErrorResponse(status int, testFunc func()) {
	s.updateServer(s.createErrorHandler(status))
	testFunc()
}

func (s *CrateServiceSuite) TestNew() {
	s.Run("Should create service with default options", func() {
		svc := New()
		s.NotNil(svc)
	})

	s.Run("Should create service with custom URL", func() {
		svc := New(WithAPIURL("http://zort-url"))
		s.NotNil(svc)
	})

	s.Run("Should create service with custom HTTP client", func() {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		svc := New(WithHTTPClient(client))
		s.NotNil(svc)
	})
}

func (s *CrateServiceSuite) TestCheckVersion() {
	s.Run("Should check for updates", func() {
		result, err := s.svc.CheckVersion("serde", "1.0.0")

		s.NoError(err)
		s.Equal("1.0.2", result.Latest)
		s.True(result.HasUpdate)
	})

	s.Run("Should handle same version", func() {
		result, err := s.svc.CheckVersion("serde", "1.0.2")

		s.NoError(err)
		s.Equal("1.0.2", result.Latest)
		s.False(result.HasUpdate)
	})

	s.Run("Should handle newer local version", func() {
		result, err := s.svc.CheckVersion("serde", "1.0.3")

		s.NoError(err)
		s.Equal("1.0.2", result.Latest)
		s.False(result.HasUpdate)
	})

	s.Run("Should handle response with meta information", func() {
		s.updateServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := struct {
				Versions []struct {
					Num string `json:"num"`
				} `json:"versions"`
				Meta struct {
					Total    int `json:"total"`
					NextPage any `json:"next_page"`
				} `json:"meta"`
			}{
				Versions: []struct {
					Num string `json:"num"`
				}{{Num: "1.0.2"}},
				Meta: struct {
					Total    int `json:"total"`
					NextPage any `json:"next_page"`
				}{
					Total:    1,
					NextPage: nil,
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		result, err := s.svc.CheckVersion("serde", "1.0.0")

		s.NoError(err)
		s.Equal("1.0.2", result.Latest)
		s.True(result.HasUpdate)
	})

	s.Run("Should fail with request creation error", func() {
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return nil, fmt.Errorf("failed to create connection")
				},
			},
		}
		s.svc = New(WithAPIURL("http://invalid-url"), WithHTTPClient(client))

		_, err := s.svc.CheckVersion("serde", "1.0.0")
		s.Error(err)
		s.Contains(err.Error(), "failed to fetch package info")
	})

	s.Run("Should fail with invalid URL", func() {
		s.svc = New(WithAPIURL("http://[::1]:namedport"))
		_, err := s.svc.CheckVersion("serde", "1.0.0")

		s.Error(err)
		s.Contains(err.Error(), "failed to create request")
	})

	s.Run("Should fail with package not found", func() {
		s.withErrorResponse(http.StatusNotFound, func() {
			_, err := s.svc.CheckVersion("non-existent", "1.0.0")

			s.Error(err)
			s.Contains(err.Error(), ErrPackageNotFound.Error())
		})
	})

	s.Run("Should fail with invalid response body", func() {
		s.updateServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid json"))
		}))
		_, err := s.svc.CheckVersion("serde", "1.0.0")

		s.Error(err)
		s.Contains(err.Error(), "failed to parse response")
	})

	s.Run("Should fail with invalid latest version", func() {
		s.updateServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := struct {
				Versions []struct {
					Num string `json:"num"`
				} `json:"versions"`
			}{
				Versions: []struct {
					Num string `json:"num"`
				}{{Num: "invalid"}},
			}
			json.NewEncoder(w).Encode(response)
		}))
		_, err := s.svc.CheckVersion("serde", "1.0.0")

		s.Error(err)
		s.Contains(err.Error(), "invalid latest version format")
	})

	s.Run("Should fail with empty package name", func() {
		s.withErrorResponse(http.StatusBadRequest, func() {
			_, err := s.svc.CheckVersion("", "1.0.0")

			s.Error(err)
			s.Contains(err.Error(), ErrEmptyInput.Error())
		})
	})

	s.Run("Should fail with empty version", func() {
		s.withErrorResponse(http.StatusBadRequest, func() {
			_, err := s.svc.CheckVersion("serde", "")

			s.Error(err)
			s.Contains(err.Error(), ErrEmptyInput.Error())
		})
	})

	s.Run("Should fail with invalid version format", func() {
		s.withErrorResponse(http.StatusBadRequest, func() {
			_, err := s.svc.CheckVersion("serde", "invalid")

			s.Error(err)
			s.Contains(err.Error(), "invalid version format")
		})
	})

	s.Run("Should fail with API errors", func() {
		s.withErrorResponse(http.StatusInternalServerError, func() {
			_, err := s.svc.CheckVersion("serde", "1.0.0")

			s.Error(err)
			s.Contains(err.Error(), "API error: status 500")
		})
	})

	s.Run("Should fail with rate limit error", func() {
		s.withErrorResponse(http.StatusTooManyRequests, func() {
			_, err := s.svc.CheckVersion("serde", "1.0.0")

			s.Error(err)
			s.Contains(err.Error(), "API error: status 429")
		})
	})

	s.Run("Should fail with forbidden error", func() {
		s.withErrorResponse(http.StatusForbidden, func() {
			_, err := s.svc.CheckVersion("serde", "1.0.0")

			s.Error(err)
			s.Contains(err.Error(), "API error: status 403")
		})
	})

	s.Run("Should fail with no versions available", func() {
		s.updateServer(s.createSuccessHandler([]string{}))
		_, err := s.svc.CheckVersion("serde", "1.0.0")

		s.Error(err)
		s.Contains(err.Error(), ErrNoVersions.Error())
	})

	s.Run("Should fail with timeout", func() {
		s.updateServer(s.createTimeoutHandler(200*time.Millisecond, testVersions))
		client := &http.Client{
			Timeout: 100 * time.Millisecond,
		}
		s.svc = New(WithAPIURL(s.server.URL), WithHTTPClient(client))

		_, err := s.svc.CheckVersion("serde", "1.0.0")

		s.Error(err)
		s.Contains(err.Error(), "failed to fetch package info")
	})
}

func TestCrateServiceSuite(t *testing.T) {
	suite.Run(t, new(CrateServiceSuite))
}
