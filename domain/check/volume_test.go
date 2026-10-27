package check_test

import (
	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// box returns a closed box mesh of the given size.
func box(x, y, z float64) *geom.Mesh {
	m := &geom.Mesh{}
	for _, v := range []geom.Vec3{{}, {X: x}, {X: x, Y: y}, {Y: y}, {Z: z}, {X: x, Z: z}, {X: x, Y: y, Z: z}, {Y: y, Z: z}} {
		m.Add(v)
	}
	for _, q := range [][4]int{{0, 3, 2, 1}, {4, 5, 6, 7}, {0, 1, 5, 4}, {1, 2, 6, 5}, {2, 3, 7, 6}, {3, 0, 4, 7}} {
		m.Quad(q[0], q[1], q[2], q[3])
	}
	return m
}

var printer = check.Context{Printer: [3]float64{256, 200, 250}}

var _ = Describe("BuildVolume", func() {
	DescribeTable("is quiet when the shape fits, either way round on the bed",
		func(x, y, z float64) {
			// Arrange
			s := family.Shape{Mesh: box(x, y, z)}

			// Act
			got := check.BuildVolume{}.Run(s, printer)

			// Assert
			Expect(got).To(BeEmpty())
		},
		Entry("as it is", 250.0, 190.0, 250.0),
		Entry("turned a quarter", 190.0, 250.0, 100.0),
	)

	It("warns when it doesn't fit either way", func() {
		// Arrange
		s := family.Shape{Mesh: box(300, 120, 80)}

		// Act
		got := check.BuildVolume{}.Run(s, printer)

		// Assert
		Expect(got).To(ConsistOf(check.Finding{Level: check.Warning,
			Message: "At 300 × 120 × 80 mm, it doesn't fit a 256 × 200 × 250 mm printer, even turned. Make it smaller, or change the printer size."}))
	})

	It("warns when it is too tall", func() {
		// Arrange
		s := family.Shape{Mesh: box(100, 100, 251.5)}

		// Act
		got := check.BuildVolume{}.Run(s, printer)

		// Assert
		Expect(got).To(HaveLen(1))
		Expect(got[0].Message).To(HavePrefix("At 100 × 100 × 251.5 mm"))
	})

	It("doesn't check flat shapes, or when no printer is given", func() {
		// Arrange
		flat := family.Shape{Outlines: []geom.Outline{square}}
		big := family.Shape{Mesh: box(900, 900, 900)}

		// Act
		gotFlat, gotNone := check.BuildVolume{}.Run(flat, printer), check.BuildVolume{}.Run(big, check.Context{})

		// Assert
		Expect(gotFlat).To(BeEmpty())
		Expect(gotNone).To(BeEmpty())
	})
})
