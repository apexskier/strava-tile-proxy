package strava

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type stravaTransport struct{}

func (t *stravaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "strava-tile-proxy")
	return http.DefaultTransport.RoundTrip(req)
}

type Client interface {
	RefreshCloudFrontCookies() error
	CloudFrontExpiresAt() time.Time
	AthleteID() (string, error)
	HttpClient() *http.Client
}

// client holds an http.Client that maintains Strava auth cookies.
type client struct {
	stravaUrl     string
	httpClient    *http.Client
	rememberToken string
	stravaSession string

	claimedRefresh bool
	claimLock      sync.Mutex
	refreshLock    sync.Mutex
}

func NewClient(rememberToken, stravaSession string) (Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	sc := &client{
		stravaUrl:     StravaDomain,
		rememberToken: rememberToken,
		stravaSession: stravaSession,
		httpClient: &http.Client{
			Transport: &stravaTransport{},
			Jar:       jar,
		},
	}
	if err := sc.setSessionCookies(); err != nil {
		return nil, err
	}
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			sc.setSessionCookies() //nolint:errcheck
		}
	}()
	return sc, nil
}

func (sc *client) setSessionCookies() error {
	u, err := url.Parse(sc.stravaUrl)
	if err != nil {
		return err
	}
	expires := time.Now().Add(365 * 24 * time.Hour)
	sc.httpClient.Jar.SetCookies(u, []*http.Cookie{
		{
			Name:     "strava_remember_token",
			Value:    sc.rememberToken,
			Domain:   "www.strava.com",
			Path:     "/",
			Expires:  expires,
			HttpOnly: true,
		},
		{
			Name:     "_strava4_session",
			Value:    sc.stravaSession,
			Domain:   ".strava.com",
			Path:     "/",
			Expires:  expires,
			Secure:   true,
			HttpOnly: true,
		},
	})
	return nil
}

func (sc *client) claimRefresh() bool {
	sc.claimLock.Lock()
	claimed := !sc.claimedRefresh
	sc.claimedRefresh = true
	sc.claimLock.Unlock()
	return claimed
}

func (sc *client) unclaimRefresh() {
	sc.claimLock.Lock()
	sc.claimedRefresh = false
	sc.claimLock.Unlock()
}

// RefreshCloudFrontCookies makes an authenticated request to Strava so it
// issues fresh CloudFront signed cookies. Concurrent callers are coalesced
// into a single actual HTTP request.
func (sc *client) RefreshCloudFrontCookies() error {
	thisClaimsRefresh := sc.claimRefresh()
	sc.refreshLock.Lock()
	defer sc.refreshLock.Unlock()

	if thisClaimsRefresh {
		resp, err := sc.httpClient.Get(sc.stravaUrl + "/maps")
		if err != nil {
			return err
		}
		resp.Body.Close()
		sc.unclaimRefresh()
	}
	return nil
}

// CloudFrontExpiresAt returns the expiry time of the CloudFront signed cookies
// by reading the _strava_CloudFront-Expires cookie from the jar.
// Returns a zero time.Time if the cookie is absent or unparseable.
func (sc *client) CloudFrontExpiresAt() time.Time {
	u, err := url.Parse(sc.stravaUrl)
	if err != nil {
		return time.Time{}
	}
	for _, cookie := range sc.httpClient.Jar.Cookies(u) {
		if cookie.Name == "_strava_CloudFront-Expires" {
			ms, err := strconv.ParseInt(cookie.Value, 10, 64)
			if err == nil {
				return time.UnixMilli(ms)
			}
		}
	}
	return time.Time{}
}

func (sc *client) HttpClient() *http.Client {
	return sc.httpClient
}

// AthleteID returns the athlete's numeric ID by parsing the strava_remember_token
// cookie as a JWT and extracting the subject claim.
func (sc *client) AthleteID() (string, error) {
	u, err := url.Parse(sc.stravaUrl)
	if err != nil {
		return "", err
	}
	for _, cookie := range sc.httpClient.Jar.Cookies(u) {
		if cookie.Name == "strava_remember_token" {
			parts := strings.SplitN(cookie.Value, ".", 3)
			if len(parts) != 3 {
				return "", errors.New("strava_remember_token is not a valid JWT")
			}
			payload, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				return "", errors.Wrap(err, "decoding JWT payload")
			}
			var claims struct {
				Sub int `json:"sub"`
			}
			if err := json.Unmarshal(payload, &claims); err != nil {
				return "", errors.Wrap(err, "parsing JWT claims")
			}
			return strconv.Itoa(claims.Sub), nil
		}
	}
	return "", errors.New("strava_remember_token cookie not found")
}
