package strava

import "errors"

const TilesNamespace = "14856714"

type Heat string

const (
	HeatOrange  Heat = "orange"
	HeatRed     Heat = "hot"
	HeatBlue    Heat = "blue"
	HeatBlueRed Heat = "bluered"
	HeatPurple  Heat = "purple"
	HeatGray    Heat = "gray"
)

func ParseHeat(raw string) (Heat, error) {
	if raw != string(HeatOrange) &&
		raw != string(HeatRed) &&
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
