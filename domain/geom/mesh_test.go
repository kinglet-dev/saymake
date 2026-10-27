package geom_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mesh", func() {
	It("numbers vertices in the order they are added", func() {
		// Arrange
		var m geom.Mesh

		// Act
		a := m.Add(geom.Vec3{X: 1})
		b := m.Add(geom.Vec3{Y: 1})

		// Assert
		Expect([]int{a, b}).To(Equal([]int{0, 1}))
		Expect(m.Vertices).To(Equal([]geom.Vec3{{X: 1}, {Y: 1}}))
	})

	It("adds a triangle by vertex numbers", func() {
		// Arrange
		var m geom.Mesh

		// Act
		m.Tri(0, 1, 2)

		// Assert
		Expect(m.Triangles).To(Equal([][3]int{{0, 1, 2}}))
	})

	It("splits a quad into two triangles with the same winding", func() {
		// Arrange
		var m geom.Mesh

		// Act
		m.Quad(0, 1, 2, 3)

		// Assert
		Expect(m.Triangles).To(Equal([][3]int{{0, 1, 2}, {0, 2, 3}}))
	})
})

var _ = Describe("Outline", func() {
	square := geom.Outline{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}}

	It("has positive area when counterclockwise (shoelace formula)", func() {
		// Arrange
		o := square

		// Act
		a := o.Area()

		// Assert
		Expect(a).To(BeNumerically("~", 4, tol))
	})

	It("has negative area when clockwise", func() {
		// Arrange
		o := geom.Outline{square[3], square[2], square[1], square[0]}

		// Act
		a := o.Area()

		// Assert
		Expect(a).To(BeNumerically("~", -4, tol))
	})

	It("is finite only when every point is", func() {
		// Arrange
		bad := geom.Outline{{X: 0}, {X: math.NaN()}, {Y: 1}}

		// Act
		good, notGood := square.Finite(), bad.Finite()

		// Assert
		Expect(good).To(BeTrue())
		Expect(notGood).To(BeFalse())
	})

	It("measures its bounding box", func() {
		// Arrange
		o := geom.Outline{{X: -1, Y: 3}, {X: 4, Y: -2}, {X: 0, Y: 5}}

		// Act
		lo, hi := o.Bounds()

		// Assert
		Expect(lo).To(Equal(geom.Vec2{X: -1, Y: -2}))
		Expect(hi).To(Equal(geom.Vec2{X: 4, Y: 5}))
	})
})
