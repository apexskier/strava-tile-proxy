package strava

import "errors"

var HeatmapServers = []string{"a", "b", "c"}

const (
	StravaDomain          = "https://www.strava.com"
	PersonalHeatmapDomain = "https://personal-heatmaps-external.strava.com"
	GlobalHeatmapDomain   = "https://heatmap-external-%s.strava.com"
	PersonalHeatmapPath   = "/tiles/%s/%s/%d/%d/%d@2x.png?%s"
	GlobalHeatmapPath     = "/tiles-auth/%s/%s/%d/%d/%d@2x.png?%s"
)

type Heat string

const (
	HeatOrange     Heat = "orange"
	HeatRed        Heat = "red"
	HeatMobileBlue Heat = "mobileblue"
	HeatBlue       Heat = "blue"
	HeatBlueRed    Heat = "bluered"
	HeatPurple     Heat = "purple"
	HeatGray       Heat = "gray"
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

type Sport string

const (
	SportAll    Sport = "all"
	SportRide   Sport = "ride"
	SportRun    Sport = "run"
	SportWater  Sport = "water"
	SportWinter Sport = "winter"
)

func ParseSport(raw string) (Sport, error) {
	if raw != string(SportAll) &&
		raw != string(SportRide) &&
		raw != string(SportRun) &&
		raw != string(SportWater) &&
		raw != string(SportWinter) {
		return "", errors.New("unknown sport")
	}
	return Sport(raw), nil
}
