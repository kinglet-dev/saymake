package outline

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
)

// Footprint returns the outline of a length × width box (length along x):
// a superellipse of the given roundness, or, with 3 or more sides, a
// regular polygon resting on a flat side, stretched to fill the box.
func Footprint(length, width, roundness float64, sides, points int) (geom.Outline, error) {
	if sides < 3 {
		return Superellipse(length/2, width/2, roundness, points)
	}
	// Starting half a side before straight down puts the first edge level
	// at the bottom.
	p, _ := Polygon(sides, 1, -math.Pi/2-math.Pi/float64(sides)) // sides ≥ 3, radius 1
	if length <= 0 || width <= 0 {
		return nil, errNotPositive
	}
	return Fit(p, length, width), nil
}
