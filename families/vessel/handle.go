package vessel

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
)

const (
	handleSteps = 48 // rings along the handle
	handleSides = 16 // points around its cross-section
)

// handleSpan is where the handle's ends are: the handle setting is its
// share of the height, centered on it.
func (c settings) handleSpan() (bottom, top float64) {
	half := c.handle * c.height / 2
	return c.height/2 - half, c.height/2 + half
}

// handleEnd is the x of a handle end at height z: the middle of the wall
// at angle 0, where the handle sits.
func (c settings) handleEnd(z float64) float64 {
	return c.profile(z/c.height)*c.pattern(0, z/c.height) - c.insetFactor()/2
}

// addHandle adds a C-shaped handle as its own closed tube: a circle of
// diameter handleThick swept along a half ellipse in the xz plane, from
// the middle of the wall at the top end, out by handleReach, and back at
// the bottom end. Its ends meet the wall at right angles, inside it;
// slicers join overlapping closed parts.
func (c settings) addHandle(m *geom.Mesh) {
	bottom, top := c.handleSpan()
	xt, xb := c.handleEnd(top), c.handleEnd(bottom)
	a := c.handleThick / 2
	var rings [][]int
	for i := 0; i <= handleSteps; i++ {
		s := math.Pi * float64(i) / handleSteps
		w := (1 - math.Cos(s)) / 2
		half := (top - bottom) / 2
		p := geom.Vec3{X: xt + (xb-xt)*w + c.handleReach*math.Sin(s), Z: c.height/2 + half*math.Cos(s)}
		dx, dz := (xb-xt)*math.Sin(s)/2+c.handleReach*math.Cos(s), -half*math.Sin(s)
		l := math.Hypot(dx, dz)
		nx, nz := -dz/l, dx/l // the path's normal in the xz plane
		ids := make([]int, handleSides)
		for k := range ids {
			phi := 2 * math.Pi * float64(k) / handleSides
			ids[k] = m.Add(geom.Vec3{X: p.X + a*math.Cos(phi)*nx, Y: a * math.Sin(phi), Z: p.Z + a*math.Cos(phi)*nz})
		}
		rings = append(rings, ids)
	}
	for i := range handleSteps {
		m.Band(rings[i], rings[i+1], true)
	}
	m.Fan(m.Add(geom.Vec3{X: xt, Z: top}), rings[0], false)
	m.Fan(m.Add(geom.Vec3{X: xb, Z: bottom}), rings[handleSteps], true)
}
