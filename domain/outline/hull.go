package outline

import (
	"cmp"
	"slices"

	"github.com/kinglet-dev/saymake/domain/geom"
)

// Hull returns the convex hull of the points, counterclockwise from the
// lowest of the leftmost points, without points on its edges; or nothing
// when the points enclose no area. It uses Andrew's monotone chain
// (A. M. Andrew, 1979): sort by x then y, then build the lower and upper
// chains, dropping any point that doesn't make a left turn.
func Hull(pts []geom.Vec2) geom.Outline {
	p := slices.Clone(pts)
	slices.SortFunc(p, func(a, b geom.Vec2) int {
		return cmp.Or(cmp.Compare(a.X, b.X), cmp.Compare(a.Y, b.Y))
	})
	var h geom.Outline
	chain := func(pt geom.Vec2, floor int) {
		for len(h) > floor && h[len(h)-1].Sub(h[len(h)-2]).Cross(pt.Sub(h[len(h)-2])) <= 0 {
			h = h[:len(h)-1]
		}
		h = append(h, pt)
	}
	for _, pt := range p {
		chain(pt, 1)
	}
	lower := len(h)
	for i := len(p) - 2; i >= 0; i-- {
		chain(p[i], lower)
	}
	if len(h) < 4 {
		return nil
	}
	return h[:len(h)-1]
}
