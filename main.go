package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/apexskier/strava-tile-proxy/strava"
)

var logger = log.Default()

func main() {
	stravaClient, err := strava.NewClient(
		os.Getenv("STRAVA_EMAIL"),
		os.Getenv("STRAVA_PASSWORD"),
	)
	if err != nil {
		panic(err)
	}

	tileRouteRe := regexp.MustCompile(`^/tiles/(?P<z>\d+)/(?P<x>\d+)/(?P<y>\d+)$`)

	serveTile := func(rw http.ResponseWriter, r *http.Request) error {
		tileRouteMatches := tileRouteRe.FindStringSubmatch(r.URL.Path)
		if len(tileRouteMatches) == 0 {
			rw.WriteHeader(http.StatusNotFound)
			return nil
		}
		z, err := strconv.ParseUint(tileRouteMatches[1], 10, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("invalid tile z"))
			return nil
		}
		x, err := strconv.ParseUint(tileRouteMatches[2], 10, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("invalid tile x"))
			return nil
		}
		y, err := strconv.ParseUint(tileRouteMatches[3], 10, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("invalid tile y"))
			return nil
		}

		q := r.URL.Query()
		heatColor := strava.HeatRed
		if heats, ok := q["color"]; ok && len(heats) > 0 {
			color, err := strava.ParseHeat(heats[0])
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(err.Error()))
				return nil
			}
			heatColor = color
		}

		sport := strava.SportAll
		if heats, ok := q["sport"]; ok && len(heats) > 0 {
			s, err := strava.ParseSport(heats[0])
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(err.Error()))
				return nil
			}
			sport = s
		}

		tileQueryParams := url.Values{
			"filter_type":            []string{string(sport)},
			"filter_start":           []string{"2011-01-01"},
			"filter_end":             []string{time.Now().Format("2006-01-02")},
			"respect_privacy_zones":  []string{strconv.FormatBool(true)},
			"include_everyone":       []string{strconv.FormatBool(true)},
			"include_followers_only": []string{strconv.FormatBool(true)},
			"include_only_me":        []string{strconv.FormatBool(true)},
		}
		url := fmt.Sprintf(
			"https://personal-heatmaps-external.strava.com/tiles/%s/%s/%d/%d/%d@2x.png?%s",
			strava.TilesNamespace,
			heatColor,
			z,
			x,
			y,
			tileQueryParams.Encode(),
		)
		tileResponse, err := stravaClient.HttpClient.Get(url)
		if err != nil {
			return err
		}
		if tileResponse.StatusCode == http.StatusUnauthorized {
			// authenticate and retry once
			logger.Println("auto-authenticating")
			if err := stravaClient.Login(); err != nil {
				return err
			}
			tileResponse, err = stravaClient.HttpClient.Get(url)
			if err != nil {
				return err
			}
		}
		return forwardResponse(tileResponse, rw)
	}

	if err := http.ListenAndServe(":8080", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Println(r.RequestURI)
		if err := serveTile(rw, r); err != nil {
			logger.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
		}
	})); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Wait()
}

func forwardResponse(res *http.Response, rw http.ResponseWriter) error {
	rw.WriteHeader(res.StatusCode)
	_, err := io.Copy(rw, res.Body)
	if err != nil {
		return err
	}
	for key, values := range res.Header {
		for _, val := range values {
			rw.Header().Add(key, val)
		}
	}
	return nil
}
