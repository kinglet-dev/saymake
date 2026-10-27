package vessel_test

import (
	"math"

	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vessel family: patterns and resolution", func() {
	It("has the corners of a regular polygon when given sides", func() {
		// Arrange
		v := cylinder()
		v["sides"] = 6

		// Act
		_, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		// The inner wall keeps the wall thickness across each flat side.
		want := polygonArea(6, 40)*100 - polygonArea(6, 40-2/math.Cos(math.Pi/6))*97
		Expect(r.Volume).To(BeNumerically("~", want, want*1e-9))
	})

	It("reads fewer than 3 sides as round", func() {
		// Arrange
		v := cylinder()
		v["sides"] = 2

		// Act
		_, r := build(v)

		// Assert
		n := 192
		want := polygonArea(n, 40)*100 - polygonArea(n, 38)*97
		Expect(r.Volume).To(BeNumerically("~", want, want*1e-9))
	})

	It("ripples the radius by the flute depth", func() {
		// Arrange
		v := cylinder()
		v["flutes"], v["fluteDepth"] = 8, 0.1

		// Act
		_, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(r.Max.X).To(BeNumerically("~", 44, 1e-6))
	})

	It("stays sound when twisted, with sides and flutes together", func() {
		// Arrange
		v := cylinder()
		v["twist"], v["sides"], v["flutes"] = 120, 5, 20

		// Act
		_, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
	})

	It("bulges at the belly and pinches at the neck", func() {
		// Arrange
		v := cylinder()
		v["bulge"], v["belly"], v["neck"], v["neckAt"] = 0.3, 0.4, 0.4, 0.8

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(r.Max.X).To(BeNumerically("~", 52, 0.05))
		Expect(s.Overhang).To(BeNumerically(">", 0))
		top := 0.0
		for _, p := range s.Mesh.Vertices {
			if p.Z == 100 {
				top = max(top, p.X)
			}
		}
		Expect(top).To(BeNumerically("~", 40, 1e-9), "the bulge has died away by the rim")
	})

	It("cuts a drainage hole through the floor", func() {
		// Arrange
		v := cylinder()
		v["drain"] = 10

		// Act
		_, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		n := 192
		want := polygonArea(n, 40)*100 - polygonArea(n, 38)*97 - polygonArea(n, 5)*3
		Expect(r.Volume).To(BeNumerically("~", want, want*1e-9))
	})

	It("uses eight points per flute when that's finer", func() {
		// Arrange
		v := cylinder()
		v["flutes"] = 32

		// Act
		s, _ := build(v)

		// Assert
		bottom := 0
		for _, p := range s.Mesh.Vertices {
			if p.Z == 0 && math.Hypot(p.X, p.Y) > 30 {
				bottom++
			}
		}
		Expect(bottom).To(Equal(256))
	})

	It("has the corners of a triangle with 3 sides", func() {
		// Arrange
		v := cylinder()
		v["sides"] = 3

		// Act
		_, r := build(v)

		// Assert
		want := polygonArea(3, 40)*100 - polygonArea(3, 40-2/math.Cos(math.Pi/3))*97
		Expect(r.Volume).To(BeNumerically("~", want, want*1e-9))
	})

	It("accepts a floor 1 mm under the rim", func() {
		// Arrange
		v := cylinder()
		v["height"], v["floor"] = 20, 19

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
	})

	It("doesn't measure a vessel without flutes as if it had them", func() {
		// Arrange: inside, 1.03 mm from the axis; flutes would dip that under 1 mm
		v := cylinder()
		v["diameter"], v["wall"], v["fluteDepth"] = 20, 8.97, 0.06

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
	})

	It("stays within its triangle budget at its finest", func() {
		// Arrange
		v := cylinder()
		v["flutes"], v["sides"], v["drain"] = 32, 12, 10

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(len(s.Mesh.Triangles)).To(BeNumerically("<=", vessel.MaxTriangles))
	})
})
