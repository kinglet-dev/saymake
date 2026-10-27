package vessel

import "github.com/kinglet-dev/saymake/domain/geom"

// footRing is how wide the ring a footed vessel stands on is, in mm.
const footRing = 4.0

type ringer func(f at, patterned bool) []int

// footRadius is the radius of the foot's recess: the base's closest point
// to the axis (a polygon's sides, the bottom of a flute) less the ring.
func (c settings) footRadius() float64 { return c.profile(0)*c.closest() - footRing }

// closeBottom closes the bottom: a flat base, or a foot ring around a recess whose
// ceiling is foot mm up; with a drainage hole through the floor either way.
func (c settings) closeBottom(m *geom.Mesh, ring ringer, outer0, inner0 []int) {
	low := outer0 // the edge the underside starts from
	underside := 0.0
	if c.foot > 0 {
		rf := c.footRadius()
		standing, recess := ring(flat(0, rf), false), ring(flat(c.foot, rf), false)
		m.Band(outer0, standing, true) // the ring it stands on, facing down
		m.Band(standing, recess, true) // the recess wall, facing the axis
		low, underside = recess, c.foot
	}
	if c.drain <= 0 {
		m.Fan(m.Add(geom.Vec3{Z: underside}), low, true)
		m.Fan(m.Add(geom.Vec3{Z: c.floor}), inner0, false)
		return
	}
	holeLow, holeHigh := ring(flat(underside, c.drain/2), false), ring(flat(c.floor, c.drain/2), false)
	m.Band(low, holeLow, true)      // the underside around the hole, facing down
	m.Band(inner0, holeHigh, false) // the floor around the hole, facing up
	m.Band(holeLow, holeHigh, true) // the hole's wall, facing the axis
}
