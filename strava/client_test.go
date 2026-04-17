package strava

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudFrontExpiresAt(t *testing.T) {
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	sc := &client{
		stravaUrl:  "https://www.strava.com",
		httpClient: &http.Client{Jar: jar},
	}

	// No cookie yet — should return zero time.
	assert.True(t, sc.CloudFrontExpiresAt().IsZero())

	// Set the cookie directly on the jar.
	expireMs := int64(1776129917000)
	u, _ := url.Parse("https://www.strava.com")
	jar.SetCookies(u, []*http.Cookie{
		{Name: "_strava_CloudFront-Expires", Value: "1776129917000", Path: "/"},
	})

	got := sc.CloudFrontExpiresAt()
	assert.Equal(t, time.UnixMilli(expireMs), got)
}

func TestRefreshCloudFrontCookies_concurrent(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/maps", r.URL.Path)
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	sc := &client{
		stravaUrl:  server.URL,
		httpClient: &http.Client{Jar: jar},
	}

	// Concurrent calls should result in a single actual HTTP request.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			assert.NoError(t, sc.RefreshCloudFrontCookies())
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, 1, requestCount)

	// A second batch should also work.
	requestCount = 0
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			assert.NoError(t, sc.RefreshCloudFrontCookies())
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, 1, requestCount)
}

func TestAthleteID(t *testing.T) {
	// No token — should return error.
	sc, err := NewClient("", "")
	require.NoError(t, err)
	_, err = sc.AthleteID()
	assert.Error(t, err)

	// Build a minimal JWT: header.payload.sig (signature not verified).
	// Payload: {"sub":98765}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOjk4NzY1fQ.fakesig"
	sc, err = NewClient(token, "dummy")
	require.NoError(t, err)

	id, err := sc.AthleteID()
	require.NoError(t, err)
	assert.Equal(t, "98765", id)
}
