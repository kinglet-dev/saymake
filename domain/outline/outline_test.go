package outline_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/outline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const tol = 1e-9

var _ = Describe("Superellipse", func() {
	It("with exponent 2 is an ellipse, counterclockwise, with the asked number of points", func() {
		// Arrange
		a, b := 40.0, 25.0

		// Act
		o, err := outline.Superellipse(a, b, 2, 64)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(o).To(HaveLen(64))
		for _, p := range o {
			Expect(p.X*p.X/(a*a) + p.Y*p.Y/(b*b)).To(BeNumerically("~", 1, 1e-9))
		}
		Expect(o.Area()).To(BeNumerically("~", math.Pi*a*b, 0.01*math.Pi*a*b))
	})

	It("approaches a rectangle as the exponent grows", func() {
		// Arrange
		a, b := 40.0, 25.0

		// Act
		o, err := outline.Superellipse(a, b, 40, 256)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(o.Area()).To(BeNumerically(">", 0.97*4*a*b))
		Expect(o.Area()).To(BeNumerically("<", 4*a*b))
	})

	DescribeTable("refuses impossible sizes",
		func(a, b, n float64, segs int, want string) {
			// Arrange
			// (inputs from the table)

			// Act
			_, err := outline.Superellipse(a, b, n, segs)

			// Assert
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("zero width", 0.0, 1.0, 2.0, 16, "half-width and half-length must be positive"),
		Entry("negative length", 1.0, -1.0, 2.0, 16, "half-width and half-length must be positive"),
		Entry("zero exponent", 1.0, 1.0, 0.0, 16, "roundness must be positive"),
		Entry("too few points", 1.0, 1.0, 2.0, 2, "at least 3 points"),
	)
})

var _ = Describe("Polygon", func() {
	It("is regular, counterclockwise, with the textbook area", func() {
		// Arrange
		r := 10.0

		// Act
		o, err := outline.Polygon(6, r, 0)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(o).To(HaveLen(6))
		for _, p := range o {
			Expect(math.Hypot(p.X, p.Y)).To(BeNumerically("~", r, tol))
		}
		Expect(o.Area()).To(BeNumerically("~", 3*r*r*math.Sin(math.Pi/3), 1e-9))
	})

	It("starts at the given angle", func() {
		// Arrange
		turn := math.Pi / 2

		// Act
		o, err := outline.Polygon(4, 1, turn)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(o[0].X).To(BeNumerically("~", 0, tol))
		Expect(o[0].Y).To(BeNumerically("~", 1, tol))
	})

	It("refuses fewer than 3 sides or a non-positive radius", func() {
		// Arrange
		// (none)

		// Act
		_, errSides := outline.Polygon(2, 1, 0)
		_, errRadius := outline.Polygon(5, 0, 0)

		// Assert
		Expect(errSides).To(MatchError(ContainSubstring("at least 3 sides")))
		Expect(errRadius).To(MatchError(ContainSubstring("radius must be positive")))
	})
})

var _ = Describe("Inset", func() {
	square := geom.Outline{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}}

	It("moves every edge of a convex outline inward by the distance", func() {
		// Arrange
		d := 0.5

		// Act
		in, err := outline.Inset(square, d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		want := []geom.Vec2{{X: -0.5, Y: -0.5}, {X: 0.5, Y: -0.5}, {X: 0.5, Y: 0.5}, {X: -0.5, Y: 0.5}}
		for i, p := range in {
			Expect(p.X).To(BeNumerically("~", want[i].X, tol))
			Expect(p.Y).To(BeNumerically("~", want[i].Y, tol))
		}
	})

	It("keeps a point between two edges that run straight on", func() {
		// Arrange
		withMid := geom.Outline{{X: -1, Y: -1}, {X: 0, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}}

		// Act
		in, err := outline.Inset(withMid, 0.25)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(in[1].X).To(BeNumerically("~", 0, tol))
		Expect(in[1].Y).To(BeNumerically("~", -0.75, tol))
	})

	It("refuses an inset that would turn the outline inside out", func() {
		// Arrange
		d := 1.2

		// Act
		_, err := outline.Inset(square, d)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("too thick for this outline")))
	})

	It("refuses an inset that would shrink the outline to nothing", func() {
		// Arrange
		d := 1.0

		// Act
		_, err := outline.Inset(square, d)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("too thick for this outline")))
	})

	It("insets an ellipse to a smaller, still counterclockwise outline", func() {
		// Arrange
		e, err := outline.Superellipse(50, 30, 2, 128)
		Expect(err).NotTo(HaveOccurred())

		// Act
		in, err := outline.Inset(e, 2)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(in.Area()).To(BeNumerically(">", 0))
		Expect(in.Area()).To(BeNumerically("<", e.Area()))
		lo, hi := in.Bounds()
		Expect(hi.X).To(BeNumerically("~", 48, 0.05))
		Expect(lo.Y).To(BeNumerically("~", -28, 0.05))
	})
})
