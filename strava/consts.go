package strava

import "errors"

const (
	StravaDomain          = "https://www.strava.com"
	PersonalHeatmapDomain = "https://personal-heatmaps-external.strava.com"
	GlobalHeatmapDomain   = "https://content-a.strava.com" // previously this was a global-... url that rotated between a/b/c. content-b/c don't resolve anymore
	PersonalHeatmapPath   = "/tiles/%s/%s/%d/%d/%d@2x.png?%s"
	GlobalHeatmapPath     = "/identified/globalheat/%s/%s/%d/%d/%d@2x.png?%s"
)

const (
	ParamFilterType           = "filter_type"            // String
	ParamFilterStart          = "filter_start"           // String 2006-01-02
	ParamFilterEnd            = "filter_end"             // String 2006-01-02
	ParamIncludeEveryone      = "include_everyone"       // Bool
	ParamIncludeFollowersOnly = "include_followers_only" // Bool
	ParamIncludeOnlyMe        = "include_only_me"        // Bool
	ParamRespectPrivacyZones  = "respect_privacy_zones"  // Bool
	ParamIncludeCommutes      = "include_commutes"       // Bool
)

func ParseHeat(raw string) (Heat, error) {
	if raw != string(HeatOrange) &&
		raw != string(HeatRed) &&
		raw != string(HeatMobileBlue) &&
		raw != string(HeatBlue) &&
		raw != string(HeatBlueRed) &&
		raw != string(HeatPurple) &&
		raw != string(HeatGray) {
		return "", errors.New("unknown heat color")
	}
	return Heat(raw), nil
}
