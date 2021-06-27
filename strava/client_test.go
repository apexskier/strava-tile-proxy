package strava

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStravaClient(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requestCount {
		case 0:
			assert.Equal(t, "/login", r.URL.Path)
			w.Write([]byte(`<html><head>
				<meta name="csrf-param" content="csrf_param_key" />
				<meta name="csrf-token" content="csrf_param_value" />
			</head></html>`))
		case 1:
			assert.Equal(t, "/session", r.URL.Path)
			body, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			query, err := url.ParseQuery(string(body))
			require.NoError(t, err)
			assert.Equal(t, url.Values{
				"csrf_param_key": []string{"csrf_param_value"},
				"email":          []string{"test@example.com"},
				"password":       []string{"password"},
			}, query)

			w.Header().Add("Set-Cookie", "auth=auth_token")
			w.WriteHeader(200)
		default:
			assert.Fail(t, "unexpected request")
		}
		requestCount++
	}))

	sc, err := NewClient("test@example.com", "password")
	require.NoError(t, err)
	sc.stravaUrl = server.URL

	require.NoError(t, sc.Login())
	// make sure the server handled requests
	assert.Equal(t, 2, requestCount)

	// auth token cookie was set in the client's jar
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	cookies := sc.HttpClient.Jar.Cookies(u)
	assert.Len(t, cookies, 1)
	assert.Equal(t, http.Cookie{Name: "auth", Value: "auth_token"}, *cookies[0])
}
