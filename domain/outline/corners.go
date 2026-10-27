package outline

import (
	"errors"
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
)

// ErrCornerTooBig means a corner radius doesn't fit the outline.
var ErrCornerTooBig = errors.New("the corner radius is too big for this outline")

// arcStep is the largest angle one segment of a corner's arc turns.
const arcStep = math.Pi / 32

// RoundCorners rounds the corners of a convex, counterclockwise outline to
// radius r, keeping its edges where they were: it is the Minkowski sum of
// the outline inset by r and a disk of radius r, so each corner becomes an
// arc around the inset outline's corner, turning from one edge's outward
// normal to the next's.
func RoundCorners(o geom.Outline, r float64) (geom.Outline, error) {
	if r <= 0 {
		return o, nil
	}
	in, err := Inset(o, r)
	if err != nil {
		return nil, ErrCornerTooBig
	}
	n := len(in)
	var out geom.Outline
	for i, c := range in {
		before := outward(in[(i+n-1)%n], c)
		after := outward(c, in[(i+1)%n])
		a0 := math.Atan2(before.Y, before.X)
		sweep := math.Mod(math.Atan2(after.Y, after.X)-a0+2*math.Pi, 2*math.Pi)
		steps := max(1, int(math.Ceil(sweep/arcStep)))
		for k := 0; k <= steps; k++ {
			a := a0 + sweep*float64(k)/float64(steps)
			out = append(out, geom.Vec2{X: c.X + r*math.Cos(a), Y: c.Y + r*math.Sin(a)})
		}
	}
	return out, nil
}

// outward is the outward normal of the edge a→b of a counterclockwise outline.
func outward(a, b geom.Vec2) geom.Vec2 {
	d := b.Sub(a)
	return geom.Vec2{X: d.Y, Y: -d.X}
}

// FootprintRounded is Footprint with a polygon's corners rounded to radius
// corner (after fitting it to the box, so the arcs stay circular). A
// footprint without sides is already rounded and is left as it is.
func FootprintRounded(length, width, roundness float64, sides int, corner float64, points int) (geom.Outline, error) {
	o, err := Footprint(length, width, roundness, sides, points)
	if err != nil || sides < 3 {
		return o, err
	}
	return RoundCorners(o, corner)
}
