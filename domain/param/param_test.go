package param_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/param"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const tol = 1e-9

var height = param.Spec{Name: "height", Label: "Height", Unit: "mm", Min: 20, Max: 400, Default: 150, Step: 1}
var flutes = param.Spec{Name: "flutes", Label: "Flutes", Min: 0, Max: 48, Default: 0, Step: 1, Whole: true}

var _ = Describe("Spec.Clamp", func() {
	It("keeps a value inside the range as it is", func() {
		// Arrange
		v := 180.5

		// Act
		got, changed := height.Clamp(v)

		// Assert
		Expect(got).To(BeNumerically("~", 180.5, tol))
		Expect(changed).To(BeFalse())
	})

	It("raises a value below the range to the minimum", func() {
		// Arrange
		v := 5.0

		// Act
		got, changed := height.Clamp(v)

		// Assert
		Expect(got).To(BeNumerically("~", 20, tol))
		Expect(changed).To(BeTrue())
	})

	It("lowers a value above the range to the maximum", func() {
		// Arrange
		v := 1000.0

		// Act
		got, changed := height.Clamp(v)

		// Assert
		Expect(got).To(BeNumerically("~", 400, tol))
		Expect(changed).To(BeTrue())
	})

	It("rounds a whole-number parameter, without calling it a change", func() {
		// Arrange
		v := 7.6

		// Act
		got, changed := flutes.Clamp(v)

		// Assert
		Expect(got).To(BeNumerically("~", 8, tol))
		Expect(changed).To(BeFalse())
	})
})

var _ = Describe("Set", func() {
	set := param.Set{height, flutes}

	It("finds a spec by name", func() {
		// Arrange
		name := "flutes"

		// Act
		s, ok := set.Find(name)

		// Assert
		Expect(ok).To(BeTrue())
		Expect(s.Label).To(Equal("Flutes"))
	})

	It("reports a missing name", func() {
		// Arrange
		name := "spout"

		// Act
		_, ok := set.Find(name)

		// Assert
		Expect(ok).To(BeFalse())
	})

	It("lists every default", func() {
		// Arrange
		s := set

		// Act
		d := s.Defaults()

		// Assert
		Expect(d).To(Equal(param.Values{"height": 150, "flutes": 0}))
	})

	DescribeTable("refuses a malformed set",
		func(bad param.Set, want string) {
			// Arrange
			s := bad

			// Act
			err := s.Validate()

			// Assert
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("empty name", param.Set{{Name: "", Min: 0, Max: 1, Step: 1}}, "has no name"),
		Entry("repeated name", param.Set{height, height}, `"height" is listed twice`),
		Entry("minimum above maximum", param.Set{{Name: "a", Min: 2, Max: 1, Step: 1, Default: 1}}, `"a": the range 2 to 1 is empty`),
		Entry("default outside range", param.Set{{Name: "a", Min: 0, Max: 1, Step: 1, Default: 2}}, `"a": default 2 is outside 0 to 1`),
		Entry("non-positive step", param.Set{{Name: "a", Min: 0, Max: 1, Step: 0}}, `"a": step must be positive`),
		Entry("non-finite bound", param.Set{{Name: "a", Min: math.Inf(-1), Max: 1, Step: 1}}, `"a": its numbers must be finite`),
		Entry("non-finite step", param.Set{{Name: "a", Min: 0, Max: 1, Step: math.NaN()}}, `"a": its numbers must be finite`),
	)

	It("accepts a well-formed set", func() {
		// Arrange
		s := set

		// Act
		err := s.Validate()

		// Assert
		Expect(err).NotTo(HaveOccurred())
	})
})
