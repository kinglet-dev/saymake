package acceptance_test

import (
	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/internal/wiring"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Slanted rims and leaning bodies, from a description it first couldn't
// read ("an oblique vase"): "oblique" alone means a slanted rim.
var _ = Describe("Slanted and leaning vessels", func() {
	var app wiring.App

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		app, err = wiring.New()
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("read and build",
		func(description, concept string, slant, lean float64) {
			// Act
			r, err := app.Interpreter.Read(description)
			Expect(err).NotTo(HaveOccurred())
			res, buildErr := app.Builder.Build(r.Family, r.Values, check.Context{})

			// Assert
			Expect(buildErr).NotTo(HaveOccurred())
			Expect(r.Concept).To(Equal(concept))
			Expect(r.Unknown).To(BeEmpty())
			Expect(r.Values["slant"]).To(Equal(slant), "slant")
			Expect(r.Values["lean"]).To(Equal(lean), "lean")
			Expect(res.Findings).NotTo(ContainElement(HaveField("Message", ContainSubstring("tip over"))))
		},
		Entry(nil, "an oblique vase", "vase", 25.0, 0.0),
		Entry(nil, "a vase with an oblique cut", "vase", 25.0, 0.0),
		Entry(nil, "vase with a slanted rim", "vase", 25.0, 0.0),
		Entry(nil, "a cup with an angled top", "cup", 25.0, 0.0),
		Entry(nil, "a vase cut at an angle", "vase", 25.0, 0.0),
		Entry(nil, "a leaning vase", "vase", 0.0, 12.0),
		Entry(nil, "a tilted jar", "jar", 0.0, 12.0),
		Entry(nil, "a leaning oblique vase", "vase", 25.0, 12.0),
		Entry(nil, "a clean vase", "vase", 0.0, 0.0),
	)

	It("keeps \"slanted\" on its own for flared walls", func() {
		// Act
		r, err := app.Interpreter.Read("a vase with slanted walls")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Values["slant"]).To(BeZero())
		Expect(r.Values["top"]).To(BeNumerically(">", 1))
	})

	It("doesn't read a tile as tilted", func() {
		// Act
		r, err := app.Interpreter.Read("a square tile")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Concept).To(Equal("plaque"))
		Expect(r.Values["lean"]).To(BeZero())
	})

	It("warns that a tall, narrow vase leaning hard may tip over", func() {
		// Arrange
		r, err := app.Interpreter.Read("a tall narrow vase")
		Expect(err).NotTo(HaveOccurred())
		r.Values["lean"] = 30

		// Act
		res, buildErr := app.Builder.Build(r.Family, r.Values, check.Context{})

		// Assert
		Expect(buildErr).NotTo(HaveOccurred())
		Expect(res.Findings).To(ContainElement(check.Finding{Level: check.Warning,
			Message: "It may tip over when empty: its center of weight isn't above what it stands on. Make the base wider, or make it lean less."}))
	})
})
