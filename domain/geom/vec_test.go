package geom_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const tol = 1e-9

var _ = Describe("Vec3", func() {
	It("subtracts component by component", func() {
		// Arrange
		a, b := geom.Vec3{X: 5, Y: 7, Z: 9}, geom.Vec3{X: 1, Y: 2, Z: 3}

		// Act
		d := a.Sub(b)

		// Assert
		Expect(d).To(Equal(geom.Vec3{X: 4, Y: 5, Z: 6}))
	})

	It("takes the right-handed cross product", func() {
		// Arrange
		x, y := geom.Vec3{X: 1}, geom.Vec3{Y: 1}

		// Act
		z := x.Cross(y)

		// Assert
		Expect(z).To(Equal(geom.Vec3{Z: 1}))
	})

	It("takes the dot product", func() {
		// Arrange
		a, b := geom.Vec3{X: 1, Y: 2, Z: 3}, geom.Vec3{X: 4, Y: -5, Z: 6}

		// Act
		d := a.Dot(b)

		// Assert
		Expect(d).To(BeNumerically("~", 12, tol))
	})

	It("is finite only when every component is", func() {
		// Arrange
		good := geom.Vec3{X: 1, Y: 2, Z: 3}
		bad := []geom.Vec3{{X: math.NaN()}, {Y: math.Inf(1)}, {Z: math.Inf(-1)}}

		// Act
		goodOK := good.Finite()

		// Assert
		Expect(goodOK).To(BeTrue())
		for _, v := range bad {
			Expect(v.Finite()).To(BeFalse(), "%v", v)
		}
	})
})

var _ = Describe("Vec2", func() {
	It("adds, subtracts and scales", func() {
		// Arrange
		a, b := geom.Vec2{X: 3, Y: 4}, geom.Vec2{X: 1, Y: -1}

		// Act
		sum, diff, scaled := a.Add(b), a.Sub(b), a.Scale(2)

		// Assert
		Expect(sum).To(Equal(geom.Vec2{X: 4, Y: 3}))
		Expect(diff).To(Equal(geom.Vec2{X: 2, Y: 5}))
		Expect(scaled).To(Equal(geom.Vec2{X: 6, Y: 8}))
	})

	It("takes the 2D cross product (z of the 3D one)", func() {
		// Arrange
		a, b := geom.Vec2{X: 1}, geom.Vec2{Y: 1}

		// Act
		c := a.Cross(b)

		// Assert
		Expect(c).To(BeNumerically("~", 1, tol))
	})

	It("is finite only when both components are", func() {
		// Arrange
		good, nanX, infY := geom.Vec2{X: 1, Y: 2}, geom.Vec2{X: math.NaN()}, geom.Vec2{Y: math.Inf(1)}

		// Act
		g, n, i := good.Finite(), nanX.Finite(), infY.Finite()

		// Assert
		Expect([]bool{g, n, i}).To(Equal([]bool{true, false, false}))
	})
})

var _ = Describe("Vec2.Dot", func() {
	It("takes the dot product", func() {
		// Arrange
		a, b := geom.Vec2{X: 2, Y: 3}, geom.Vec2{X: 4, Y: -1}

		// Act
		d := a.Dot(b)

		// Assert
		Expect(d).To(BeNumerically("~", 5, tol))
	})
})
