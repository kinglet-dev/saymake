package outline_test

import (
	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/outline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fit", func() {
	It("stretches and centers an outline to fill a box", func() {
		// Arrange
		tri := geom.Outline{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 1, Y: 4}}

		// Act
		got := outline.Fit(tri, 10, 6)

		// Assert
		lo, hi := got.Bounds()
		Expect(lo.X).To(BeNumerically("~", -5, tol))
		Expect(hi.X).To(BeNumerically("~", 5, tol))
		Expect(lo.Y).To(BeNumerically("~", -3, tol))
		Expect(hi.Y).To(BeNumerically("~", 3, tol))
		Expect(got.Area()).To(BeNumerically(">", 0))
	})
})

var _ = Describe("Inset outward", func() {
	It("moves every edge outward for a negative distance", func() {
		// Arrange
		square := geom.Outline{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}}

		// Act
		out, err := outline.Inset(square, -1)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		lo, hi := out.Bounds()
		Expect(lo.X).To(BeNumerically("~", -2, tol))
		Expect(hi.Y).To(BeNumerically("~", 2, tol))
	})
})

var _ = Describe("Contains", func() {
	square := geom.Outline{{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}}

	DescribeTable("tells whether a point is inside a convex outline",
		func(p geom.Vec2, want bool) {
			// Arrange
			// (point from the table)

			// Act
			got := outline.Contains(square, p)

			// Assert
			Expect(got).To(Equal(want))
		},
		Entry("center", geom.Vec2{}, true),
		Entry("near a corner", geom.Vec2{X: 0.99, Y: 0.99}, true),
		Entry("right of it", geom.Vec2{X: 1.01}, false),
		Entry("above it", geom.Vec2{Y: 1.5}, false),
		Entry("on an edge counts as outside", geom.Vec2{X: 1}, false),
	)
})
