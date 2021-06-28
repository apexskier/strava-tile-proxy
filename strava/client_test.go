package strava

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStravaClient(t *testing.T) {
	requestCount := 0
	authTokenCount := 0
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

			w.Header().Add("Set-Cookie", fmt.Sprintf("auth=auth_token_%d", authTokenCount))
			authTokenCount++
			w.WriteHeader(200)
		default:
			assert.Fail(t, "unexpected request")
		}
		requestCount++
	}))

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	sc := &client{
		stravaUrl: server.URL,
		email:     "test@example.com",
		password:  "password",
		httpClient: &http.Client{
			Jar: jar,
		},
	}

	// concurrent logins should result in a single actual login
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			assert.NoError(t, sc.Login())
			wg.Done()
		}()
	}
	wg.Wait()

	// make sure the server handled requests
	assert.Equal(t, 2, requestCount)

	// auth token cookie was set in the client's jar
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	cookies := jar.Cookies(u)
	require.Len(t, cookies, 1)
	assert.Equal(t, http.Cookie{Name: "auth", Value: "auth_token_0"}, *cookies[0])

	requestCount = 0

	// the next batch of concurrent logins should also work
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			assert.NoError(t, sc.Login())
			wg.Done()
		}()
	}
	wg.Wait()

	// make sure the server handled requests
	assert.Equal(t, 2, requestCount)

	// auth token cookie was set in the client's jar
	cookies = jar.Cookies(u)
	require.Len(t, cookies, 1)
	assert.Equal(t, http.Cookie{Name: "auth", Value: "auth_token_1"}, *cookies[0])
}
