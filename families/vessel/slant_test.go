package vessel_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// highestWhere returns the highest z among vertices that pass keep.
func highestWhere(vs []geom.Vec3, keep func(geom.Vec3) bool) float64 {
	h := math.Inf(-1)
	for _, v := range vs {
		if keep(v) {
			h = max(h, v.Z)
		}
	}
	return h
}

var _ = Describe("A slanted rim", func() {
	It("cuts the top at the angle, high on the handle's side and low opposite", func() {
		// Arrange
		v := cylinder()
		v["slant"] = 20
		drop := 2 * 40 * math.Tan(20*math.Pi/180)

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(r.Max.Z).To(BeNumerically("~", 100, 1e-9))
		Expect(highestWhere(s.Mesh.Vertices, func(p geom.Vec3) bool { return p.X > 39.9 })).To(BeNumerically("~", 100, 1e-9))
		Expect(highestWhere(s.Mesh.Vertices, func(p geom.Vec3) bool { return p.X < -39.99 })).To(BeNumerically("~", 100-drop, 1e-9))
		Expect(highestWhere(s.Mesh.Vertices, func(p geom.Vec3) bool { return p.Y > 39.99 })).To(BeNumerically("~", 100-drop/2, 1e-9))
	})

	It("holds less than the same vessel cut flat", func() {
		// Arrange
		flat, slanted := cylinder(), cylinder()
		slanted["slant"] = 20

		// Act
		_, rf := build(flat)
		_, rs := build(slanted)

		// Assert
		Expect(rs.Volume).To(BeNumerically("<", rf.Volume))
	})

	DescribeTable("builds a sound solid with every rim style",
		func(style float64) {
			// Arrange
			v := cylinder()
			v["slant"], v["rimStyle"], v["rim"] = 30, style, 1

			// Act
			_, r := build(v)

			// Assert
			Expect(r.Sound()).To(Succeed())
		},
		Entry("flat", 0.0),
		Entry("chamfered", 1.0),
		Entry("rounded", 2.0),
	)

	DescribeTable("needs the low side to stay over 1 mm above the floor",
		func(height float64, ok bool) {
			// Arrange
			v := cylinder()
			v["slant"], v["height"] = 45, height

			// Act
			_, err := vessel.New().Build(v)

			// Assert
			if ok {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(err).To(MatchError("The rim slants too steeply for this vessel: its low side would reach the floor. Make the slant smaller or the vessel taller."))
		},
		Entry("low side 5 mm up", 85.0, true),
		Entry("low side 3 mm up", 83.0, false),
	)

	It("checks the rim's edge on the low side too, where the profile differs", func() {
		// Arrange
		v := cylinder()
		v["neck"], v["neckAt"], v["rimStyle"], v["rim"] = 0.8, 0.6, 1, 1.5
		_, flatErr := vessel.New().Build(v)
		v["slant"] = 20

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(flatErr).NotTo(HaveOccurred())
		Expect(err).To(MatchError("The rim's edge is as wide as the wall. Make the rim smaller or the wall thicker."))
	})
})
