package text_test

import (
	"github.com/kinglet-dev/saymake/domain/text"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Millimeters", func() {
	DescribeTable("gives the size of a unit in millimeters",
		func(unit string, want float64) {
			// Arrange
			// (unit from the table)

			// Act
			mm, ok := text.Millimeters(unit)

			// Assert
			Expect(ok).To(BeTrue())
			Expect(mm).To(BeNumerically("~", want, 1e-12))
		},
		Entry("mm", "mm", 1.0),
		Entry("millimeters", "millimeters", 1.0),
		Entry("millimetres", "millimetres", 1.0),
		Entry("cm", "cm", 10.0),
		Entry("centimeter", "centimeter", 10.0),
		Entry("centimetres", "centimetres", 10.0),
		Entry("m", "m", 1000.0),
		Entry("meters", "meters", 1000.0),
		Entry("in", "in", 25.4),
		Entry("inch", "inch", 25.4),
		Entry("inches", "inches", 25.4),
		Entry("ft", "ft", 304.8),
		Entry("foot", "foot", 304.8),
		Entry("feet", "feet", 304.8),
	)

	It("knows no other units", func() {
		// Arrange
		unit := "cubit"

		// Act
		_, ok := text.Millimeters(unit)

		// Assert
		Expect(ok).To(BeFalse())
	})
})
