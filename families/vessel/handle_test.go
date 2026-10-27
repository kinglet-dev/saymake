package vessel_test

import (
	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vessel handles", func() {
	It("adds a closed C-shaped handle on the side, its ends inside the wall", func() {
		// Arrange
		v := cylinder()
		v["handle"], v["handleReach"], v["handleThick"] = 0.5, 25, 8

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		_, without := build(cylinder())
		Expect(r.Volume-without.Volume).To(BeNumerically(">", 0), "the handle faces outward (a handle turned inside out would take volume away)")
		Expect(r.Max.X).To(BeNumerically("~", 39+25+4, 0.05), "it reaches out from the middle of the wall")
		Expect(r.Max.Z).To(BeNumerically("~", 100, 1e-9))
		Expect(s.Overhang).To(BeNumerically(">=", 90), "a handle's underside overhangs")
	})

	It("sizes the handle as a share of the height", func() {
		// Arrange: half of a 60 mm cup
		v := cylinder()
		v["height"], v["handle"], v["handleThick"] = 60, 0.5, 6

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		top := 0.0
		for _, p := range s.Mesh.Vertices {
			if p.X > 41 {
				top = max(top, p.Z)
			}
		}
		Expect(top).To(BeNumerically("~", 30+15+3, 0.5))
	})

	It("adds nothing without a handle", func() {
		// Arrange
		v := cylinder()

		// Act
		_, r := build(v)

		// Assert
		Expect(r.Max.X).To(BeNumerically("~", 40, 1e-9))
	})

	It("stays sound with every feature at once", func() {
		// Arrange
		v := cylinder()
		for k, x := range map[string]float64{"handle": 0.4, "flutes": 12, "twist": 60, "sides": 0, "rimStyle": 2, "rim": 1,
			"foot": 2, "floor": 4, "drain": 8, "base": 0.8, "top": 1.1, "bulge": 0.2} {
			v[k] = x
		}

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(len(s.Mesh.Triangles)).To(BeNumerically("<=", vessel.MaxTriangles))
	})

	It("accepts a handle that reaches exactly 1 mm from the top and the bottom", func() {
		// Arrange: 0.9 of 100 mm, 8 mm thick, ends 1 mm in from both
		v := cylinder()
		v["handle"], v["handleThick"] = 0.9, 8

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
	})

	It("refuses a handle too tall for the vessel", func() {
		// Arrange
		v := cylinder()
		v["handle"], v["handleThick"] = 0.9, 10

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).To(MatchError("The handle is too tall for this vessel. Make it shorter or the vessel taller."))
	})
})
