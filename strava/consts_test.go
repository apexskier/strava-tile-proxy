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
