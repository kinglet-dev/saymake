package flat_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/families/flat"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const tol = 1e-9

func with(change map[string]float64) param.Values {
	v := flat.New().Params().Defaults()
	for k, x := range change {
		v[k] = x
	}
	return v
}

var _ = Describe("Flat family", func() {
	It("registers cleanly, scaling width with length", func() {
		// Arrange
		f := flat.New()

		// Act
		_, err := family.NewRegistry(f)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{f.ID(), f.Label(), f.MainSize()}).To(Equal([]string{"flat", "Flat shape", "length"}))
		Expect(f.Proportional()).To(Equal([]string{"width"}))
	})

	It("builds its defaults as one cut outline filling its box, with no solid", func() {
		// Arrange
		v := with(nil)

		// Act
		s, err := flat.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(s.Mesh).To(BeNil())
		Expect(s.Outlines).To(HaveLen(1))
		Expect(s.Scores).To(BeEmpty())
		lo, hi := s.Outlines[0].Bounds()
		Expect(hi.X - lo.X).To(BeNumerically("~", v["length"], tol))
		Expect(hi.Y - lo.Y).To(BeNumerically("~", v["width"], tol))
	})

	It("scores a border inside the edge", func() {
		// Arrange
		v := with(map[string]float64{"sides": 4, "length": 100, "width": 60, "border": 5})

		// Act
		s, err := flat.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(s.Scores).To(HaveLen(1))
		Expect(s.Scores[0].Area()).To(BeNumerically("~", 90*50, 1e-6))
	})

	It("cuts a round hanging hole below the top edge", func() {
		// Arrange
		v := with(map[string]float64{"length": 100, "width": 100, "hole": 6, "holeFrom": 8})

		// Act
		s, err := flat.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(s.Outlines).To(HaveLen(2))
		lo, hi := s.Outlines[1].Bounds()
		Expect((lo.Y + hi.Y) / 2).To(BeNumerically("~", 42, 1e-6))
		Expect((lo.X + hi.X) / 2).To(BeNumerically("~", 0, 1e-6))
		Expect(hi.X - lo.X).To(BeNumerically("~", 6, 1e-6))
		Expect(s.Outlines[1].Area()).To(BeNumerically("~", math.Pi*9, 0.05))
	})

	It("rounds the corners of a polygon", func() {
		// Arrange
		v := with(map[string]float64{"sides": 6, "corner": 8})

		// Act
		s, err := flat.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(len(s.Outlines[0])).To(BeNumerically(">", 6))
	})

	DescribeTable("refuses shapes it can't make",
		func(change map[string]float64, drop, want string) {
			// Arrange
			v := with(change)
			delete(v, drop)

			// Act
			_, err := flat.New().Build(v)

			// Assert
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("corner radius too big", map[string]float64{"sides": 4, "length": 20, "width": 20, "corner": 12}, "", "The corner radius is too big for this shape"),
		Entry("border too far in", map[string]float64{"length": 20, "width": 20, "border": 12}, "", "The border is too far in for this shape"),
		Entry("hole past the edge", map[string]float64{"hole": 20, "holeFrom": 3}, "", "The hole doesn't fit inside the shape"),
		Entry("hole in a shape too small for any", map[string]float64{"length": 1.5, "width": 1.5, "hole": 0.5, "holeFrom": 0.75}, "", "The hole doesn't fit inside the shape"),
		Entry("a size given directly outside the ranges", map[string]float64{"length": 0}, "", "must be positive"),
		Entry("a missing setting", nil, "border", `setting "border" is missing`),
	)
})
