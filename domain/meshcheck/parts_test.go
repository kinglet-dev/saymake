package meshcheck_test

import (
	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/meshcheck"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// twoCubes is a big cube and a small one beside it; flip turns the small
// one inside out.
func twoCubes(flip bool) geom.Mesh {
	m := cube()
	small := cube()
	base := len(m.Vertices)
	for _, v := range small.Vertices {
		m.Add(geom.Vec3{X: 5 + v.X*0.5, Y: v.Y * 0.5, Z: v.Z * 0.5})
	}
	for _, t := range small.Triangles {
		if flip {
			t[1], t[2] = t[2], t[1]
		}
		m.Tri(base+t[0], base+t[1], base+t[2])
	}
	return m
}

var _ = Describe("Inspect with several parts", func() {
	It("counts separate closed parts, and finds them sound when each faces out", func() {
		// Arrange
		m := twoCubes(false)

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Parts).To(Equal(2))
		Expect(r.Sound()).To(Succeed())
	})

	It("says how many parts are inside out", func() {
		// Arrange: a big cube, and two small ones inside out
		m := twoCubes(true)
		extra := twoCubes(true)
		base := len(m.Vertices)
		for _, v := range extra.Vertices[8:] {
			m.Add(geom.Vec3{X: v.X + 5, Y: v.Y, Z: v.Z})
		}
		for _, t := range extra.Triangles[12:] {
			m.Tri(base+t[0]-8, base+t[1]-8, base+t[2]-8)
		}

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Parts).To(Equal(3))
		Expect(r.Sound()).To(MatchError(ContainSubstring("2 parts are inside out")))
	})

	It("refuses a closed part with no volume beside a sound one", func() {
		// Arrange: a cube, and one triangle front and back
		m := cube()
		a, b, c := m.Add(geom.Vec3{X: 5}), m.Add(geom.Vec3{X: 6}), m.Add(geom.Vec3{X: 5, Y: 1})
		m.Tri(a, b, c)
		m.Tri(a, c, b)

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Sound()).To(MatchError(ContainSubstring("1 part is inside out")))
	})

	It("refuses a part turned inside out, even when the whole has volume", func() {
		// Arrange
		m := twoCubes(true)

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Volume).To(BeNumerically(">", 0))
		Expect(r.InsideOutParts).To(Equal(1))
		Expect(r.Sound()).To(MatchError(ContainSubstring("1 part is inside out")))
	})
})
