package vessel

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
)

// lean shears the finished solid toward angle 0 (x grows by z·tan lean), so
// the base stays flat where it was and every horizontal slice keeps its
// shape and area: the volume is unchanged.
func (c settings) leanOver(m *geom.Mesh) {
	k := math.Tan(c.lean * math.Pi / 180)
	for i, v := range m.Vertices {
		m.Vertices[i].X = v.X + k*v.Z
	}
}

// leaning is the steepest slope from upright once leaning: the lean's
// slope adds to the profile's on the side it leans toward.
func (c settings) leaning(steepest float64) float64 {
	if c.lean <= 0 {
		return steepest
	}
	slope := math.Tan(steepest*math.Pi/180) + math.Tan(c.lean*math.Pi/180)
	return math.Atan(slope) * 180 / math.Pi
}
