package vessel

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
)

const (
	baseAround = 192 // points around at least
	rings      = 96  // rings up the height
)

func (c settings) around() int {
	n := baseAround
	if c.flutes > 0 {
		n = max(n, 8*c.flutes)
	}
	if c.sides >= 3 {
		n = c.sides * int(math.Ceil(float64(n)/float64(c.sides)))
	}
	return n
}

// rimDrop is how far below the top the outer wall's rim edge starts.
func (c settings) rimDrop() float64 {
	if c.rimStyle > 0 {
		return c.rim
	}
	return 0
}

// mesh builds the solid as structured grids of rings: outer wall (up to the
// rim edge), the rim edge, the top, inner wall, and the base (see base.go),
// plus a handle (handle.go). Faces wind counterclockwise seen from outside.
func (c settings) mesh() *geom.Mesh {
	m := &geom.Mesh{}
	n := c.around()
	ring := func(f at, patterned bool) []int {
		ids := make([]int, n)
		for i := range ids {
			t := 2 * math.Pi * float64(i) / float64(n)
			z, r := f(t)
			if patterned {
				r *= c.pattern(t, z/c.height)
			}
			ids[i] = m.Add(geom.Vec3{X: r * math.Cos(t), Y: r * math.Sin(t), Z: z})
		}
		return ids
	}
	outer, inner := make([][]int, rings+1), make([][]int, rings+1)
	for j := range rings + 1 {
		u := float64(j) / rings
		outer[j] = ring(func(t float64) (float64, float64) {
			z := (c.rimTop(t) - c.rimDrop()) * u
			return z, c.profile(z / c.height)
		}, true)
		inner[j] = ring(func(t float64) (float64, float64) {
			z := c.floor + (c.rimTop(t)-c.floor)*u
			return z, c.profile(z/c.height) - c.insetFactor()
		}, true)
	}
	for j := range rings {
		m.Band(outer[j], outer[j+1], false)
		m.Band(inner[j], inner[j+1], true)
	}
	edge := outer[rings]
	for k := range c.rimEdge(c.height) {
		next := ring(func(t float64) (float64, float64) {
			l := c.rimEdge(c.rimTop(t))[k]
			return l.z, l.r
		}, true)
		m.Band(edge, next, false)
		edge = next
	}
	m.Band(edge, inner[rings], false) // the top
	c.closeBottom(m, ring, outer[0], inner[0])
	if c.handle > 0 {
		c.addHandle(m)
	}
	if c.lean > 0 {
		c.leanOver(m)
	}
	return m
}

// at gives a ring's height and radius (before the pattern) at angle t.
type at func(t float64) (z, r float64)

// flat is a ring at one height and radius.
func flat(z, r float64) at { return func(float64) (float64, float64) { return z, r } }
