// Package tray is the tray-and-dish family: a footprint outline (a
// superellipse or a polygon) with a wall and a floor, optionally flared.
package tray

import (
	"errors"
	"math"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/outline"
	"github.com/kinglet-dev/saymake/domain/param"
)

// MaxTriangles bounds the mesh size (see the threat model).
const MaxTriangles = 270_000

const points = 192 // points around a round footprint

var specs = param.Set{
	{Name: "length", Label: "Length", Unit: "mm", Min: 20, Max: 500, Default: 250, Step: 1},
	{Name: "width", Label: "Width", Unit: "mm", Min: 20, Max: 500, Default: 180, Step: 1},
	{Name: "height", Label: "Height", Unit: "mm", Min: 5, Max: 150, Default: 30, Step: 1},
	{Name: "roundness", Label: "Squareness", Min: 2, Max: 20, Default: 5, Step: 0.1},
	{Name: "sides", Label: "Sides (0 = rounded)", Min: 0, Max: 12, Default: 0, Step: 1, Whole: true},
	{Name: "corner", Label: "Corner radius (with sides)", Unit: "mm", Min: 0, Max: 50, Default: 0, Step: 0.5},
	{Name: "flare", Label: "Flare", Unit: "°", Min: 0, Max: 30, Default: 0, Step: 1},
	{Name: "wall", Label: "Wall", Unit: "mm", Min: 0.8, Max: 10, Default: 2.5, Step: 0.1},
	{Name: "floor", Label: "Floor", Unit: "mm", Min: 0.8, Max: 10, Default: 3, Step: 0.1},
}

// Tray is the family.
type Tray struct{}

// New returns the family.
func New() Tray { return Tray{} }

// ID implements family.Family.
func (Tray) ID() string { return "tray" }

// Label implements family.Family.
func (Tray) Label() string { return "Tray" }

// MainSize implements family.Family.
func (Tray) MainSize() string { return "length" }

// Proportional implements family.Family.
func (Tray) Proportional() []string { return []string{"width"} }

// Params implements family.Family.
func (Tray) Params() param.Set { return append(param.Set(nil), specs...) }

// Build implements family.Family. The outer wall at height z is the
// footprint offset outward by z·tan(flare); the inner wall is that offset
// inward by the wall thickness (convex offsetting, see package outline).
func (Tray) Build(v param.Values) (family.Shape, error) {
	if err := specs.Require(v); err != nil {
		return family.Shape{}, err
	}
	h, wall, floor := v["height"], v["wall"], v["floor"]
	if floor > h-0.5 {
		return family.Shape{}, errors.New("The floor is too thick for the height. Make the floor thinner or the tray taller.")
	}
	base, err := outline.FootprintRounded(v["length"], v["width"], v["roundness"], int(v["sides"]), v["corner"], points)
	if errors.Is(err, outline.ErrCornerTooBig) {
		return family.Shape{}, errors.New("The corner radius is too big for this tray. Make it smaller or the tray bigger.")
	}
	if err != nil {
		return family.Shape{}, err
	}
	flare := v["flare"] * math.Pi / 180
	at := func(z float64) geom.Outline {
		o, _ := outline.Inset(base, -z*math.Tan(flare)) // outward offsets never collapse
		return o
	}
	innerBottom, err := outline.Inset(at(floor), wall)
	// The rim's outline contains the floor's (the walls only flare out), so
	// if the floor's inset works, the rim's does too.
	innerTop, _ := outline.Inset(at(h), wall)
	if err != nil {
		return family.Shape{}, errors.New("The wall is too thick for this tray. Make the wall thinner or the tray bigger.")
	}
	m := &geom.Mesh{}
	o0, o1 := m.Ring(base, 0), m.Ring(at(h), h)
	i0, i1 := m.Ring(innerBottom, floor), m.Ring(innerTop, h)
	m.Band(o0, o1, false)
	m.Band(i0, i1, true)
	m.Band(o1, i1, false) // rim
	m.Fan(m.Add(geom.Vec3{}), o0, true)
	m.Fan(m.Add(geom.Vec3{Z: floor}), i0, false)
	return family.Shape{Mesh: m, MinWall: min(floor, wall*math.Cos(flare)), Overhang: v["flare"]}, nil
}
