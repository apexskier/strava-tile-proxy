package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/apexskier/strava-tile-proxy/strava"
)

var logger = log.New(os.Stdout, "service", log.LstdFlags)

var tileRouteRe = regexp.MustCompile(`^/tiles/(?P<z>\d+)/(?P<x>\d+)/(?P<y>\d+)$`)

type Service struct {
	stravaClient     strava.Client
	stravaHeatmapUrl string
	logger           *log.Logger

	revealPrivacyZones           bool
	revealOnlyMeActivities       bool
	revealFollowerOnlyActivities bool
	revealPublicActivities       bool
}

func New() (*Service, error) {
	stravaClient, err := strava.NewClient(
		os.Getenv("STRAVA_EMAIL"),
		os.Getenv("STRAVA_PASSWORD"),
	)
	if err != nil {
		return nil, err
	}

	revealPrivacyZones, err := strconv.ParseBool(os.Getenv("REVEAL_PRIVACY_ZONES"))
	if err != nil {
		return nil, err
	}
	revealOnlyMeActivities, err := strconv.ParseBool(os.Getenv("REVEAL_ONLY_ME_ACTIVITIES"))
	if err != nil {
		return nil, err
	}
	revealFollowerOnlyActivities, err := strconv.ParseBool(os.Getenv("REVEAL_FOLLOWER_ONLY_ACTIVITIES"))
	if err != nil {
		return nil, err
	}
	revealPublicActivities, err := strconv.ParseBool(os.Getenv("REVEAL_PUBLIC_ACTIVITIES"))
	if err != nil {
		return nil, err
	}

	return &Service{
		stravaClient:                 stravaClient,
		stravaHeatmapUrl:             "https://personal-heatmaps-external.strava.com",
		logger:                       logger,
		revealPrivacyZones:           revealPrivacyZones,
		revealOnlyMeActivities:       revealOnlyMeActivities,
		revealFollowerOnlyActivities: revealFollowerOnlyActivities,
		revealPublicActivities:       revealPublicActivities,
	}, nil
}

func (s *Service) ServeTile(rw http.ResponseWriter, r *http.Request) error {
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
		"respect_privacy_zones":  []string{strconv.FormatBool(!s.revealPrivacyZones)},
		"include_everyone":       []string{strconv.FormatBool(s.revealPublicActivities)},
		"include_followers_only": []string{strconv.FormatBool(s.revealFollowerOnlyActivities)},
		"include_only_me":        []string{strconv.FormatBool(s.revealOnlyMeActivities)},
	}
	url := fmt.Sprintf(
		"%s/tiles/%s/%s/%d/%d/%d@2x.png?%s",
		s.stravaHeatmapUrl,
		strava.TilesNamespace,
		heatColor,
		z,
		x,
		y,
		tileQueryParams.Encode(),
	)
	tileResponse, err := s.stravaClient.HttpClient().Get(url)
	if err != nil {
		return err
	}
	if tileResponse.StatusCode == http.StatusUnauthorized {
		// authenticate and retry once
		s.logger.Println("auto-authenticating")
		if err := s.stravaClient.Login(); err != nil {
			return err
		}
		tileResponse, err = s.stravaClient.HttpClient().Get(url)
		if err != nil {
			return err
		}
	}
	return forwardResponse(tileResponse, rw)
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
