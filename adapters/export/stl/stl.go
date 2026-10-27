// Package stl writes binary STL (3D Systems' StereoLithography Interface
// Specification, 1988): an 80-byte header, a triangle count, and for each
// triangle its normal, three corners and a zero attribute, little-endian.
package stl

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
)

// header must not start with "solid": readers take that as text STL.
const header = "Binary STL from Kinglet Saymake, in millimeters"

// Exporter writes STL.
type Exporter struct{}

// ID is the format's id.
func (Exporter) ID() string { return "stl" }

// Label is the format's name.
func (Exporter) Label() string { return "STL" }

// Extension is the file extension.
func (Exporter) Extension() string { return ".stl" }

// MediaType is the file's media type.
func (Exporter) MediaType() string { return "model/stl" }

// Supports reports whether the shape can be written.
func (Exporter) Supports(s family.Shape) bool { return s.Mesh != nil }

// Write writes the shape's solid.
func (Exporter) Write(w io.Writer, s family.Shape) error {
	if s.Mesh == nil {
		return errors.New("STL needs a solid shape")
	}
	m := s.Mesh
	bw := bufio.NewWriter(w)
	var head [80]byte
	copy(head[:], header)
	bw.Write(head[:])
	binary.Write(bw, binary.LittleEndian, uint32(len(m.Triangles)))
	rec := make([]float32, 12)
	for _, t := range m.Triangles {
		a, b, c := m.Vertices[t[0]], m.Vertices[t[1]], m.Vertices[t[2]]
		n := unit(b.Sub(a).Cross(c.Sub(a)))
		for i, v := range []geom.Vec3{n, a, b, c} {
			rec[3*i], rec[3*i+1], rec[3*i+2] = float32(v.X), float32(v.Y), float32(v.Z)
		}
		binary.Write(bw, binary.LittleEndian, rec)
		binary.Write(bw, binary.LittleEndian, uint16(0))
	}
	return bw.Flush() // bufio keeps the first write error and returns it here
}

func unit(v geom.Vec3) geom.Vec3 {
	l := math.Sqrt(v.Dot(v))
	return geom.Vec3{X: v.X / l, Y: v.Y / l, Z: v.Z / l}
}
