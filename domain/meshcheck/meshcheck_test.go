package meshcheck_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/meshcheck"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const tol = 1e-9

// cube returns a closed unit cube wound outward.
func cube() geom.Mesh {
	var m geom.Mesh
	for _, v := range []geom.Vec3{
		{}, {X: 1}, {X: 1, Y: 1}, {Y: 1},
		{Z: 1}, {X: 1, Z: 1}, {X: 1, Y: 1, Z: 1}, {Y: 1, Z: 1},
	} {
		m.Add(v)
	}
	m.Quad(0, 3, 2, 1) // bottom, facing −z
	m.Quad(4, 5, 6, 7) // top
	m.Quad(0, 1, 5, 4) // front (−y)
	m.Quad(1, 2, 6, 5) // right (+x)
	m.Quad(2, 3, 7, 6) // back (+y)
	m.Quad(3, 0, 4, 7) // left (−x)
	return m
}

var _ = Describe("Inspect", func() {
	It("finds a closed, outward cube sound, with its volume and bounds", func() {
		// Arrange
		m := cube()

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Sound()).To(Succeed())
		Expect(r.Volume).To(BeNumerically("~", 1, tol))
		Expect(r.Min).To(Equal(geom.Vec3{}))
		Expect(r.Max).To(Equal(geom.Vec3{X: 1, Y: 1, Z: 1}))
		Expect(r.Triangles).To(Equal(12))
	})

	It("counts the open edges of a mesh with a hole", func() {
		// Arrange
		m := cube()
		m.Triangles = m.Triangles[:11]

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.OpenEdges).To(Equal(3))
		Expect(r.Sound()).To(MatchError(ContainSubstring("3 open edges")))
	})

	It("counts edges shared by triangles wound the same way", func() {
		// Arrange
		m := cube()
		t := m.Triangles[0]
		m.Triangles[0] = [3]int{t[0], t[2], t[1]}

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.MisorientedEdges).To(Equal(3))
		Expect(r.Sound()).To(MatchError(ContainSubstring("3 edges where faces are wound inconsistently")))
	})

	It("counts edges used by more than two triangles", func() {
		// Arrange
		m := cube()
		m.Tri(0, 1, 2)
		m.Tri(0, 2, 1)

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.OverusedEdges).To(Equal(3))
		Expect(r.Sound()).To(MatchError(ContainSubstring("3 edges shared by more than two faces")))
	})

	It("refuses an inside-out mesh by its negative volume", func() {
		// Arrange
		m := cube()
		for i, t := range m.Triangles {
			m.Triangles[i] = [3]int{t[0], t[2], t[1]}
		}

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Volume).To(BeNumerically("~", -1, tol))
		Expect(r.Sound()).To(MatchError(ContainSubstring("inside out")))
	})

	It("refuses a triangle that names a missing vertex", func() {
		// Arrange
		m := cube()
		m.Tri(0, 1, 99)

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Sound()).To(MatchError(ContainSubstring("missing vertex")))
	})

	It("refuses a vertex number one past the last", func() {
		// Arrange
		m := cube()
		m.Tri(0, 1, len(m.Vertices))

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Sound()).To(MatchError(ContainSubstring("missing vertex")))
	})

	It("refuses a closed, consistent surface with no volume", func() {
		// Arrange: one triangle, front and back
		var m geom.Mesh
		a, b, c := m.Add(geom.Vec3{}), m.Add(geom.Vec3{X: 1}), m.Add(geom.Vec3{Y: 1})
		m.Tri(a, b, c)
		m.Tri(a, c, b)

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Volume).To(BeNumerically("==", 0))
		Expect(r.Sound()).To(MatchError(ContainSubstring("inside out")))
	})

	It("refuses a negative vertex number", func() {
		// Arrange
		m := cube()
		m.Tri(-1, 1, 2)

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Sound()).To(MatchError(ContainSubstring("missing vertex")))
	})

	It("refuses a vertex that isn't a finite number", func() {
		// Arrange
		m := cube()
		m.Vertices[6].Z = math.Inf(1)

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Sound()).To(MatchError(ContainSubstring("not a finite number")))
	})

	It("counts triangles with no area, which slicers can't use", func() {
		// Arrange: two triangles whose corners lie on one line, closed
		// against each other (a flat sliver).
		var m geom.Mesh
		a, b, c := m.Add(geom.Vec3{}), m.Add(geom.Vec3{X: 1}), m.Add(geom.Vec3{X: 2})
		m.Tri(a, b, c)
		m.Tri(a, c, b)

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Degenerate).To(Equal(2))
		Expect(r.Sound()).To(MatchError(ContainSubstring("2 triangles with no area")))
	})

	It("refuses an empty mesh", func() {
		// Arrange
		var m geom.Mesh

		// Act
		r := meshcheck.Inspect(m)

		// Assert
		Expect(r.Sound()).To(MatchError(ContainSubstring("no triangles")))
	})
})
