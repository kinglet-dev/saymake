package vessel_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("A leaning vessel", func() {
	It("leans its whole body over, with the base flat where it was", func() {
		// Arrange
		v := cylinder()
		v["lean"] = 20
		shift := 100 * math.Tan(20*math.Pi/180)

		// Act
		s, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(r.Min.Z).To(BeNumerically("~", 0, 1e-9))
		Expect(highestWhere(s.Mesh.Vertices, func(p geom.Vec3) bool { return p.Z < 1e-9 && p.X < -39.9 })).To(BeNumerically("~", 0, 1e-9), "the base is where it was")
		Expect(r.Max.X).To(BeNumerically("~", 40+shift, 1e-9))
	})

	It("holds as much as the same vessel upright", func() {
		// Arrange
		upright, leaning := cylinder(), cylinder()
		leaning["lean"] = 20

		// Act
		_, ru := build(upright)
		_, rl := build(leaning)

		// Assert
		Expect(rl.Volume).To(BeNumerically("~", ru.Volume, 1e-6*ru.Volume))
	})

	It("adds its lean to the overhang, and thins the wall across", func() {
		// Arrange
		v := cylinder()
		v["lean"] = 20

		// Act
		s, _ := build(v)

		// Assert
		Expect(s.Overhang).To(BeNumerically("~", 20, 1e-9))
		Expect(s.MinWall).To(BeNumerically("~", 2*math.Cos(20*math.Pi/180), 1e-9))
	})

	It("adds the lean's slope to the profile's", func() {
		// Arrange
		v := cylinder()
		v["top"], v["lean"] = 1.2, 10
		upright := cylinder()
		upright["top"] = 1.2
		su, _ := build(upright)

		// Act
		s, _ := build(v)

		// Assert
		want := math.Atan(math.Tan(su.Overhang*math.Pi/180)+math.Tan(10*math.Pi/180)) * 180 / math.Pi
		Expect(s.Overhang).To(BeNumerically("~", want, 1e-9))
	})

	It("builds sound with a handle and a slanted rim", func() {
		// Arrange
		v := cylinder()
		v["lean"], v["slant"], v["handle"] = 15, 20, 0.6

		// Act
		_, r := build(v)

		// Assert
		Expect(r.Sound()).To(Succeed())
	})
})
