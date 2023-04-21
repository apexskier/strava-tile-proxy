package service

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/apexskier/strava-tile-proxy/strava"
	"github.com/pkg/errors"
)

var logger = log.New(os.Stdout, "service", log.LstdFlags)

var tileXYZRe = regexp.MustCompile(`/(?P<z>\d+)/(?P<x>\d+)/(?P<y>\d+)$`)

type Service struct {
	stravaClient strava.Client
	logger       *log.Logger
	rand         *rand.Rand

	personalHeatmapDomain string
	globalHeatmapDomain   string

	athleteID                    string
	revealPrivacyZones           bool
	revealOnlyMeActivities       bool
	revealFollowerOnlyActivities bool
	revealPublicActivities       bool
}

func New() (*Service, error) {
	email, err := mail.ParseAddress(os.Getenv("STRAVA_EMAIL"))
	if err != nil {
		return nil, errors.Wrap(err, "bad STRAVA_EMAIL")
	}
	password := os.Getenv("STRAVA_PASSWORD")
	if password == "" {
		return nil, errors.New("missing STRAVA_PASSWORD")
	}
	stravaClient, err := strava.NewClient(email.Address, password)
	if err != nil {
		return nil, err
	}

	athleteID := os.Getenv("ATHLETE_ID")
	if athleteID == "" {
		return nil, errors.New("missing ATHLETE_ID")
	}
	revealPrivacyZones, err := strconv.ParseBool(os.Getenv("REVEAL_PRIVACY_ZONES"))
	if err != nil {
		return nil, errors.Wrap(err, "bad REVEAL_PRIVACY_ZONES")
	}
	revealOnlyMeActivities, err := strconv.ParseBool(os.Getenv("REVEAL_ONLY_ME_ACTIVITIES"))
	if err != nil {
		return nil, errors.Wrap(err, "bad REVEAL_ONLY_ME_ACTIVITIES")
	}
	revealFollowerOnlyActivities, err := strconv.ParseBool(os.Getenv("REVEAL_FOLLOWER_ONLY_ACTIVITIES"))
	if err != nil {
		return nil, errors.Wrap(err, "bad REVEAL_FOLLOWER_ONLY_ACTIVITIES")
	}
	revealPublicActivities, err := strconv.ParseBool(os.Getenv("REVEAL_PUBLIC_ACTIVITIES"))
	if err != nil {
		return nil, errors.Wrap(err, "bad REVEAL_PUBLIC_ACTIVITIES")
	}

	return &Service{
		stravaClient:                 stravaClient,
		logger:                       logger,
		rand:                         rand.New(rand.NewSource(rand.Int63())),
		personalHeatmapDomain:        strava.PersonalHeatmapDomain,
		globalHeatmapDomain:          strava.GlobalHeatmapDomain,
		athleteID:                    athleteID,
		revealPrivacyZones:           revealPrivacyZones,
		revealOnlyMeActivities:       revealOnlyMeActivities,
		revealFollowerOnlyActivities: revealFollowerOnlyActivities,
		revealPublicActivities:       revealPublicActivities,
	}, nil
}

var ErrNotFound = errors.New("not found")

type ErrBadCoord struct {
	coord string
}

func (err ErrBadCoord) Error() string {
	return fmt.Sprintf("invalid tile %s", err.coord)
}

type ErrBadQuery struct {
	err   error
	query string
}

func (err ErrBadQuery) Unwrap() error {
	return err.err
}

func (err ErrBadQuery) Error() string {
	return fmt.Sprintf("invalid query parameter %s: %v", err.query, err.err)
}

type Params struct {
	x         uint64
	y         uint64
	z         uint64
	sport     strava.Sport
	heatColor strava.Heat
}

func extractParams(u *url.URL) (p Params, err error) {
	tileRouteMatches := tileXYZRe.FindStringSubmatch(u.Path)
	if len(tileRouteMatches) == 0 {
		return p, ErrNotFound
	}
	p.z, err = strconv.ParseUint(tileRouteMatches[1], 10, 64)
	if err != nil {
		return p, ErrBadCoord{coord: "z"}
	}
	p.x, err = strconv.ParseUint(tileRouteMatches[2], 10, 64)
	if err != nil {
		return p, ErrBadCoord{coord: "x"}
	}
	p.y, err = strconv.ParseUint(tileRouteMatches[3], 10, 64)
	if err != nil {
		return p, ErrBadCoord{coord: "y"}
	}

	q := u.Query()

	p.heatColor = strava.HeatRed
	if heats, ok := q["color"]; ok && len(heats) > 0 {
		p.heatColor, err = strava.ParseHeat(heats[0])
		if err != nil {
			return p, ErrBadQuery{query: "color", err: err}
		}
	}

	p.sport = strava.SportAll
	if heats, ok := q["sport"]; ok && len(heats) > 0 {
		p.sport, err = strava.ParseSport(heats[0])
		if err != nil {
			return p, ErrBadQuery{query: "sport", err: err}
		}
	}

	return
}

func (s *Service) ServeGlobalTile(rw http.ResponseWriter, r *http.Request) error {
	p, err := extractParams(r.URL)
	if err != nil {
		var badCoordErr ErrBadCoord
		var badQueryErr ErrBadQuery
		if errors.As(err, &badCoordErr) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
			return nil
		} else if errors.As(err, &badQueryErr) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
			return nil
		} else if errors.Is(err, ErrNotFound) {
			rw.WriteHeader(http.StatusNotFound)
			return nil
		}
		return err
	}

	heatmapServer := strava.HeatmapServers[s.rand.Intn(len(strava.HeatmapServers))]

	tileQueryParams := url.Values{
		// "v": []string{strconv.FormatUint(19, 10)},
	}
	url := fmt.Sprintf(
		s.globalHeatmapDomain+strava.GlobalHeatmapPath,
		heatmapServer,
		p.sport,
		p.heatColor,
		p.z,
		p.x,
		p.y,
		tileQueryParams.Encode(),
	)
	tileResponse, err := s.stravaClient.HttpClient().Get(url)
	if err != nil {
		return err
	}
	if tileResponse.StatusCode == http.StatusForbidden || tileResponse.StatusCode == http.StatusUnauthorized {
		// authenticate and retry once
		s.logger.Println("auto-authenticating cf")
		if err := s.stravaClient.CloudFrontAuth(heatmapServer); err != nil {
			return err
		}
		tileResponse, err = s.stravaClient.HttpClient().Get(url)
		if err != nil {
			return err
		}
	}
	return forwardResponse(tileResponse, rw)
}

func (s *Service) ServePersonalTile(rw http.ResponseWriter, r *http.Request) error {
	p, err := extractParams(r.URL)
	if err != nil {
		var badCoordErr ErrBadCoord
		var badQueryErr ErrBadQuery
		if errors.As(err, &badCoordErr) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
			return nil
		} else if errors.As(err, &badQueryErr) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
			return nil
		} else if errors.Is(err, ErrNotFound) {
			rw.WriteHeader(http.StatusNotFound)
			return nil
		}
		return err
	}

	tileQueryParams := url.Values{
		"filter_type":            []string{string(p.sport)},
		"filter_start":           []string{"2011-01-01"},
		"filter_end":             []string{time.Now().Format("2006-01-02")},
		"respect_privacy_zones":  []string{strconv.FormatBool(!s.revealPrivacyZones)},
		"include_everyone":       []string{strconv.FormatBool(s.revealPublicActivities)},
		"include_followers_only": []string{strconv.FormatBool(s.revealFollowerOnlyActivities)},
		"include_only_me":        []string{strconv.FormatBool(s.revealOnlyMeActivities)},
	}
	url := fmt.Sprintf(
		s.personalHeatmapDomain+strava.PersonalHeatmapPath,
		s.athleteID,
		p.heatColor,
		p.z,
		p.x,
		p.y,
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
