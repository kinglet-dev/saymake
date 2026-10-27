// Package meshcheck decides whether a triangle mesh is a sound solid: closed,
// consistently wound, outward-facing and made of finite numbers.
//
// Closed and oriented means every edge is used by exactly two triangles, once
// in each direction (an oriented 2-manifold; Botsch et al., Polygon Mesh
// Processing, 2010). Volume is the sum of signed tetrahedra from the origin
// (divergence theorem; Zhang & Chen, 2001), positive when faces point out.
package meshcheck

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kinglet-dev/saymake/domain/geom"
)

// minArea2 is the smallest squared doubled area a triangle may have
// (1e-6 mm² of area).
const minArea2 = 4e-12

// Report is what Inspect found.
type Report struct {
	Triangles                                  int
	OpenEdges, MisorientedEdges, OverusedEdges int
	Degenerate                                 int // triangles with no area
	Parts, InsideOutParts                      int // connected parts, and those with no positive volume
	BadIndex, NonFinite                        bool
	Volume                                     float64
	Min, Max                                   geom.Vec3
}

// Inspect examines a mesh.
func Inspect(m geom.Mesh) Report {
	r := Report{Triangles: len(m.Triangles)}
	for i, v := range m.Vertices {
		if !v.Finite() {
			r.NonFinite = true
		}
		if i == 0 {
			r.Min, r.Max = v, v
		}
		r.Min = geom.Vec3{X: min(r.Min.X, v.X), Y: min(r.Min.Y, v.Y), Z: min(r.Min.Z, v.Z)}
		r.Max = geom.Vec3{X: max(r.Max.X, v.X), Y: max(r.Max.Y, v.Y), Z: max(r.Max.Z, v.Z)}
	}
	directed := make(map[[2]int]int, 3*len(m.Triangles))
	for _, t := range m.Triangles {
		for _, i := range t {
			if i < 0 || i >= len(m.Vertices) {
				r.BadIndex = true
				return r
			}
		}
		a, b, c := m.Vertices[t[0]], m.Vertices[t[1]], m.Vertices[t[2]]
		r.Volume += a.Dot(b.Cross(c)) / 6
		if n := b.Sub(a).Cross(c.Sub(a)); n.Dot(n) <= minArea2 {
			r.Degenerate++
		}
		for k := range 3 {
			directed[[2]int{t[k], t[(k+1)%3]}]++
		}
	}
	r.countEdges(directed)
	r.Parts, r.InsideOutParts = parts(m)
	return r
}

func (r *Report) countEdges(directed map[[2]int]int) {
	for e, n := range directed {
		back := directed[[2]int{e[1], e[0]}]
		if e[0] > e[1] && back > 0 {
			continue // counted from the other direction
		}
		switch total := n + back; {
		case total == 1:
			r.OpenEdges++
		case total > 2:
			r.OverusedEdges++
		case n != back:
			r.MisorientedEdges++
		}
	}
}

// Sound returns nil for a closed, oriented, outward solid, or an error
// saying what is wrong.
func (r Report) Sound() error {
	switch {
	case r.Triangles == 0:
		return errors.New("the mesh has no triangles")
	case r.BadIndex:
		return errors.New("a triangle names a missing vertex")
	case r.NonFinite:
		return errors.New("a vertex is not a finite number")
	}
	var p []string
	if r.OpenEdges > 0 {
		p = append(p, fmt.Sprintf("%d open edges", r.OpenEdges))
	}
	if r.MisorientedEdges > 0 {
		p = append(p, fmt.Sprintf("%d edges where faces are wound inconsistently", r.MisorientedEdges))
	}
	if r.Degenerate > 0 {
		p = append(p, fmt.Sprintf("%d triangles with no area", r.Degenerate))
	}
	if r.OverusedEdges > 0 {
		p = append(p, fmt.Sprintf("%d edges shared by more than two faces", r.OverusedEdges))
	}
	switch {
	case len(p) > 0:
	case r.Volume <= 0:
		p = append(p, "it is inside out (its volume is not positive)")
	case r.InsideOutParts == 1:
		p = append(p, "1 part is inside out")
	case r.InsideOutParts > 1:
		p = append(p, fmt.Sprintf("%d parts are inside out", r.InsideOutParts))
	}
	if len(p) > 0 {
		return fmt.Errorf("the mesh is not a closed solid: %s", strings.Join(p, ", "))
	}
	return nil
}
