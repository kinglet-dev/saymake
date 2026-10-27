package check_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func tetra() *geom.Mesh {
	m := &geom.Mesh{}
	a, b := m.Add(geom.Vec3{}), m.Add(geom.Vec3{X: 1})
	c, d := m.Add(geom.Vec3{Y: 1}), m.Add(geom.Vec3{Z: 1})
	m.Tri(a, c, b)
	m.Tri(a, b, d)
	m.Tri(b, c, d)
	m.Tri(c, a, d)
	return m
}

var square = geom.Outline{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}}

var _ = Describe("Solid", func() {
	It("finds nothing wrong with a closed solid and real outlines", func() {
		// Arrange
		s := family.Shape{Mesh: tetra(), Outlines: []geom.Outline{square}, Scores: []geom.Outline{square}}

		// Act
		got := check.Solid{}.Run(s, check.Context{})

		// Assert
		Expect(got).To(BeEmpty())
	})

	It("reports an open mesh as an error", func() {
		// Arrange
		m := tetra()
		m.Triangles = m.Triangles[:3]

		// Act
		got := check.Solid{}.Run(family.Shape{Mesh: m}, check.Context{})

		// Assert
		Expect(got).To(ConsistOf(check.Finding{Level: check.Error, Message: "The shape didn't come out as a closed solid (the mesh is not a closed solid: 3 open edges). Please report this with your description."}))
	})

	DescribeTable("reports a broken outline as an error",
		func(o geom.Outline, asScore bool) {
			// Arrange
			s := family.Shape{Outlines: []geom.Outline{o}}
			if asScore {
				s = family.Shape{Scores: []geom.Outline{o}}
			}

			// Act
			got := check.Solid{}.Run(s, check.Context{})

			// Assert
			Expect(got).To(ConsistOf(check.Finding{Level: check.Error, Message: "An outline didn't come out right. Please report this with your description."}))
		},
		Entry("too few points", geom.Outline{{X: 0}, {X: 1}}, false),
		Entry("not a number", geom.Outline{{X: math.NaN()}, {X: 1}, {Y: 1}}, false),
		Entry("no area", geom.Outline{{X: 0}, {X: 1}, {X: 2}}, false),
		Entry("a broken score line", geom.Outline{{X: 0}, {X: 1}}, true),
	)
})

var _ = Describe("Wall", func() {
	It("warns when the thinnest wall is under the minimum", func() {
		// Arrange
		s := family.Shape{MinWall: 0.64}

		// Act
		got := check.Wall{Min: 0.8}.Run(s, check.Context{})

		// Assert
		Expect(got).To(ConsistOf(check.Finding{Level: check.Warning, Message: "The thinnest wall is about 0.6 mm (estimated), under 0.8 mm. Thicker walls print more reliably."}))
	})

	It("is quiet at or above the minimum, or when walls weren't measured", func() {
		// Arrange
		at, flat := family.Shape{MinWall: 0.8}, family.Shape{}

		// Act
		gotAt, gotFlat := check.Wall{Min: 0.8}.Run(at, check.Context{}), check.Wall{Min: 0.8}.Run(flat, check.Context{})

		// Assert
		Expect(gotAt).To(BeEmpty())
		Expect(gotFlat).To(BeEmpty())
	})
})

var _ = Describe("Overhang", func() {
	It("warns when walls lean further than the limit", func() {
		// Arrange
		s := family.Shape{Mesh: tetra(), Overhang: 52.4}

		// Act
		got := check.Overhang{Max: 45}.Run(s, check.Context{})

		// Assert
		Expect(got).To(ConsistOf(check.Finding{Level: check.Warning, Message: "Parts of the shape lean about 52° from upright, more than 45°. They may need supports to print."}))
	})

	It("is quiet at the limit", func() {
		// Arrange
		s := family.Shape{Mesh: tetra(), Overhang: 45}

		// Act
		got := check.Overhang{Max: 45}.Run(s, check.Context{})

		// Assert
		Expect(got).To(BeEmpty())
	})
})

var _ = Describe("Blocking", func() {
	It("is true only when some finding is an error", func() {
		// Arrange
		warn := []check.Finding{{Level: check.Warning}}
		both := append(warn, check.Finding{Level: check.Error})

		// Act
		w, b, none := check.Blocking(warn), check.Blocking(both), check.Blocking(nil)

		// Assert
		Expect([]bool{w, b, none}).To(Equal([]bool{false, true, false}))
	})
})

var _ = Describe("Bridge", func() {
	It("warns about a wide flat ceiling facing down", func() {
		// Arrange
		s := family.Shape{Mesh: tetra(), Bridge: 72.4}

		// Act
		got := check.Bridge{Max: 20}.Run(s, check.Context{})

		// Assert
		Expect(got).To(ConsistOf(check.Finding{Level: check.Warning, Message: "A flat ceiling about 72 mm across faces down (under the raised foot); printers bridge more than 20 mm poorly, so it may need supports."}))
	})

	It("is quiet at the limit, and when there is no ceiling", func() {
		// Arrange
		at, none := family.Shape{Bridge: 20}, family.Shape{}

		// Act
		gotAt, gotNone := check.Bridge{Max: 20}.Run(at, check.Context{}), check.Bridge{Max: 20}.Run(none, check.Context{})

		// Assert
		Expect(gotAt).To(BeEmpty())
		Expect(gotNone).To(BeEmpty())
	})
})
