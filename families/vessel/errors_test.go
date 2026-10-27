package vessel_test

import (
	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vessel family: shapes it can't make", func() {
	It("refuses a wall too thick for the narrowest part", func() {
		// Arrange
		v := cylinder()
		v["neck"], v["wall"] = 0.8, 10

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("The wall is too thick for the narrowest part")))
	})

	It("refuses a floor as thick as the vessel is tall", func() {
		// Arrange
		v := cylinder()
		v["height"], v["floor"] = 20, 19.5

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("The floor is too thick for the height")))
	})

	It("refuses a drainage hole wider than the floor", func() {
		// Arrange
		v := cylinder()
		v["diameter"], v["drain"] = 30, 30

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("The drainage hole is too wide for the floor")))
	})

	It("measures the floor of a polygon across its flat sides, not its corners", func() {
		// Arrange: a triangle 60 mm across its corners, wall 2 mm, has a
		// floor 26 mm to its corners but only 13 mm to its sides.
		v := cylinder()
		v["diameter"], v["sides"], v["drain"] = 60, 3, 30

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("The drainage hole is too wide for the floor")))
	})

	It("measures a fluted floor at the bottom of its flutes", func() {
		// Arrange: inside, the flutes reach 38 mm from the axis but dip to 30.4 mm
		v := cylinder()
		v["flutes"], v["fluteDepth"], v["drain"] = 8, 0.2, 66

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("The drainage hole is too wide for the floor")))
	})

	It("measures the narrowest part of a polygon across its flat sides", func() {
		// Arrange: inside, the corners are 1.6 mm from the axis but the
		// sides only 0.8 mm (a triangle's sides are half as far as its corners).
		v := cylinder()
		v["diameter"], v["sides"], v["wall"] = 20, 3, 4.2

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("The wall is too thick for the narrowest part")))
	})

	It("refuses values it wasn't given", func() {
		// Arrange
		v := cylinder()
		delete(v, "wall")

		// Act
		_, err := vessel.New().Build(v)

		// Assert
		Expect(err).To(MatchError(`setting "wall" is missing`))
	})
})
