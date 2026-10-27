package stl_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"

	"github.com/kinglet-dev/saymake/adapters/export/stl"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func tetra() *geom.Mesh {
	m := &geom.Mesh{}
	a, b := m.Add(geom.Vec3{}), m.Add(geom.Vec3{X: 10})
	c, d := m.Add(geom.Vec3{Y: 10}), m.Add(geom.Vec3{Z: 10})
	m.Tri(a, c, b) // bottom, facing −z
	m.Tri(a, b, d)
	m.Tri(b, c, d)
	m.Tri(c, a, d)
	return m
}

type record struct {
	Normal, A, B, C [3]float32
	Attr            uint16
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

var _ = Describe("STL exporter", func() {
	It("describes itself", func() {
		// Arrange
		e := stl.Exporter{}

		// Act
		got := []string{e.ID(), e.Label(), e.Extension(), e.MediaType()}

		// Assert
		Expect(got).To(Equal([]string{"stl", "STL", ".stl", "model/stl"}))
	})

	It("supports shapes with a solid only", func() {
		// Arrange
		e := stl.Exporter{}

		// Act
		solid, flat := e.Supports(family.Shape{Mesh: tetra()}), e.Supports(family.Shape{})

		// Assert
		Expect([]bool{solid, flat}).To(Equal([]bool{true, false}))
	})

	It("writes binary STL: a header that isn't mistaken for text, the count, and each triangle with its outward normal", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		err := stl.Exporter{}.Write(&buf, family.Shape{Mesh: tetra()})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		b := buf.Bytes()
		Expect(b).To(HaveLen(80 + 4 + 4*50))
		Expect(string(b[:5])).NotTo(Equal("solid"), "a binary header starting with \"solid\" makes some readers parse it as text STL")
		Expect(string(b[:80])).To(HavePrefix("Binary STL from Kinglet Saymake, in millimeters"))
		Expect(binary.LittleEndian.Uint32(b[80:84])).To(Equal(uint32(4)))
		var r record
		Expect(binary.Read(bytes.NewReader(b[84:134]), binary.LittleEndian, &r)).To(Succeed())
		Expect(r.Normal).To(Equal([3]float32{0, 0, -1}))
		Expect(r.A).To(Equal([3]float32{0, 0, 0}))
		Expect(r.B).To(Equal([3]float32{0, 10, 0}))
		Expect(r.C).To(Equal([3]float32{10, 0, 0}))
		Expect(r.Attr).To(BeZero())
	})

	It("writes a unit normal for a slanted face", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		err := stl.Exporter{}.Write(&buf, family.Shape{Mesh: tetra()})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		var r record
		Expect(binary.Read(bytes.NewReader(buf.Bytes()[84+2*50:84+3*50]), binary.LittleEndian, &r)).To(Succeed())
		k := float32(1 / math.Sqrt(3))
		for _, c := range r.Normal {
			Expect(c).To(BeNumerically("~", k, 1e-6))
		}
	})

	It("refuses a shape without a solid", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		err := stl.Exporter{}.Write(&buf, family.Shape{})

		// Assert
		Expect(err).To(MatchError("STL needs a solid shape"))
	})

	It("passes on a write error", func() {
		// Arrange
		w := failingWriter{}

		// Act
		err := stl.Exporter{}.Write(w, family.Shape{Mesh: tetra()})

		// Assert
		Expect(err).To(MatchError("disk full"))
	})
})
