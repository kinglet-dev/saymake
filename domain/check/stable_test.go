package check_test

import (
	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// leaning returns a box of the given size sheared in x by k mm per mm up,
// so its base stays put and its top moves k·z across.
func leaning(x, y, z, k float64) *geom.Mesh {
	m := box(x, y, z)
	for i, v := range m.Vertices {
		m.Vertices[i].X = v.X + k*v.Z
	}
	return m
}

// apexDown returns a tetrahedron standing on one corner.
func apexDown() *geom.Mesh {
	m := &geom.Mesh{}
	for _, v := range []geom.Vec3{{}, {X: 10, Z: 10}, {Y: 10, Z: 10}, {X: -10, Y: -10, Z: 10}} {
		m.Add(v)
	}
	for _, t := range [][3]int{{0, 2, 1}, {0, 1, 3}, {0, 3, 2}, {1, 2, 3}} {
		m.Tri(t[0], t[1], t[2])
	}
	return m
}

var tips = check.Finding{Level: check.Warning,
	Message: "It may tip over when empty: its center of weight isn't above what it stands on. Make the base wider, or make it lean less."}

var _ = Describe("Stable", func() {
	DescribeTable("is quiet when the center of weight is above what it stands on",
		func(m *geom.Mesh) {
			// Act
			got := check.Stable{}.Run(family.Shape{Mesh: m}, check.Context{})

			// Assert
			Expect(got).To(BeEmpty())
		},
		Entry("an upright box", box(10, 10, 50)),
		Entry("a box leaning less than half its width", leaning(10, 10, 50, 0.15)),
		Entry("no mesh (a flat shape)", nil),
		Entry("a mesh with no volume, which the solid check reports", &geom.Mesh{Vertices: []geom.Vec3{{}, {X: 1}, {Y: 1}}, Triangles: [][3]int{{0, 1, 2}}}),
		Entry("a mesh whose volume cancels to exactly zero, but whose weights don't", &geom.Mesh{
			Vertices:  []geom.Vec3{{X: 6}, {Y: 6}, {Z: 1}, {X: 1}, {Y: 1}, {Z: 36}},
			Triangles: [][3]int{{0, 1, 2}, {3, 5, 4}}}),
	)

	DescribeTable("warns when it would tip over",
		func(m *geom.Mesh) {
			// Act
			got := check.Stable{}.Run(family.Shape{Mesh: m}, check.Context{})

			// Assert
			Expect(got).To(ConsistOf(tips))
		},
		Entry("a box leaning more than half its width", leaning(10, 10, 50, 0.5)),
		Entry("standing on a corner", apexDown()),
	)

	It("counts the weight of parts that don't touch the bed", func() {
		// Arrange
		m := box(10, 10, 10)
		far := box(10, 10, 10)
		for _, v := range far.Vertices {
			m.Add(geom.Vec3{X: v.X + 40, Y: v.Y, Z: v.Z + 20})
		}
		for _, t := range far.Triangles {
			m.Tri(t[0]+8, t[1]+8, t[2]+8)
		}

		// Act
		got := check.Stable{}.Run(family.Shape{Mesh: m}, check.Context{})

		// Assert
		Expect(got).To(ConsistOf(tips))
	})
})
