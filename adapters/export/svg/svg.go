// Package svg writes flat shapes as SVG 1.1 for laser software: millimeter
// width and height with a matching viewBox (one user unit is 1 mm), cut lines
// in a red group and score lines in a blue one, unfilled. Laser programs
// such as LightBurn assign layers by color. Only numbers from the shape are
// written, so the file can carry no script.
package svg

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strconv"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
)

// Exporter writes SVG.
type Exporter struct{}

// ID is the format's id.
func (Exporter) ID() string { return "svg" }

// Label is the format's name.
func (Exporter) Label() string { return "SVG" }

// Extension is the file extension.
func (Exporter) Extension() string { return ".svg" }

// MediaType is the file's media type.
func (Exporter) MediaType() string { return "image/svg+xml" }

// Supports reports whether the shape can be written.
func (Exporter) Supports(s family.Shape) bool { return len(s.Outlines) > 0 }

// Write writes the shape's outlines.
func (Exporter) Write(w io.Writer, s family.Shape) error {
	if len(s.Outlines) == 0 {
		return errors.New("SVG needs a flat shape")
	}
	lo, hi := s.Outlines[0].Bounds()
	for _, o := range append(append([]geom.Outline{}, s.Outlines...), s.Scores...) {
		l, h := o.Bounds()
		lo = geom.Vec2{X: min(lo.X, l.X), Y: min(lo.Y, l.Y)}
		hi = geom.Vec2{X: max(hi.X, h.X), Y: max(hi.Y, h.Y)}
	}
	width, height := num(hi.X-lo.X), num(hi.Y-lo.Y)
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" version="1.1" width="` + width + `mm" height="` + height +
		`mm" viewBox="` + num(lo.X) + " " + num(-hi.Y) + " " + width + " " + height + `">` + "\n")
	group(&b, "cut", "#ff0000", s.Outlines)
	group(&b, "score", "#0000ff", s.Scores)
	b.WriteString("</svg>\n")
	_, err := w.Write(b.Bytes())
	return err
}

func group(b *bytes.Buffer, id, color string, os []geom.Outline) {
	if len(os) == 0 {
		return
	}
	b.WriteString(`<g id="` + id + `" fill="none" stroke="` + color + `" stroke-width="0.1">` + "\n")
	for _, o := range os {
		b.WriteString(`<path d="`)
		for i, p := range o {
			cmd := "L"
			if i == 0 {
				cmd = "M"
			}
			b.WriteString(cmd + num(p.X) + " " + num(-p.Y))
		}
		b.WriteString(`Z"/>` + "\n")
	}
	b.WriteString("</g>\n")
}

// num writes a coordinate rounded to 0.001 mm, without needless digits.
func num(f float64) string {
	r := math.Round(f*1000) / 1000
	if r == 0 {
		r = 0 // no "-0"
	}
	return strconv.FormatFloat(r, 'f', -1, 64)
}
