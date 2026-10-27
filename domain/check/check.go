// Package check holds the checks a shape must pass before it is offered for
// download (errors) or that the user should know about (warnings).
package check

import (
	"fmt"
	"math"
	"slices"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/meshcheck"
)

// Level is how serious a finding is.
type Level string

// The levels.
const (
	Error   Level = "error"
	Warning Level = "warning"
)

// Finding is one thing a check found.
type Finding struct {
	Level   Level
	Message string
}

// Context is what a check may need besides the shape.
type Context struct {
	Printer [3]float64 // build volume x, y, z in mm; zero: not given
}

// Check examines a shape (Strategy; new checks are registered in the
// composition root).
type Check interface {
	Run(family.Shape, Context) []Finding
}

// Solid requires a closed, outward mesh and real outlines.
type Solid struct{}

// Run implements Check.
func (Solid) Run(s family.Shape, _ Context) []Finding {
	var fs []Finding
	if s.Mesh != nil {
		if err := meshcheck.Inspect(*s.Mesh).Sound(); err != nil {
			fs = append(fs, Finding{Error, fmt.Sprintf("The shape didn't come out as a closed solid (%v). Please report this with your description.", err)})
		}
	}
	if slices.ContainsFunc(append(slices.Clone(s.Outlines), s.Scores...), broken) {
		fs = append(fs, Finding{Error, "An outline didn't come out right. Please report this with your description."})
	}
	return fs
}

func broken(o geom.Outline) bool {
	return !o.Finite() || math.Abs(o.Area()) < 1e-9 // fewer than 3 points have no area
}

// Wall warns about walls thinner than Min millimeters.
type Wall struct{ Min float64 }

// Run implements Check.
func (w Wall) Run(s family.Shape, _ Context) []Finding {
	if s.MinWall <= 0 || s.MinWall >= w.Min {
		return nil
	}
	return []Finding{{Warning, fmt.Sprintf("The thinnest wall is about %.1f mm (estimated), under %g mm. Thicker walls print more reliably.", s.MinWall, w.Min)}}
}

// Overhang warns about walls leaning more than Max degrees from upright.
type Overhang struct{ Max float64 }

// Run implements Check.
func (o Overhang) Run(s family.Shape, _ Context) []Finding {
	if s.Overhang <= o.Max {
		return nil
	}
	return []Finding{{Warning, fmt.Sprintf("Parts of the shape lean about %.0f° from upright, more than %g°. They may need supports to print.", s.Overhang, o.Max)}}
}

// Blocking reports whether any finding is an error.
func Blocking(fs []Finding) bool {
	return slices.ContainsFunc(fs, func(f Finding) bool { return f.Level == Error })
}

// Bridge warns about flat ceilings facing down wider than Max millimeters.
type Bridge struct{ Max float64 }

// Run implements Check.
func (b Bridge) Run(s family.Shape, _ Context) []Finding {
	if s.Bridge <= b.Max {
		return nil
	}
	return []Finding{{Warning, fmt.Sprintf("A flat ceiling about %.0f mm across faces down (under the raised foot); printers bridge more than %g mm poorly, so it may need supports.", s.Bridge, b.Max)}}
}
