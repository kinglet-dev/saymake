package param_test

import (
	"math"

	"github.com/kinglet-dev/saymake/domain/param"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Set.Apply", func() {
	set := param.Set{height, flutes}

	It("changes the named values and keeps the rest", func() {
		// Arrange
		base := param.Values{"height": 150, "flutes": 0}

		// Act
		got, notes, err := set.Apply(base, param.Values{"flutes": 8})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(param.Values{"height": 150, "flutes": 8}))
		Expect(notes).To(BeEmpty())
		Expect(base["flutes"]).To(BeNumerically("==", 0), "the base values are not changed")
	})

	It("clamps out-of-range values and says so", func() {
		// Arrange
		base := set.Defaults()

		// Act
		got, notes, err := set.Apply(base, param.Values{"height": 2000})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got["height"]).To(BeNumerically("~", 400, tol))
		Expect(notes).To(ConsistOf("Height 2000 mm is more than the largest allowed, 400 mm; using 400 mm."))
	})

	It("says when a value was raised to the minimum", func() {
		// Arrange
		base := set.Defaults()

		// Act
		_, notes, err := set.Apply(base, param.Values{"height": 3})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(notes).To(ConsistOf("Height 3 mm is less than the smallest allowed, 20 mm; using 20 mm."))
	})

	It("refuses an unknown parameter", func() {
		// Arrange
		base := set.Defaults()

		// Act
		_, _, err := set.Apply(base, param.Values{"spout": 3})

		// Assert
		Expect(err).To(MatchError(`unknown setting "spout"`))
	})

	It("refuses a value that isn't a finite number", func() {
		// Arrange
		base := set.Defaults()

		// Act
		_, _, err := set.Apply(base, param.Values{"height": math.NaN()})

		// Assert
		Expect(err).To(MatchError("Height must be a finite number"))
	})
})

var _ = Describe("Set.Apply with several unknown settings", func() {
	It("names the first one alphabetically, every time", func() {
		// Arrange
		set := param.Set{height}
		changes := param.Values{"zeta": 1, "beta": 1, "alpha": 1, "omega": 1, "delta": 1}

		// Act
		var msgs []string
		for range 20 {
			_, _, err := set.Apply(set.Defaults(), changes)
			msgs = append(msgs, err.Error())
		}

		// Assert
		for _, m := range msgs {
			Expect(m).To(Equal(`unknown setting "alpha"`))
		}
	})
})

var _ = Describe("Spec.Format", func() {
	It("writes a bare number for a setting without a unit", func() {
		// Arrange
		s := flutes

		// Act
		got := s.Format(8)

		// Assert
		Expect(got).To(Equal("8"))
	})

	It("writes the unit after the number", func() {
		// Arrange
		s := height

		// Act
		got := s.Format(182.5)

		// Assert
		Expect(got).To(Equal("182.5 mm"))
	})
})

var _ = Describe("Set.Require", func() {
	It("accepts values that include every setting", func() {
		// Arrange
		set := param.Set{height, flutes}

		// Act
		err := set.Require(param.Values{"height": 1, "flutes": 2, "extra": 3})

		// Assert
		Expect(err).NotTo(HaveOccurred())
	})

	It("names the first missing setting in display order", func() {
		// Arrange
		set := param.Set{height, flutes}

		// Act
		err := set.Require(param.Values{})

		// Assert
		Expect(err).To(MatchError(`setting "height" is missing`))
	})
})
