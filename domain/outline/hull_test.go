package outline_test

import (
	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/outline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hull", func() {
	It("keeps the corners, counterclockwise from the lowest-leftmost, and drops points inside and on edges", func() {
		// Arrange
		pts := []geom.Vec2{{X: 2, Y: 2}, {X: 4, Y: 4}, {X: 0, Y: 4}, {X: 1, Y: 1}, {X: 0, Y: 0}, {X: 4, Y: 0}, {X: 2, Y: 0}, {X: 4, Y: 0}}

		// Act
		h := outline.Hull(pts)

		// Assert
		Expect(h).To(Equal(geom.Outline{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}}))
		Expect(h.Area()).To(BeNumerically("~", 16, tol))
	})

	It("breaks ties on x by y, so the start is the lowest of the leftmost", func() {
		// Arrange
		pts := []geom.Vec2{{X: 0, Y: 3}, {X: 3, Y: 1}, {X: 0, Y: 1}}

		// Act
		h := outline.Hull(pts)

		// Assert
		Expect(h).To(Equal(geom.Outline{{X: 0, Y: 1}, {X: 3, Y: 1}, {X: 0, Y: 3}}))
	})

	DescribeTable("returns nothing when the points enclose no area",
		func(pts []geom.Vec2) {
			// Act
			h := outline.Hull(pts)

			// Assert
			Expect(h).To(BeEmpty())
		},
		Entry("no points", []geom.Vec2{}),
		Entry("one point", []geom.Vec2{{X: 1, Y: 1}}),
		Entry("the same point twice", []geom.Vec2{{X: 1, Y: 1}, {X: 1, Y: 1}}),
		Entry("points on a line", []geom.Vec2{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2}}),
	)

	It("leaves its input as it was", func() {
		// Arrange
		pts := []geom.Vec2{{X: 4, Y: 0}, {X: 0, Y: 0}, {X: 0, Y: 4}}

		// Act
		outline.Hull(pts)

		// Assert
		Expect(pts).To(Equal([]geom.Vec2{{X: 4, Y: 0}, {X: 0, Y: 0}, {X: 0, Y: 4}}))
	})
})
