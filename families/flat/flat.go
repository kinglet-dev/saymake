// Package flat is the flat-outline family for laser cutting: coasters,
// plaques and tags, with an optional scored border and hanging hole.
package flat

import (
	"errors"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/outline"
	"github.com/kinglet-dev/saymake/domain/param"
)

const (
	points     = 192 // points around a round outline
	holePoints = 64  // points around the hanging hole
	holeMargin = 1.0 // mm of material kept between the hole and the edge
)

var specs = param.Set{
	{Name: "length", Label: "Length", Unit: "mm", Min: 10, Max: 600, Default: 100, Step: 1},
	{Name: "width", Label: "Width", Unit: "mm", Min: 10, Max: 600, Default: 100, Step: 1},
	{Name: "roundness", Label: "Squareness", Min: 2, Max: 20, Default: 2, Step: 0.1},
	{Name: "sides", Label: "Sides (0 = rounded)", Min: 0, Max: 12, Default: 0, Step: 1, Whole: true},
	{Name: "corner", Label: "Corner radius (with sides)", Unit: "mm", Min: 0, Max: 50, Default: 0, Step: 0.5},
	{Name: "border", Label: "Scored border", Unit: "mm", Min: 0, Max: 50, Default: 0, Step: 0.5},
	{Name: "hole", Label: "Hanging hole", Unit: "mm", Min: 0, Max: 20, Default: 0, Step: 0.5},
	{Name: "holeFrom", Label: "Hole from top", Unit: "mm", Min: 3, Max: 100, Default: 8, Step: 0.5},
}

// Flat is the family.
type Flat struct{}

// New returns the family.
func New() Flat { return Flat{} }

// ID implements family.Family.
func (Flat) ID() string { return "flat" }

// Label implements family.Family.
func (Flat) Label() string { return "Flat shape" }

// MainSize implements family.Family.
func (Flat) MainSize() string { return "length" }

// Proportional implements family.Family.
func (Flat) Proportional() []string { return []string{"width"} }

// Params implements family.Family.
func (Flat) Params() param.Set { return append(param.Set(nil), specs...) }

// Build implements family.Family.
func (Flat) Build(v param.Values) (family.Shape, error) {
	if err := specs.Require(v); err != nil {
		return family.Shape{}, err
	}
	outer, err := outline.FootprintRounded(v["length"], v["width"], v["roundness"], int(v["sides"]), v["corner"], points)
	if errors.Is(err, outline.ErrCornerTooBig) {
		return family.Shape{}, errors.New("The corner radius is too big for this shape. Make it smaller or the shape bigger.")
	}
	if err != nil {
		return family.Shape{}, err
	}
	s := family.Shape{Outlines: []geom.Outline{outer}}
	if v["border"] > 0 {
		score, err := outline.Inset(outer, v["border"])
		if err != nil {
			return family.Shape{}, errors.New("The border is too far in for this shape. Move it out or make the shape bigger.")
		}
		s.Scores = append(s.Scores, score)
	}
	if v["hole"] > 0 {
		hole, err := hangingHole(outer, v["hole"], v["holeFrom"])
		if err != nil {
			return family.Shape{}, err
		}
		s.Outlines = append(s.Outlines, hole)
	}
	return s, nil
}

// hangingHole returns a round hole centered below the top edge, refusing
// one that would come closer than holeMargin to the edge.
func hangingHole(outer geom.Outline, diameter, from float64) (geom.Outline, error) {
	errFit := errors.New("The hole doesn't fit inside the shape. Make it smaller or move it further in.")
	margin, err := outline.Inset(outer, holeMargin)
	if err != nil {
		return nil, errFit
	}
	_, hi := outer.Bounds()
	circle, _ := outline.Polygon(holePoints, diameter/2, 0) // diameter > 0 here
	center := geom.Vec2{Y: hi.Y - from}
	hole := make(geom.Outline, len(circle))
	for i, p := range circle {
		hole[i] = p.Add(center)
		if !outline.Contains(margin, hole[i]) {
			return nil, errFit
		}
	}
	return hole, nil
}
