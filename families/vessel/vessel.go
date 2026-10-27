// Package vessel is the revolved-vessel family: vases, bowls, cups and
// planters, as a surface of revolution with an optional pattern around it.
package vessel

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/param"
)

// MaxTriangles bounds the mesh size (see the threat model).
const MaxTriangles = 270_000

// Vessel is the family.
type Vessel struct{}

// New returns the family.
func New() Vessel { return Vessel{} }

// ID implements family.Family.
func (Vessel) ID() string { return "vessel" }

// Label implements family.Family.
func (Vessel) Label() string { return "Vessel" }

// MainSize implements family.Family.
func (Vessel) MainSize() string { return "height" }

// Proportional implements family.Family.
func (Vessel) Proportional() []string { return []string{"diameter"} }

// Params implements family.Family.
func (Vessel) Params() param.Set { return append(param.Set(nil), specs...) }

// Build implements family.Family.
func (Vessel) Build(v param.Values) (family.Shape, error) {
	c, err := read(v)
	if err != nil {
		return family.Shape{}, err
	}
	if err := c.check(); err != nil {
		return family.Shape{}, err
	}
	steepest := c.leaning(c.overhang(rings))
	thinnest := c.wall * math.Cos(steepest*math.Pi/180)
	if c.flutes > 0 {
		thinnest *= 1 - c.fluteDepth
	}
	if c.rimStyle > 0 {
		thinnest = min(thinnest, c.wall-c.rim)
	}
	s := family.Shape{Mesh: c.mesh(), MinWall: min(thinnest, c.floor), Overhang: steepest}
	if c.foot > 0 {
		s.Bridge = 2 * c.footRadius()
	}
	if c.handle > 0 {
		s.Overhang = max(s.Overhang, 90) // the handle's underside faces down
	}
	return s, nil
}
