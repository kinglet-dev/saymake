package check

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/outline"
)

// onBed is how close to the lowest point a vertex must be to count as
// standing on the bed, in mm.
const onBed = 1e-6

// Stable warns when a solid's center of weight, empty, isn't above the
// convex hull of what it stands on (its lowest points), so it would tip
// over. The center is the volume-weighted mean of the signed tetrahedra's
// centroids (divergence theorem, as for the volume). Meshes with no volume
// are left to the solid check.
type Stable struct{}

// Run implements Check.
func (Stable) Run(s family.Shape, _ Context) []Finding {
	if s.Mesh == nil {
		return nil
	}
	m := *s.Mesh
	var volume float64
	var mx, my float64
	low := math.Inf(1)
	for _, t := range m.Triangles {
		a, b, c := m.Vertices[t[0]], m.Vertices[t[1]], m.Vertices[t[2]]
		v := a.Dot(b.Cross(c)) / 6
		volume += v
		mx += v * (a.X + b.X + c.X) / 4
		my += v * (a.Y + b.Y + c.Y) / 4
	}
	if volume <= 0 {
		return nil
	}
	for _, v := range m.Vertices {
		low = min(low, v.Z)
	}
	var feet []geom.Vec2
	for _, v := range m.Vertices {
		if v.Z <= low+onBed {
			feet = append(feet, geom.Vec2{X: v.X, Y: v.Y})
		}
	}
	base := outline.Hull(feet)
	if len(base) > 0 && outline.Contains(base, geom.Vec2{X: mx / volume, Y: my / volume}) {
		return nil
	}
	return []Finding{{Warning, "It may tip over when empty: its center of weight isn't above what it stands on. Make the base wider, or make it lean less."}}
}
