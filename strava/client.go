package strava

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	colly "github.com/gocolly/colly/v2"
)

type stravaTransport struct{}

func (t *stravaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "strava-tile-proxy")
	return http.DefaultTransport.RoundTrip(req)
}

type Client interface {
	Login() error
	CloudFrontAuth(server string) error
	HttpClient() *http.Client
}

// the strava client holds an http client that maintains auth cookies
type client struct {
	stravaUrl  string
	email      string
	password   string
	httpClient *http.Client

	claimedLogin bool
	claimLock    sync.Mutex
	loginLock    sync.Mutex
}

func NewClient(email, password string) (Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &client{
		stravaUrl: StravaDomain,
		email:     email,
		password:  password,
		httpClient: &http.Client{
			Transport: &stravaTransport{},
			Jar:       jar,
		},
	}, nil
}

func (sc *client) getCSRF() (string, string, error) {
	c := colly.NewCollector()
	c.SetCookieJar(sc.httpClient.Jar)
	var wg sync.WaitGroup
	var csrfParam, csrfToken string
	wg.Add(1)

	c.OnHTML("html", func(e *colly.HTMLElement) {
		csrfToken = e.ChildAttr(`meta[name="csrf-token"]`, "content")
		csrfParam = e.ChildAttr(`meta[name="csrf-param"]`, "content")
		wg.Done()
	})

	err := c.Visit(sc.stravaUrl + "/login")
	if err != nil {
		return "", "", err
	}

	wg.Wait()
	return csrfParam, csrfToken, nil
}

func (sc *client) claimLogin() bool {
	sc.claimLock.Lock()
	claimed := !sc.claimedLogin
	sc.claimedLogin = true
	sc.claimLock.Unlock()
	return claimed
}

func (sc *client) unclaimLogin() {
	sc.claimLock.Lock()
	sc.claimedLogin = false
	sc.claimLock.Unlock()
}

func (sc *client) CloudFrontAuth(server string) error {
	sc.Login()

	sc.loginLock.Lock()
	defer sc.loginLock.Unlock()

	_, err := sc.httpClient.Get(fmt.Sprintf(GlobalHeatmapDomain, server) + "/auth")
	return err
}

func (sc *client) Login() error {
	thisClaimsLogin := sc.claimLogin()
	sc.loginLock.Lock()
	defer sc.loginLock.Unlock()

	if thisClaimsLogin {
		csrfParam, csrfToken, err := sc.getCSRF()
		if err != nil {
			return err
		}

		sessionResponse, err := sc.httpClient.PostForm(sc.stravaUrl+"/session", url.Values{
			csrfParam:  []string{csrfToken},
			"email":    []string{sc.email},
			"password": []string{sc.password},
		})
		if err != nil {
			return err
		}
		sessionResponse.Body.Close()

		sc.unclaimLogin()
	}
	return nil
}

func (sc *client) HttpClient() *http.Client {
	return sc.httpClient
}
