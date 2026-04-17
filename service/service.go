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
	"strings"

	"github.com/apexskier/strava-tile-proxy/strava"
	"github.com/pkg/errors"
)

var logger = log.New(os.Stdout, "service", log.LstdFlags)

var tileXYZRe = regexp.MustCompile(`/(?P<z>\d+)/(?P<x>\d+)/(?P<y>\d+)$`)

type Service struct {
	stravaClient strava.Client
	logger       *log.Logger

	apiToken string

	personalHeatmapDomain string
	globalHeatmapDomain   string

	revealPrivacyZones           bool
	revealOnlyMeActivities       bool
	revealFollowerOnlyActivities bool
	revealPublicActivities       bool
}

func New() (*Service, error) {
	rememberToken := os.Getenv("STRAVA_REMEMBER_TOKEN")
	if rememberToken == "" {
		return nil, errors.New("missing STRAVA_REMEMBER_TOKEN")
	}
	stravaSession := os.Getenv("STRAVA4_SESSION")
	if stravaSession == "" {
		return nil, errors.New("missing STRAVA4_SESSION")
	}
	stravaClient, err := strava.NewClient(rememberToken, stravaSession)
	if err != nil {
		return nil, err
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

	apiToken := os.Getenv("API_TOKEN")

	return &Service{
		stravaClient:                 stravaClient,
		logger:                       logger,
		apiToken:                     apiToken,
		personalHeatmapDomain:        strava.PersonalHeatmapDomain,
		globalHeatmapDomain:          strava.GlobalHeatmapDomain,
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
	sports    string
	heatColor strava.Heat
}

func (s *Service) extractParams(u *url.URL) (p Params, err error) {
	q := u.Query()

	var providedToken string
	providedTokens := q["api_token"]
	if len(providedTokens) > 0 {
		providedToken = providedTokens[0]
	}
	if providedToken != s.apiToken {
		return p, ErrBadQuery{query: "api_token", err: errors.New("incorrect api token")}
	}

	if heats, ok := q["color"]; ok && len(heats) > 0 {
		p.heatColor, err = strava.ParseHeat(heats[0])
		if err != nil {
			return p, ErrBadQuery{query: "color", err: err}
		}
	}

	p.sports = string(strava.SportAll)
	if sports, ok := q["sports"]; ok && len(sports) > 0 {
		p.sports = strings.Join(sports, ",")
	}

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

	return
}

func (s *Service) ServeGlobalTile(rw http.ResponseWriter, r *http.Request) error {
	p, err := s.extractParams(r.URL)
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

	if p.heatColor == "" {
		p.heatColor = strava.HeatBlue
	}

	tileQueryParams := url.Values{
		"v": []string{"19"},
	}
	url := fmt.Sprintf(
		s.globalHeatmapDomain+strava.GlobalHeatmapPath,
		p.sports,
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
		// refresh CloudFront cookies and retry once
		s.logger.Println("refreshing CloudFront cookies")
		if err := s.stravaClient.RefreshCloudFrontCookies(); err != nil {
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
	p, err := s.extractParams(r.URL)
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

	if p.heatColor == "" {
		p.heatColor = strava.HeatOrange
	}

	tileQueryParams := url.Values{
		strava.ParamFilterType:           []string{string(p.sports)},
		strava.ParamRespectPrivacyZones:  []string{strconv.FormatBool(!s.revealPrivacyZones)},
		strava.ParamIncludeEveryone:      []string{strconv.FormatBool(s.revealPublicActivities)},
		strava.ParamIncludeFollowersOnly: []string{strconv.FormatBool(s.revealFollowerOnlyActivities)},
		strava.ParamIncludeOnlyMe:        []string{strconv.FormatBool(s.revealOnlyMeActivities)},
	}
	athleteID, err := s.stravaClient.AthleteID()
	if err != nil {
		return err
	}
	url := fmt.Sprintf(
		s.personalHeatmapDomain+strava.PersonalHeatmapPath,
		athleteID,
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
		// refresh CloudFront cookies and retry once
		s.logger.Println("refreshing CloudFront cookies")
		if err := s.stravaClient.RefreshCloudFrontCookies(); err != nil {
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
