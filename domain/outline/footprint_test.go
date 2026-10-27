package outline_test

import (
	"github.com/kinglet-dev/saymake/domain/outline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Footprint", func() {
	It("is a superellipse filling the box when it has no sides", func() {
		// Arrange
		length, width := 200.0, 120.0

		// Act
		o, err := outline.Footprint(length, width, 2, 0, 64)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(o).To(HaveLen(64))
		lo, hi := o.Bounds()
		Expect(hi.X).To(BeNumerically("~", 100, tol))
		Expect(lo.Y).To(BeNumerically("~", -60, tol))
	})

	It("is a rectangle with 4 sides", func() {
		// Arrange
		length, width := 200.0, 120.0

		// Act
		o, err := outline.Footprint(length, width, 2, 4, 64)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(o).To(HaveLen(4))
		Expect(o.Area()).To(BeNumerically("~", 200*120, 1e-6))
	})

	It("sits a polygon on a flat side, filling the box", func() {
		// Arrange
		length, width := 90.0, 90.0

		// Act
		o, err := outline.Footprint(length, width, 2, 3, 64)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(o[0].Y).To(BeNumerically("~", o[1].Y, tol), "the first edge is level")
		Expect(o[0].Y).To(BeNumerically("~", -45, tol), "and at the bottom")
		lo, hi := o.Bounds()
		Expect(hi.X - lo.X).To(BeNumerically("~", 90, tol))
		Expect(hi.Y - lo.Y).To(BeNumerically("~", 90, tol))
	})

	It("refuses an impossible size, round or with sides", func() {
		// Arrange
		length := 0.0

		// Act
		_, errRound := outline.Footprint(length, 10, 2, 0, 64)
		_, errSides := outline.Footprint(10, length, 2, 4, 64)
		_, errLength := outline.Footprint(length, 10, 2, 4, 64)

		// Assert
		Expect(errRound).To(MatchError(ContainSubstring("must be positive")))
		Expect(errSides).To(MatchError(ContainSubstring("must be positive")))
		Expect(errLength).To(MatchError(ContainSubstring("must be positive")))
	})
})
