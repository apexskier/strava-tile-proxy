package strava

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeat(t *testing.T) {
	color, err := ParseHeat("red")
	assert.Equal(t, HeatRed, color)
	assert.NoError(t, err)

	_, err = ParseHeat("garbage")
	assert.Error(t, err)
}

func TestParseSport(t *testing.T) {
	color, err := ParseSport("run")
	assert.Equal(t, SportRun, color)
	assert.NoError(t, err)

	_, err = ParseSport("garbage")
	assert.Error(t, err)
}
