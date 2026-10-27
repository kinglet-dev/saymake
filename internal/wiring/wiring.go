// Package wiring is part of the composition root: it lists the built-in
// families, vocabulary, checks and exporters, and assembles the tool from
// them. cmd/ calls it, and the acceptance tests use it to run the real
// assembly.
package wiring

import (
	"github.com/kinglet-dev/saymake/adapters/api"
	"github.com/kinglet-dev/saymake/adapters/export/stl"
	"github.com/kinglet-dev/saymake/adapters/export/svg"
	"github.com/kinglet-dev/saymake/adapters/export/threemf"
	"github.com/kinglet-dev/saymake/adapters/vocabjson"
	"github.com/kinglet-dev/saymake/app/build"
	"github.com/kinglet-dev/saymake/app/interpret"
	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/vocab"
	"github.com/kinglet-dev/saymake/families/flat"
	"github.com/kinglet-dev/saymake/families/tray"
	"github.com/kinglet-dev/saymake/families/vessel"
)

// Printing limits (open decision 1 in DESIGN.md).
const (
	minWall     = 0.8 // mm: two lines of a 0.4 mm nozzle
	maxOverhang = 45  // degrees from upright
	maxBridge   = 30  // mm of flat ceiling (provisional; see DESIGN.md)
)

// Parts are what the tool is assembled from.
type Parts struct {
	Families   []family.Family
	Vocabulary func() (vocab.Vocabulary, error)
	Checks     []check.Check
	Exporters  []build.Exporter
}

// BuiltIn returns the built-in parts. A new family, check or format is
// added here (Open/Closed: nothing else changes).
func BuiltIn() Parts {
	return Parts{
		Families:   []family.Family{vessel.New(), tray.New(), flat.New()},
		Vocabulary: vocabjson.Default,
		Checks:     []check.Check{check.Solid{}, check.Wall{Min: minWall}, check.Overhang{Max: maxOverhang}, check.Bridge{Max: maxBridge}, check.BuildVolume{}, check.Stable{}},
		Exporters:  []build.Exporter{stl.Exporter{}, threemf.Exporter{}, svg.Exporter{}},
	}
}

// App is the assembled tool.
type App struct {
	Families    *family.Registry
	Interpreter *interpret.Interpreter
	Builder     *build.Builder
}

// New assembles the tool from the built-in parts.
func New() (App, error) { return Assemble(BuiltIn()) }

// Assemble checks the parts fit together and assembles the tool.
func Assemble(p Parts) (App, error) {
	fams, err := family.NewRegistry(p.Families...)
	if err != nil {
		return App{}, err
	}
	v, err := p.Vocabulary()
	if err != nil {
		return App{}, err
	}
	in, err := interpret.New(v, fams)
	if err != nil {
		return App{}, err
	}
	b, err := build.New(fams, p.Checks, p.Exporters...)
	if err != nil {
		return App{}, err
	}
	return App{Families: fams, Interpreter: in, Builder: b}, nil
}

// Handler assembles the tool and returns the page's request handler; if the
// parts don't fit, the handler answers every request with the reason.
func Handler(p Parts) *api.Handler {
	app, err := Assemble(p)
	if err != nil {
		return api.Unavailable(err)
	}
	return api.New(app.Interpreter, app.Builder, app.Families)
}
