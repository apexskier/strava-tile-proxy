package service

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStravaClient struct {
	loginCalls int
	client     *http.Client
}

func (m *mockStravaClient) Login() error {
	m.loginCalls++
	return nil
}

func (m *mockStravaClient) HttpClient() *http.Client {
	return m.client
}

func TestTileService_OK(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tiles/14856714/hot/1/2/3@2x.png", r.URL.Path)
		q := r.URL.Query()
		assert.NotEmpty(t, q["filter_end"])
		assert.Equal(t, []string{"2011-01-01"}, q["filter_start"])
		assert.Equal(t, []string{"all"}, q["filter_type"])
		assert.Equal(t, []string{"true"}, q["include_everyone"])
		assert.Equal(t, []string{"true"}, q["include_followers_only"])
		assert.Equal(t, []string{"true"}, q["include_only_me"])
		assert.Equal(t, []string{"false"}, q["respect_privacy_zones"])
		rw.WriteHeader(http.StatusOK)
		requestCount++
	}))

	stravaClient := mockStravaClient{
		client: mockServer.Client(),
	}
	s := Service{
		stravaClient:     &stravaClient,
		stravaHeatmapUrl: mockServer.URL,
		logger:           log.Default(),

		revealPrivacyZones:           true,
		revealOnlyMeActivities:       true,
		revealFollowerOnlyActivities: true,
		revealPublicActivities:       true,
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3", nil)
	w := httptest.NewRecorder()

	err := s.ServeTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, 1, requestCount)
	assert.Equal(t, 0, stravaClient.loginCalls)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTileService_OK_custom_params(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tiles/14856714/purple/1/2/3@2x.png", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, []string{"winter"}, q["filter_type"])
		rw.WriteHeader(http.StatusOK)
		requestCount++
	}))

	stravaClient := mockStravaClient{
		client: mockServer.Client(),
	}
	s := Service{
		stravaClient:     &stravaClient,
		stravaHeatmapUrl: mockServer.URL,
		logger:           log.Default(),

		revealPrivacyZones:           true,
		revealOnlyMeActivities:       true,
		revealFollowerOnlyActivities: true,
		revealPublicActivities:       true,
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3?color=purple&sport=winter", nil)
	w := httptest.NewRecorder()

	err := s.ServeTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, 1, requestCount)
	assert.Equal(t, 0, stravaClient.loginCalls)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTileService_401(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tiles/14856714/hot/1/2/3@2x.png", r.URL.Path)
		if requestCount == 0 {
			rw.WriteHeader(http.StatusUnauthorized)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		requestCount++
	}))

	stravaClient := mockStravaClient{
		client: mockServer.Client(),
	}
	s := Service{
		stravaClient:     &stravaClient,
		stravaHeatmapUrl: mockServer.URL,
		logger:           log.Default(),

		revealPrivacyZones:           true,
		revealOnlyMeActivities:       true,
		revealFollowerOnlyActivities: true,
		revealPublicActivities:       true,
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3", nil)
	w := httptest.NewRecorder()

	err := s.ServeTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, 2, requestCount)
	assert.Equal(t, 1, stravaClient.loginCalls)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTileService_404(t *testing.T) {
	stravaClient := mockStravaClient{}
	s := Service{
		stravaClient: &stravaClient,
		logger:       log.Default(),

		revealPrivacyZones:           true,
		revealOnlyMeActivities:       true,
		revealFollowerOnlyActivities: true,
		revealPublicActivities:       true,
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/garbage", nil)
	err := s.ServeTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTileService_400(t *testing.T) {
	stravaClient := mockStravaClient{}
	s := Service{
		stravaClient: &stravaClient,
		logger:       log.Default(),

		revealPrivacyZones:           true,
		revealOnlyMeActivities:       true,
		revealFollowerOnlyActivities: true,
		revealPublicActivities:       true,
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3?color=garbage", nil)
	err := s.ServeTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "https://example.com/tiles/1/2/3?sport=garbage", nil)
	err = s.ServeTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
