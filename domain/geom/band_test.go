package geom_test

import (
	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mesh.Band", func() {
	It("joins two rings with quads, wrapping around", func() {
		// Arrange
		var m geom.Mesh

		// Act
		m.Band([]int{0, 1, 2}, []int{3, 4, 5}, false)

		// Assert
		Expect(m.Triangles).To(Equal([][3]int{
			{0, 1, 4}, {0, 4, 3}, {1, 2, 5}, {1, 5, 4}, {2, 0, 3}, {2, 3, 5},
		}))
	})

	It("reverses the faces when flipped", func() {
		// Arrange
		var m geom.Mesh

		// Act
		m.Band([]int{0, 1}, []int{2, 3}, true)

		// Assert
		Expect(m.Triangles).To(Equal([][3]int{{0, 2, 3}, {0, 3, 1}, {1, 3, 2}, {1, 2, 0}}))
	})
})

var _ = Describe("Mesh.Fan", func() {
	It("joins a ring to a center, facing up", func() {
		// Arrange
		var m geom.Mesh

		// Act
		m.Fan(9, []int{0, 1, 2}, false)

		// Assert
		Expect(m.Triangles).To(Equal([][3]int{{9, 0, 1}, {9, 1, 2}, {9, 2, 0}}))
	})

	It("faces down when asked", func() {
		// Arrange
		var m geom.Mesh

		// Act
		m.Fan(9, []int{0, 1, 2}, true)

		// Assert
		Expect(m.Triangles).To(Equal([][3]int{{9, 1, 0}, {9, 2, 1}, {9, 0, 2}}))
	})
})

var _ = Describe("Mesh.Ring", func() {
	It("adds an outline's points at a height and returns their numbers", func() {
		// Arrange
		var m geom.Mesh
		m.Add(geom.Vec3{})

		// Act
		ids := m.Ring(geom.Outline{{X: 1, Y: 2}, {X: 3, Y: 4}}, 5)

		// Assert
		Expect(ids).To(Equal([]int{1, 2}))
		Expect(m.Vertices[1:]).To(Equal([]geom.Vec3{{X: 1, Y: 2, Z: 5}, {X: 3, Y: 4, Z: 5}}))
	})
})
