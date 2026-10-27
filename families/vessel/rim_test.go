package vessel_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// widestAt returns the largest distance from the axis among vertices at z.
func widestAt(vs []geom.Vec3, z float64) float64 {
	w := 0.0
	for _, v := range vs {
		if math.Abs(v.Z-z) < 1e-9 {
			w = max(w, math.Hypot(v.X, v.Y))
		}
	}
	return w
}

var _ = Describe("Vessel rims", func() {
	It("cuts a 45° chamfer into the outer top edge", func() {
		// Arrange
		v := cylinder()
		v["rimStyle"], v["rim"] = 1, 1

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(widestAt(s.Mesh.Vertices, 100)).To(BeNumerically("~", 39, 1e-9))
		Expect(widestAt(s.Mesh.Vertices, 99)).To(BeNumerically("~", 40, 1e-9))
		Expect(s.MinWall).To(BeNumerically("~", 1, 1e-9), "the wall is thinnest at the top of the chamfer")
	})

	It("rounds the outer top edge with a quarter circle", func() {
		// Arrange
		v := cylinder()
		v["rimStyle"], v["rim"] = 2, 1.5

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(widestAt(s.Mesh.Vertices, 100)).To(BeNumerically("~", 38.5, 1e-9))
		between := 0
		for _, p := range s.Mesh.Vertices {
			if p.Z > 98.5+1e-9 && p.Z < 100-1e-9 {
				between++
			}
		}
		Expect(between).To(BeNumerically(">", 0), "points along the arc")
	})

	It("keeps a flat rim when the style is flat, whatever the rim size", func() {
		// Arrange
		v := cylinder()
		v["rimStyle"], v["rim"] = 0, 1.5

		// Act
		s, _ := build(v)

		// Assert
		Expect(widestAt(s.Mesh.Vertices, 100)).To(BeNumerically("~", 40, 1e-9))
	})

	It("refuses a rim as wide as the wall", func() {
		// Arrange
		v := cylinder()
		v["rimStyle"], v["rim"], v["wall"] = 1, 2, 2

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).To(MatchError("The rim's edge is as wide as the wall. Make the rim smaller or the wall thicker."))
	})
})
