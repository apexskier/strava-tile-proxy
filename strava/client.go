package strava

import (
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

// the strava client holds an http client that maintains auth cookies
type stravaClient struct {
	stravaUrl  string
	email      string
	password   string
	HttpClient *http.Client
}

func NewClient(email, password string) (*stravaClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &stravaClient{
		stravaUrl: "https://www.strava.com",
		email:     email,
		password:  password,
		HttpClient: &http.Client{
			Transport: &stravaTransport{},
			Jar:       jar,
		},
	}, nil
}

func (sc *stravaClient) getCSRF() (string, string, error) {
	c := colly.NewCollector()
	c.SetCookieJar(sc.HttpClient.Jar)
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

func (sc *stravaClient) Login() error {
	csrfParam, csrfToken, err := sc.getCSRF()
	if err != nil {
		return err
	}

	sessionResponse, err := sc.HttpClient.PostForm(sc.stravaUrl+"/session", url.Values{
		csrfParam:  []string{csrfToken},
		"email":    []string{sc.email},
		"password": []string{sc.password},
	})
	if err != nil {
		return err
	}
	sessionResponse.Body.Close()
	return nil
}
