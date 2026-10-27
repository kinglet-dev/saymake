package acceptance_test

import (
	"bytes"
	"github.com/kinglet-dev/saymake/domain/check"

	"github.com/kinglet-dev/saymake/internal/wiring"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const tol = 1e-6

var _ = Describe("Describing objects", func() {
	var app wiring.App

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		app, err = wiring.New()
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("lands on the object, with the asked-for size and features, and makes a file",
		func(description, concept string, want map[string]float64) {
			// Arrange
			d := description

			// Act
			reading, err := app.Interpreter.Read(d)
			Expect(err).NotTo(HaveOccurred())
			result, buildErr := app.Builder.Build(reading.Family, reading.Values, check.Context{})

			// Assert
			Expect(reading.Concept).To(Equal(concept))
			for k, v := range want {
				Expect(reading.Values[k]).To(BeNumerically("~", v, tol), k)
			}
			Expect(reading.Unknown).To(BeEmpty())
			Expect(buildErr).NotTo(HaveOccurred())
			Expect(result.Formats).NotTo(BeEmpty(), "%v", result.Findings)
			var buf bytes.Buffer
			_, err = app.Builder.Export(reading.Family, reading.Values, check.Context{}, result.Formats[0].ID, &buf)
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.Len()).To(BeNumerically(">", 100))
		},
		Entry(nil, "a tall fluted vase, 20 cm", "vase", map[string]float64{"height": 200, "flutes": 16}),
		Entry(nil, "small bud vase with a narrow neck", "bud-vase", map[string]float64{"neck": 0.35, "diameter": 42}),
		Entry(nil, "tiny slim bottle with a long neck", "bottle", map[string]float64{"neck": 0.35}),
		Entry(nil, "small narrow tapered vase, thin walled", "vase", map[string]float64{"top": 0.6}),
		Entry(nil, "hexagonal planter 12cm wide with drainage", "planter", map[string]float64{"sides": 6, "diameter": 120, "drain": 10}),
		Entry(nil, "shallow oval serving tray", "serving-tray", map[string]float64{"sides": 0, "roundness": 2, "width": 196, "height": 22}),
		Entry(nil, "round coaster 9 cm", "coaster", map[string]float64{"length": 90, "width": 90}),
		Entry(nil, "rectangular plaque with a border, 30 cm long", "plaque", map[string]float64{"sides": 4, "border": 5, "length": 300}),
		Entry(nil, "gift tag with a hanging hole", "tag", map[string]float64{"hole": 5}),
		Entry(nil, "twisted pencil holder", "pencil-cup", map[string]float64{"twist": 120}),
		Entry(nil, "deep ring dish", "valet-tray", map[string]float64{"height": 35.2}),
		Entry(nil, "a 6 sided bowl", "bowl", map[string]float64{"sides": 6}),
		Entry(nil, "bowl with 24 ribs", "bowl", map[string]float64{"flutes": 24}),
		Entry(nil, "vaze with fluting", "vase", map[string]float64{"flutes": 16}),
		Entry(nil, "planter with no drainage", "planter", map[string]float64{"drain": 0}),
		Entry(nil, `spiral vase 8" tall`, "vase", map[string]float64{"height": 203.2, "twist": 120}),
		Entry(nil, "square tray", "tray", map[string]float64{"sides": 4}),
		Entry(nil, "keychain 5 cm", "tag", map[string]float64{"length": 50, "width": 75}),
		Entry(nil, "big jar", "jar", map[string]float64{"height": 168, "diameter": 140}),
		Entry(nil, "mini hexagonal succulent pot", "planter", map[string]float64{"sides": 6, "drain": 10}),
		Entry(nil, "large deep rectangular box 30cm long", "box", map[string]float64{"sides": 4, "length": 300}),
	)

	It("says what it can make when it doesn't recognize the object", func() {
		// Arrange
		d := "a lamp"

		// Act
		_, err := app.Interpreter.Read(d)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("I couldn't tell what to make.")))
	})
})
