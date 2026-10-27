package api

import (
	"encoding/binary"
	"math"

	"github.com/kinglet-dev/saymake/app/build"
	"github.com/kinglet-dev/saymake/app/interpret"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
)

type reading struct {
	Concept string   `json:"concept"`
	Family  string   `json:"family"`
	Path    []string `json:"path"`
	Uses    []use    `json:"uses"`
	Unknown []string `json:"unknown"`
	Notes   []string `json:"notes"`
}

type use struct {
	Typed   string `json:"typed"`
	Meaning string `json:"meaning"`
	Kind    string `json:"kind"`
}

type paramAnswer struct {
	Name  string  `json:"name"`
	Label string  `json:"label"`
	Unit  string  `json:"unit"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Step  float64 `json:"step"`
	Whole bool    `json:"whole"`
	Value float64 `json:"value"`
}

type finding struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type format struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Extension string `json:"extension"`
}

type shape struct {
	Family      string        `json:"family"`
	FamilyLabel string        `json:"familyLabel"`
	Params      []paramAnswer `json:"params"`
	Notes       []string      `json:"notes"`
	Findings    []finding     `json:"findings"`
	Formats     []format      `json:"formats"`
	Triangles   int           `json:"triangles"`
	Outlines    [][]float64   `json:"outlines"`
	Scores      [][]float64   `json:"scores"`
	Size        [3]float64    `json:"size"` // extent in x, y, z, mm
}

func toReading(in interpret.Interpretation) *reading {
	r := &reading{Concept: in.Concept, Family: in.Family, Path: in.Path, Unknown: in.Unknown, Notes: in.Notes}
	for _, u := range in.Uses {
		r.Uses = append(r.Uses, use{Typed: u.Typed, Meaning: u.Meaning, Kind: string(u.Kind)})
	}
	return r
}

// toShape describes a built shape and encodes its mesh: little-endian uint32
// vertex count, float32 x, y, z per vertex, uint32 triangle count, uint32
// corners per triangle (4-byte aligned for typed arrays).
func toShape(f family.Family, res build.Result) (*shape, []byte) {
	s := &shape{Family: f.ID(), FamilyLabel: f.Label(), Notes: res.Notes}
	for _, p := range f.Params() {
		s.Params = append(s.Params, paramAnswer{p.Name, p.Label, p.Unit, p.Min, p.Max, p.Step, p.Whole, res.Values[p.Name]})
	}
	for _, x := range res.Findings {
		s.Findings = append(s.Findings, finding{string(x.Level), x.Message})
	}
	for _, x := range res.Formats {
		s.Formats = append(s.Formats, format{x.ID, x.Label, x.Extension})
	}
	lo, hi := geom.Vec3{X: math.Inf(1), Y: math.Inf(1), Z: math.Inf(1)}, geom.Vec3{X: math.Inf(-1), Y: math.Inf(-1), Z: math.Inf(-1)}
	grow := func(v geom.Vec3) {
		lo = geom.Vec3{X: min(lo.X, v.X), Y: min(lo.Y, v.Y), Z: min(lo.Z, v.Z)}
		hi = geom.Vec3{X: max(hi.X, v.X), Y: max(hi.Y, v.Y), Z: max(hi.Z, v.Z)}
	}
	s.Outlines, s.Scores = flatten(res.Shape.Outlines, grow), flatten(res.Shape.Scores, grow)
	var mesh []byte
	if m := res.Shape.Mesh; m != nil {
		s.Triangles = len(m.Triangles)
		mesh = encode(m)
		for _, v := range m.Vertices {
			grow(v)
		}
	}
	if lo.X <= hi.X { // nothing was drawn when a build failed
		s.Size = [3]float64{hi.X - lo.X, hi.Y - lo.Y, hi.Z - lo.Z}
	}
	return s, mesh
}

func flatten(os []geom.Outline, grow func(geom.Vec3)) [][]float64 {
	var out [][]float64
	for _, o := range os {
		var xs []float64
		for _, p := range o {
			xs = append(xs, p.X, p.Y)
			grow(geom.Vec3{X: p.X, Y: p.Y})
		}
		out = append(out, xs)
	}
	return out
}

func encode(m *geom.Mesh) []byte {
	b := make([]byte, 0, 8+12*len(m.Vertices)+12*len(m.Triangles))
	b = binary.LittleEndian.AppendUint32(b, uint32(len(m.Vertices)))
	for _, v := range m.Vertices {
		for _, c := range []float64{v.X, v.Y, v.Z} {
			b = binary.LittleEndian.AppendUint32(b, math.Float32bits(float32(c)))
		}
	}
	b = binary.LittleEndian.AppendUint32(b, uint32(len(m.Triangles)))
	for _, t := range m.Triangles {
		for _, i := range t {
			b = binary.LittleEndian.AppendUint32(b, uint32(i))
		}
	}
	return b
}
