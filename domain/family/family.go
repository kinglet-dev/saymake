// Package family defines the plugin contract every shape family implements,
// and the registry that holds them (Open/Closed: a new family is a new
// implementation, registered in the composition root).
package family

import (
	"errors"
	"fmt"

	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/param"
)

// Shape is what a family builds: a solid, flat outlines, or both, plus the
// measurements the printing checks need.
type Shape struct {
	Mesh     *geom.Mesh     // a closed solid, if the shape is 3D
	Outlines []geom.Outline // cut lines in the XY plane, if the shape is flat
	Scores   []geom.Outline // lines scored or engraved, not cut
	MinWall  float64        // thinnest wall or floor, in mm (0: not measured)
	Overhang float64        // steepest outward overhang, in degrees from vertical
	Bridge   float64        // widest flat ceiling facing down, in mm (0: none)
}

// Family is a parametric shape family.
type Family interface {
	ID() string
	Label() string
	Params() param.Set
	MainSize() string // the setting a bare size ("20 cm") goes to
	// Proportional lists settings scaled along with the main size when a
	// bare size is given, so "a 30 cm tray" is the same tray, bigger.
	Proportional() []string
	Build(param.Values) (Shape, error)
}

// Registry holds the families by id, in registration order.
type Registry struct {
	byID  map[string]Family
	order []string
}

// NewRegistry checks each family's contract and registers it.
func NewRegistry(fs ...Family) (*Registry, error) {
	r := &Registry{byID: map[string]Family{}}
	for _, f := range fs {
		id := f.ID()
		switch _, dup := r.byID[id]; {
		case id == "":
			return nil, errors.New("a family has no id")
		case dup:
			return nil, fmt.Errorf("family %q is registered twice", id)
		}
		if err := f.Params().Validate(); err != nil {
			return nil, fmt.Errorf("family %q: %w", id, err)
		}
		main, ok := f.Params().Find(f.MainSize())
		switch {
		case !ok:
			return nil, fmt.Errorf("family %q: its main size %q is not one of its settings", id, f.MainSize())
		case main.Min <= 0:
			return nil, fmt.Errorf("family %q: its main size %q must have a positive minimum", id, main.Name)
		}
		for _, name := range f.Proportional() {
			if _, ok := f.Params().Find(name); !ok {
				return nil, fmt.Errorf("family %q: %q scales with its main size but is not one of its settings", id, name)
			}
		}
		r.byID[id] = f
		r.order = append(r.order, id)
	}
	return r, nil
}

// Get returns the family with the given id.
func (r *Registry) Get(id string) (Family, bool) {
	f, ok := r.byID[id]
	return f, ok
}

// IDs returns the registered ids in registration order.
func (r *Registry) IDs() []string { return append([]string(nil), r.order...) }
