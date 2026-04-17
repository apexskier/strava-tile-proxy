package strava

type Sport string

const (
	SportAll Sport = "all"

	// Foot
	SportRun      Sport = "sport_Run"
	SportTrailRun Sport = "sport_TrailRun"
	SportWalk     Sport = "sport_Walk"
	SportHike     Sport = "sport_Hike"

	// Cycle
	SportRide              Sport = "sport_Ride"
	SportMountainBikeRide  Sport = "sport_MountainBikeRide"
	SportGravelRide        Sport = "sport_GravelRide"
	SportEBikeRide         Sport = "sport_EBikeRide"
	SportEMountainBikeRide Sport = "sport_EMountainBikeRide"
	SportVelomobile        Sport = "sport_Velomobile"

	// Water
	SportCanoeing        Sport = "sport_Canoeing"
	SportKayaking        Sport = "sport_Kayaking"
	SportKitesurf        Sport = "sport_Kitesurf"
	SportRowing          Sport = "sport_Rowing"
	SportSail            Sport = "sport_Sail"
	SportStandUpPaddling Sport = "sport_StandUpPaddling"
	SportSurfing         Sport = "sport_Surfing"
	SportSwim            Sport = "sport_Swim"
	SportWindsurf        Sport = "sport_Windsurf"

	// Winter
	SportAlpineSki      Sport = "sport_AlpineSki"
	SportBackcountrySki Sport = "sport_BackcountrySki"
	SportIceSkate       Sport = "sport_IceSkate"
	SportNordicSki      Sport = "sport_NordicSki"
	SportSnowboard      Sport = "sport_Snowboard"
	SportSnowshoe       Sport = "sport_Snowshoe"

	// this isn't complete
)
