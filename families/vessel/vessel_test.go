package vessel_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/meshcheck"
	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func build(v param.Values) (family.Shape, meshcheck.Report) {
	s, err := vessel.New().Build(v)
	Expect(err).NotTo(HaveOccurred())
	Expect(s.Mesh).NotTo(BeNil())
	return s, meshcheck.Inspect(*s.Mesh)
}

// polygonArea is the area of a regular n-gon with circumradius r.
func polygonArea(n int, r float64) float64 {
	return float64(n) / 2 * r * r * math.Sin(2*math.Pi/float64(n))
}

var _ = Describe("Vessel family", func() {
	It("registers cleanly: valid settings, height as its main size", func() {
		// Arrange
		f := vessel.New()

		// Act
		_, err := family.NewRegistry(f)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(f.ID()).To(Equal("vessel"))
		Expect(f.Label()).To(Equal("Vessel"))
		Expect(f.MainSize()).To(Equal("height"))
		Expect(f.Proportional()).To(Equal([]string{"diameter"}))
	})

	It("builds its defaults as a sound solid", func() {
		// Arrange
		v := vessel.New().Params().Defaults()

		// Act
		_, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
	})

	It("builds a straight cup with the volume of a hollow cylinder", func() {
		// Arrange
		v := cylinder()

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		n := 192
		want := polygonArea(n, 40)*100 - polygonArea(n, 38)*97
		Expect(r.Volume).To(BeNumerically("~", want, want*1e-9))
		Expect(r.Min.Z).To(BeNumerically("~", 0, 1e-9))
		Expect(r.Max.Z).To(BeNumerically("~", 100, 1e-9))
		Expect(r.Max.X).To(BeNumerically("~", 40, 1e-9))
		Expect(s.Overhang).To(BeNumerically("~", 0, 1e-9))
		Expect(s.MinWall).To(BeNumerically("~", 2, 1e-9))
	})

	It("reports the floor as the thinnest part when it is thinner than the wall", func() {
		// Arrange
		v := cylinder()
		v["floor"] = 1.2

		// Act
		s, _ := build(v)

		// Assert
		Expect(s.MinWall).To(BeNumerically("~", 1.2, 1e-9))
	})

	It("tapers from the base to the top", func() {
		// Arrange
		v := cylinder()
		v["base"], v["top"] = 0.5, 1.5

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(r.Max.X).To(BeNumerically("~", 60, 1e-6))
		want := math.Atan(40.0/100) * 180 / math.Pi
		Expect(s.Overhang).To(BeNumerically("~", want, 0.01))
		Expect(s.MinWall).To(BeNumerically("~", 2*math.Cos(want*math.Pi/180), 0.01))
	})
})
