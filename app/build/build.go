// Package build is the use case that turns a family and settings into a
// checked shape, and writes it in a chosen format.
package build

import (
	"fmt"
	"io"

	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/param"
)

// Exporter writes shapes in one file format (a port; adapters implement it).
type Exporter interface {
	ID() string
	Label() string
	Extension() string
	MediaType() string
	Supports(family.Shape) bool
	Write(io.Writer, family.Shape) error
}

// Format describes a file format on offer.
type Format struct{ ID, Label, Extension, MediaType string }

// Result is a built, checked shape.
type Result struct {
	Family   string
	Values   param.Values
	Notes    []string
	Shape    family.Shape
	Findings []check.Finding
	Formats  []Format
}

// Builder builds and exports shapes.
type Builder struct {
	families  *family.Registry
	checks    []check.Check
	exporters []Exporter
}

// New prepares a builder; exporters are offered in the order given.
func New(r *family.Registry, checks []check.Check, exporters ...Exporter) (*Builder, error) {
	seen := map[string]bool{}
	for _, e := range exporters {
		if seen[e.ID()] {
			return nil, fmt.Errorf("format %q is registered twice", e.ID())
		}
		seen[e.ID()] = true
	}
	return &Builder{families: r, checks: checks, exporters: exporters}, nil
}

// Build builds a family's shape from its defaults with the given changes,
// runs every check, and lists the formats that can hold it (none if a check
// found an error).
func (b *Builder) Build(familyID string, values param.Values, ctx check.Context) (Result, error) {
	f, ok := b.families.Get(familyID)
	if !ok {
		return Result{}, fmt.Errorf("unknown family %q", familyID)
	}
	ps := f.Params()
	vals, notes, err := ps.Apply(ps.Defaults(), values)
	if err != nil {
		return Result{}, err
	}
	shape, err := f.Build(vals)
	if err != nil {
		return Result{}, err
	}
	r := Result{Family: familyID, Values: vals, Notes: notes, Shape: shape}
	for _, c := range b.checks {
		r.Findings = append(r.Findings, c.Run(shape, ctx)...)
	}
	if check.Blocking(r.Findings) {
		return r, nil
	}
	for _, e := range b.exporters {
		if e.Supports(shape) {
			r.Formats = append(r.Formats, describe(e))
		}
	}
	return r, nil
}

// Export builds the shape and writes it in the chosen format. Nothing is
// written unless the shape passed every error-level check.
func (b *Builder) Export(familyID string, values param.Values, ctx check.Context, format string, w io.Writer) (Format, error) {
	r, err := b.Build(familyID, values, ctx)
	if err != nil {
		return Format{}, err
	}
	for _, f := range r.Findings {
		if f.Level == check.Error {
			return Format{}, fmt.Errorf("This shape can't be downloaded: %s", f.Message)
		}
	}
	for _, e := range b.exporters {
		if e.ID() != format {
			continue
		}
		if !e.Supports(r.Shape) {
			return Format{}, fmt.Errorf("%s isn't available for this shape.", e.Label())
		}
		return describe(e), e.Write(w, r.Shape)
	}
	return Format{}, fmt.Errorf("unknown format %q", format)
}

func describe(e Exporter) Format {
	return Format{ID: e.ID(), Label: e.Label(), Extension: e.Extension(), MediaType: e.MediaType()}
}
