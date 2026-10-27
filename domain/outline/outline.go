// Package outline makes closed convex outlines: superellipses (Lamé curves)
// and regular polygons, and insets them by a wall thickness.
package outline

import (
	"errors"
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
)

// Superellipse returns the Lamé curve |x/a|^n + |y/b|^n = 1 as points
// counterclockwise from (a, 0), using Piet Hein's parametric form
// x = a·sgn(cos t)·|cos t|^(2/n), y = b·sgn(sin t)·|sin t|^(2/n).
// n = 2 is an ellipse; larger n approaches a rectangle.
func Superellipse(a, b, n float64, points int) (geom.Outline, error) {
	switch {
	case a <= 0 || b <= 0:
		return nil, errNotPositive
	case n <= 0:
		return nil, errors.New("roundness must be positive")
	case points < 3:
		return nil, errors.New("an outline needs at least 3 points")
	}
	o := make(geom.Outline, points)
	for k := range o {
		t := 2 * math.Pi * float64(k) / float64(points)
		o[k] = geom.Vec2{X: a * signedPow(math.Cos(t), 2/n), Y: b * signedPow(math.Sin(t), 2/n)}
	}
	return o, nil
}

func signedPow(v, e float64) float64 {
	return math.Copysign(math.Pow(math.Abs(v), e), v)
}

// Polygon returns a regular polygon with the given circumradius, its first
// corner at angle turn (radians), counterclockwise.
func Polygon(sides int, radius, turn float64) (geom.Outline, error) {
	switch {
	case sides < 3:
		return nil, errors.New("a polygon needs at least 3 sides")
	case radius <= 0:
		return nil, errors.New("a polygon's radius must be positive")
	}
	o := make(geom.Outline, sides)
	for k := range o {
		t := turn + 2*math.Pi*float64(k)/float64(sides)
		o[k] = geom.Vec2{X: radius * math.Cos(t), Y: radius * math.Sin(t)}
	}
	return o, nil
}

// ErrTooThick means an inset would collapse or turn the outline inside out.
var ErrTooThick = errors.New("the wall is too thick for this outline")

// Inset moves every edge of a convex, counterclockwise outline inward by d:
// each edge is shifted along its inward normal and each corner becomes the
// intersection of its two shifted edges (convex polygon offsetting). Where
// two edges run straight on, the corner is shifted along their normal.
func Inset(o geom.Outline, d float64) (geom.Outline, error) {
	n := len(o)
	dir := make([]geom.Vec2, n)
	for i := range o {
		e := o[(i+1)%n].Sub(o[i])
		dir[i] = e.Scale(1 / math.Hypot(e.X, e.Y))
	}
	in := make(geom.Outline, n)
	for i := range o {
		prev := dir[(i+n-1)%n]
		a := o[i].Add(leftNormal(prev).Scale(d)) // on the shifted previous edge
		b := o[i].Add(leftNormal(dir[i]).Scale(d))
		den := prev.Cross(dir[i])
		if math.Abs(den) < 1e-12 {
			in[i] = b
			continue
		}
		s := b.Sub(a).Cross(dir[i]) / den
		in[i] = a.Add(prev.Scale(s))
	}
	for i := range in {
		if in[(i+1)%n].Sub(in[i]).Dot(dir[i]) <= 1e-9 {
			return nil, ErrTooThick
		}
	}
	return in, nil
}

func leftNormal(u geom.Vec2) geom.Vec2 { return geom.Vec2{X: -u.Y, Y: u.X} }

// Fit stretches an outline so its bounding box is w × h, centered on the
// origin.
func Fit(o geom.Outline, w, h float64) geom.Outline {
	lo, hi := o.Bounds()
	mid := lo.Add(hi).Scale(0.5)
	sx, sy := w/(hi.X-lo.X), h/(hi.Y-lo.Y)
	out := make(geom.Outline, len(o))
	for i, p := range o {
		d := p.Sub(mid)
		out[i] = geom.Vec2{X: d.X * sx, Y: d.Y * sy}
	}
	return out
}

// Contains reports whether p is strictly inside a convex, counterclockwise
// outline.
func Contains(o geom.Outline, p geom.Vec2) bool {
	for i, a := range o {
		b := o[(i+1)%len(o)]
		if b.Sub(a).Cross(p.Sub(a)) <= 0 {
			return false
		}
	}
	return true
}

var errNotPositive = errors.New("half-width and half-length must be positive")
