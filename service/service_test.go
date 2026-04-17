package service

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apexskier/strava-tile-proxy/strava"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockStravaClient struct {
	strava.Client
	mock.Mock
}

func (m *mockStravaClient) HttpClient() *http.Client {
	args := m.Called()
	return args.Get(0).(*http.Client)
}

func (m *mockStravaClient) RefreshCloudFrontCookies() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockStravaClient) AthleteID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestTileService_BadApiToken(t *testing.T) {
	stravaClient := mockStravaClient{}
	defer stravaClient.AssertExpectations(t)

	s := Service{
		stravaClient: &stravaClient,
		logger:       log.Default(),
		apiToken:     "token",
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3?api_token=garbage", nil)
	w := httptest.NewRecorder()

	err := s.ServeGlobalTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTileService_GlobalOK(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server-a/identified/globalheat/all/blue/1/2/3@2x.png", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, []string{"19"}, q["v"])
		rw.WriteHeader(http.StatusOK)
		requestCount++
	}))

	stravaClient := mockStravaClient{}
	defer stravaClient.AssertExpectations(t)

	stravaClient.On("HttpClient").Return(mockServer.Client())

	s := Service{
		stravaClient: &stravaClient,
		logger:       log.Default(),
		apiToken:     "token",

		globalHeatmapDomain: mockServer.URL + "/server-a",
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3?api_token=token", nil)
	w := httptest.NewRecorder()

	err := s.ServeGlobalTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, 1, requestCount)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTileService_GlobalOK_custom_params(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server-a/identified/globalheat/winter/purple/1/2/3@2x.png", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, []string{"19"}, q["v"])
		rw.WriteHeader(http.StatusOK)
		requestCount++
	}))

	stravaClient := mockStravaClient{}
	defer stravaClient.AssertExpectations(t)

	stravaClient.On("HttpClient").Return(mockServer.Client())

	s := Service{
		stravaClient: &stravaClient,
		logger:       log.Default(),

		globalHeatmapDomain: mockServer.URL + "/server-a",
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3?color=purple&sports=winter", nil)
	w := httptest.NewRecorder()

	err := s.ServeGlobalTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, 1, requestCount)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTileService_PersonalOK(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tiles/12321/orange/1/2/3@2x.png", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, []string{"true"}, q["include_everyone"])
		assert.Equal(t, []string{"true"}, q["include_followers_only"])
		assert.Equal(t, []string{"true"}, q["include_only_me"])
		assert.Equal(t, []string{"false"}, q["respect_privacy_zones"])
		rw.WriteHeader(http.StatusOK)
		requestCount++
	}))

	stravaClient := mockStravaClient{}
	defer stravaClient.AssertExpectations(t)

	stravaClient.On("HttpClient").Return(mockServer.Client())
	stravaClient.On("AthleteID").Return("12321", nil)

	s := Service{
		stravaClient: &stravaClient,
		logger:       log.Default(),

		personalHeatmapDomain:        mockServer.URL,
		revealPrivacyZones:           true,
		revealOnlyMeActivities:       true,
		revealFollowerOnlyActivities: true,
		revealPublicActivities:       true,
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3", nil)
	w := httptest.NewRecorder()

	err := s.ServePersonalTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, 1, requestCount)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTileService_PersonalOK_custom_params(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tiles/12321/purple/1/2/3@2x.png", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, []string{"winter"}, q["filter_type"])
		rw.WriteHeader(http.StatusOK)
		requestCount++
	}))

	stravaClient := mockStravaClient{}
	defer stravaClient.AssertExpectations(t)

	stravaClient.On("HttpClient").Return(mockServer.Client())
	stravaClient.On("AthleteID").Return("12321", nil)

	s := Service{
		stravaClient: &stravaClient,
		logger:       log.Default(),

		personalHeatmapDomain:        mockServer.URL,
		revealPrivacyZones:           true,
		revealOnlyMeActivities:       true,
		revealFollowerOnlyActivities: true,
		revealPublicActivities:       true,
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3?color=purple&sports=winter", nil)
	w := httptest.NewRecorder()

	err := s.ServePersonalTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, 1, requestCount)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTileService_Personal401(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tiles/12321/orange/1/2/3@2x.png", r.URL.Path)
		if requestCount == 0 {
			rw.WriteHeader(http.StatusUnauthorized)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		requestCount++
	}))

	stravaClient := mockStravaClient{}
	defer stravaClient.AssertExpectations(t)

	stravaClient.On("HttpClient").Return(mockServer.Client())
	stravaClient.On("RefreshCloudFrontCookies").Return(nil)
	stravaClient.On("AthleteID").Return("12321", nil)

	s := Service{
		stravaClient: &stravaClient,
		logger:       log.Default(),

		personalHeatmapDomain:        mockServer.URL,
		revealPrivacyZones:           true,
		revealOnlyMeActivities:       true,
		revealFollowerOnlyActivities: true,
		revealPublicActivities:       true,
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3", nil)
	w := httptest.NewRecorder()

	err := s.ServePersonalTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, 2, requestCount)
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
	err := s.ServePersonalTile(w, req)

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
	err := s.ServePersonalTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTileService_Global403(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server-a/identified/globalheat/all/blue/1/2/3@2x.png", r.URL.Path)
		if requestCount == 0 {
			rw.WriteHeader(http.StatusUnauthorized)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		requestCount++
	}))

	stravaClient := mockStravaClient{}
	defer stravaClient.AssertExpectations(t)

	stravaClient.On("HttpClient").Return(mockServer.Client())
	stravaClient.On("RefreshCloudFrontCookies").Return(nil)

	s := Service{
		stravaClient: &stravaClient,
		logger:       log.Default(),

		globalHeatmapDomain: mockServer.URL + "/server-a",
	}

	req := httptest.NewRequest("GET", "https://example.com/tiles/1/2/3", nil)
	w := httptest.NewRecorder()

	err := s.ServeGlobalTile(w, req)

	require.NoError(t, err)
	assert.Equal(t, 2, requestCount)
	assert.Equal(t, http.StatusOK, w.Code)
}
