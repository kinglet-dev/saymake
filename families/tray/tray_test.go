package tray_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/meshcheck"
	"github.com/kinglet-dev/saymake/families/tray"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tray family", func() {
	It("registers cleanly, with length as its main size", func() {
		// Arrange
		f := tray.New()

		// Act
		_, err := family.NewRegistry(f)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{f.ID(), f.Label(), f.MainSize()}).To(Equal([]string{"tray", "Tray", "length"}))
		Expect(f.Proportional()).To(Equal([]string{"width"}))
	})

	It("builds its defaults as a sound solid that fills its box", func() {
		// Arrange
		v := tray.New().Params().Defaults()

		// Act
		s, err := tray.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		r := meshcheck.Inspect(*s.Mesh)
		Expect(r.Sound()).To(Succeed())
		Expect(r.Max.X).To(BeNumerically("~", v["length"]/2, 1e-9))
		Expect(r.Max.Y).To(BeNumerically("~", v["width"]/2, 1e-9))
		Expect(r.Max.Z).To(BeNumerically("~", v["height"], 1e-9))
	})

	It("builds a rectangular tray with the volume of a hollow box", func() {
		// Arrange
		v := box()

		// Act
		s, err := tray.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		r := meshcheck.Inspect(*s.Mesh)
		Expect(r.Sound()).To(Succeed())
		want := 200.0*120*30 - 196.0*116*27
		Expect(r.Volume).To(BeNumerically("~", want, want*1e-9))
		Expect(s.MinWall).To(BeNumerically("~", 2, 1e-9))
		Expect(s.Overhang).To(BeNumerically("~", 0, 1e-9))
	})

	It("flares its walls outward as they rise", func() {
		// Arrange
		v := box()
		v["flare"] = 20

		// Act
		s, err := tray.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		r := meshcheck.Inspect(*s.Mesh)
		Expect(r.Sound()).To(Succeed())
		Expect(r.Max.X).To(BeNumerically("~", 100+30*math.Tan(20*math.Pi/180), 1e-6))
		Expect(s.Overhang).To(BeNumerically("~", 20, 1e-9))
		Expect(s.MinWall).To(BeNumerically("~", 2*math.Cos(20*math.Pi/180), 1e-9))
	})

	It("reports the floor as the thinnest part when it is thinner than the wall", func() {
		// Arrange
		v := box()
		v["floor"] = 1

		// Act
		s, err := tray.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(s.MinWall).To(BeNumerically("~", 1, 1e-9))
	})

	It("rounds the corners of a polygon tray", func() {
		// Arrange
		v := box()
		v["corner"] = 10

		// Act
		s, err := tray.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		r := meshcheck.Inspect(*s.Mesh)
		Expect(r.Sound()).To(Succeed())
		Expect(r.Max.X).To(BeNumerically("~", 100, 1e-9), "the sides stay where they were")
		Expect(r.Volume).To(BeNumerically("<", 200.0*120*30-196.0*116*27))
	})

	It("refuses a corner radius too big for the tray", func() {
		// Arrange
		v := box()
		v["corner"] = 60

		// Act
		_, err := tray.New().Build(v)

		// Assert
		Expect(err).To(MatchError("The corner radius is too big for this tray. Make it smaller or the tray bigger."))
	})

	It("accepts a floor half a millimeter under the rim", func() {
		// Arrange
		v := box()
		v["height"], v["floor"] = 10, 9.5

		// Act
		_, err := tray.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
	})

	It("stays within its triangle budget", func() {
		// Arrange
		v := tray.New().Params().Defaults()

		// Act
		s, err := tray.New().Build(v)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(len(s.Mesh.Triangles)).To(BeNumerically("<=", tray.MaxTriangles))
	})

	DescribeTable("refuses shapes it can't make",
		func(change map[string]float64, drop string, want string) {
			// Arrange
			v := box()
			for k, x := range change {
				v[k] = x
			}
			delete(v, drop)

			// Act
			_, err := tray.New().Build(v)

			// Assert
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("wall too thick", map[string]float64{"length": 20, "width": 20, "wall": 10}, "", "The wall is too thick for this tray"),
		Entry("wall too thick at the floor of a flared tray, though not at the rim",
			map[string]float64{"sides": 3, "length": 20, "width": 20, "wall": 7, "floor": 0.8, "height": 150, "flare": 30}, "",
			"The wall is too thick for this tray"),
		Entry("floor too thick", map[string]float64{"height": 5, "floor": 5}, "", "The floor is too thick for the height"),
		Entry("a missing setting", nil, "wall", `setting "wall" is missing`),
		Entry("a size given directly outside the ranges", map[string]float64{"length": 0}, "", "must be positive"),
	)
})
