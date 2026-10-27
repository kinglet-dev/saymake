package outline_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/outline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoundCorners", func() {
	square := geom.Outline{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}}

	It("rounds each corner with an arc, keeping the edges where they were", func() {
		// Arrange
		r := 0.5

		// Act
		o, err := outline.RoundCorners(square, r)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		lo, hi := o.Bounds()
		Expect([]float64{lo.X, lo.Y, hi.X, hi.Y}).To(Equal([]float64{-1, -1, 1, 1}))
		want := 4 - (4-math.Pi)*r*r
		Expect(o.Area()).To(BeNumerically("~", want, 0.01))
		Expect(o.Area()).To(BeNumerically("<", want), "arcs are drawn through points on the circle")
		for _, p := range o {
			if math.Abs(p.X) > 0.5 && math.Abs(p.Y) > 0.5 {
				c := geom.Vec2{X: math.Copysign(0.5, p.X), Y: math.Copysign(0.5, p.Y)}
				Expect(math.Hypot(p.X-c.X, p.Y-c.Y)).To(BeNumerically("~", r, 1e-9), "corner points lie on the arc")
			}
		}
	})

	It("draws finer arcs at sharper corners", func() {
		// Arrange
		tri, _ := outline.Polygon(3, 10, 0)
		hex, _ := outline.Polygon(6, 10, 0)

		// Act
		t, errT := outline.RoundCorners(tri, 1)
		h, errH := outline.RoundCorners(hex, 1)

		// Assert
		Expect(errT).NotTo(HaveOccurred())
		Expect(errH).NotTo(HaveOccurred())
		Expect(len(t) / 3).To(BeNumerically(">", len(h)/6))
	})

	It("returns the outline as it is for a radius of zero", func() {
		// Arrange
		r := 0.0

		// Act
		o, err := outline.RoundCorners(square, r)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(o).To(Equal(square))
	})

	It("refuses a radius too big for the outline", func() {
		// Arrange
		r := 1.0

		// Act
		_, err := outline.RoundCorners(square, r)

		// Assert
		Expect(err).To(MatchError(outline.ErrCornerTooBig))
	})
})

var _ = Describe("Footprint with rounded corners", func() {
	It("rounds a polygon's corners after fitting it to the box", func() {
		// Arrange
		length, width := 100.0, 60.0

		// Act
		o, err := outline.FootprintRounded(length, width, 2, 6, 8, 64)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		sharp, _ := outline.Footprint(length, width, 2, 6, 64)
		Expect(o.Area()).To(BeNumerically("<", sharp.Area()))
		Expect(len(o)).To(BeNumerically(">", 6))
	})

	It("rounds a triangle's corners too", func() {
		// Arrange
		length, width := 100.0, 100.0

		// Act
		o, err := outline.FootprintRounded(length, width, 2, 3, 8, 64)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(len(o)).To(BeNumerically(">", 3))
	})

	It("ignores the corner radius for a shape without sides", func() {
		// Arrange
		length, width := 100.0, 60.0

		// Act
		o, err := outline.FootprintRounded(length, width, 4, 0, 8, 64)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		plain, _ := outline.Footprint(length, width, 4, 0, 64)
		Expect(o).To(Equal(plain))
	})

	It("passes on a refusal", func() {
		// Arrange
		length := 0.0

		// Act
		_, errSize := outline.FootprintRounded(length, 10, 2, 4, 1, 64)
		_, errCorner := outline.FootprintRounded(10, 10, 2, 4, 6, 64)

		// Assert
		Expect(errSize).To(HaveOccurred())
		Expect(errCorner).To(MatchError(outline.ErrCornerTooBig))
	})
})
