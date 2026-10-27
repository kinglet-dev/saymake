package vessel_test

import (
	"math"

	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vessel feet", func() {
	It("stands on a ring, with the base recessed inside it", func() {
		// Arrange
		v := cylinder()
		v["foot"], v["floor"] = 3, 5

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		n := 192
		ring := 40 - 4.0 // the foot ring is 4 mm wide
		want := polygonArea(n, 40)*100 - polygonArea(n, 38)*95 - polygonArea(n, ring)*3
		Expect(r.Volume).To(BeNumerically("~", want, want*1e-9))
		Expect(widestAt(s.Mesh.Vertices, 3)).To(BeNumerically(">", 0), "the recess ceiling is 3 mm up")
		Expect(s.Bridge).To(BeNumerically("~", 2*ring, 1e-9))
	})

	It("runs a drainage hole through the recessed base", func() {
		// Arrange
		v := cylinder()
		v["foot"], v["floor"], v["drain"] = 3, 5, 10

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		n := 192
		want := polygonArea(n, 40)*100 - polygonArea(n, 38)*95 - polygonArea(n, 36)*3 - polygonArea(n, 5)*2
		Expect(r.Volume).To(BeNumerically("~", want, want*1e-9))
		Expect(s.Bridge).To(BeNumerically("~", 72, 1e-9))
	})

	It("has no ceiling to bridge without a foot", func() {
		// Arrange
		v := cylinder()

		// Act
		s, _ := build(v)

		// Assert
		Expect(s.Bridge).To(BeZero())
	})

	DescribeTable("refuses feet it can't make",
		func(change map[string]float64, want string) {
			// Arrange
			v := cylinder()
			for k, x := range change {
				v[k] = x
			}

			// Act
			_, err := vessel.New().Build(v)

			// Assert
			Expect(err).To(MatchError(want))
		},
		Entry("floor not thicker than the foot", map[string]float64{"foot": 3, "floor": 3.5},
			"The floor must be at least 1 mm thicker than the foot is tall. Make the floor thicker or the foot lower."),
		Entry("base too narrow for a foot ring", map[string]float64{"foot": 2, "floor": 4, "diameter": 20, "wall": 1, "base": 0.5, "top": 0.5},
			"The base is too narrow for a foot ring. Make the base wider or remove the foot."),
		Entry("drainage hole wider than the recess", map[string]float64{"foot": 2, "floor": 4, "drain": 72},
			"The drainage hole is too wide for the foot. Make the hole smaller or remove the foot."),
	)

	It("keeps the foot ring inside a polygon's flat sides", func() {
		// Arrange: a triangle's sides are half as far as its corners
		v := cylinder()
		v["foot"], v["floor"], v["sides"] = 2, 4, 3

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(s.Bridge).To(BeNumerically("~", 2*(40*math.Cos(math.Pi/3)-4), 1e-9))
	})
})
