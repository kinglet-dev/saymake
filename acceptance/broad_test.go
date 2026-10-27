package acceptance_test

import (
	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/internal/wiring"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// A wider set of descriptions in everyday phrasing. They were written with
// the vocabulary in view, so they test its coverage, not how well it reads
// people's own words; descriptions it got wrong in use are in
// screenshots_test.go and oblique_test.go.
var _ = Describe("Everyday descriptions", func() {
	var app wiring.App

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		app, err = wiring.New()
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("land on a buildable object",
		func(description, concept string, values map[string]float64, note string) {
			// Arrange
			d := description

			// Act
			r, err := app.Interpreter.Read(d)
			Expect(err).NotTo(HaveOccurred())
			res, buildErr := app.Builder.Build(r.Family, r.Values, check.Context{Printer: [3]float64{256, 256, 256}})

			// Assert
			Expect(r.Concept).To(Equal(concept))
			for k, v := range values {
				Expect(r.Values[k]).To(BeNumerically("~", v, tol), k)
			}
			if note == "" {
				Expect(r.Unknown).To(BeEmpty())
			} else {
				Expect(r.Notes).To(ContainElement(ContainSubstring(note)))
			}
			Expect(buildErr).NotTo(HaveOccurred())
			Expect(res.Formats).NotTo(BeEmpty(), "%v", res.Findings)
		},
		Entry(nil, "a coffee mug", "mug", map[string]float64{"handle": 0.6}, ""),
		Entry(nil, "mug with a thin handle, 9cm tall", "mug", map[string]float64{"handleThick": 6, "height": 90}, ""),
		Entry(nil, "espresso cup with a handle", "teacup", map[string]float64{"handle": 0.55}, ""),
		Entry(nil, "a water jug with a chunky handle", "pitcher", map[string]float64{"handleThick": 12}, ""),
		Entry(nil, "handleless teacup", "teacup", map[string]float64{"handle": 0}, ""),
		Entry(nil, "a ramekin", "small-bowl", nil, ""),
		Entry(nil, "dog bowl, 18 cm across, raised foot", "bowl", map[string]float64{"diameter": 180, "foot": 3}, ""),
		Entry(nil, "salad bowl with a rounded rim", "bowl", map[string]float64{"rimStyle": 2}, ""),
		Entry(nil, "cylindrical pen holder with a beveled edge", "pencil-cup", map[string]float64{"rimStyle": 1, "top": 1}, ""),
		Entry(nil, "toothbrush holder, 12 cm", "pencil-cup", map[string]float64{"height": 120}, ""),
		Entry(nil, "votive candle holder, hexagon", "candle-holder", map[string]float64{"sides": 6}, ""),
		Entry(nil, "orchid pot with drainage and a foot ring", "planter", map[string]float64{"drain": 10, "foot": 3}, ""),
		Entry(nil, "squat wide planter", "planter", nil, ""),
		Entry(nil, "cookie jar 20 cm", "jar", map[string]float64{"height": 200}, ""),
		Entry(nil, "tall skinny bottle", "bottle", nil, ""),
		Entry(nil, "a bulbous vase with a narrow neck and 12 ribs", "vase", map[string]float64{"flutes": 12, "neck": 0.35}, ""),
		Entry(nil, "petal vase", "vase", map[string]float64{"flutes": 8}, ""),
		Entry(nil, "octagonal vase with straight sides", "vase", map[string]float64{"sides": 8, "top": 1}, ""),
		Entry(nil, "3d printed spiral vase", "vase", map[string]float64{"twist": 120}, ""),
		Entry(nil, "jewelry dish with sloped sides", "valet-tray", map[string]float64{"flare": 20}, ""),
		Entry(nil, "key tray, rectangular with rounded corners", "valet-tray", map[string]float64{"sides": 4, "corner": 6}, ""),
		Entry(nil, "large serving tray with straight walls", "serving-tray", map[string]float64{"flare": 0}, ""),
		Entry(nil, "desk organizer 15cm long, 10 cm wide", "box", map[string]float64{"length": 150, "width": 100}, ""),
		Entry(nil, "drawer tray with 3 mm thick walls", "box", map[string]float64{"wall": 3}, ""),
		Entry(nil, "soap dish", "dish", nil, ""),
		Entry(nil, "hexagon coaster with a border", "coaster", map[string]float64{"sides": 6, "border": 5}, ""),
		Entry(nil, "square coaster with 8 mm corners", "coaster", map[string]float64{"sides": 4, "corner": 8}, ""),
		Entry(nil, "door sign 25cm", "plaque", map[string]float64{"length": 250}, ""),
		Entry(nil, "luggage tag", "tag", nil, ""),
		Entry(nil, "a bookmark", "bookmark", nil, ""),
		Entry(nil, "pet tag with a hole", "tag", map[string]float64{"hole": 5}, ""),
		Entry(nil, "a vase with a lid", "vase", nil, "Lids aren't supported yet."),
		Entry(nil, "red heart shaped coaster", "coaster", nil, "Free-form shapes"),
	)

	It("explains an object it can't make, instead of a generic refusal", func() {
		// Arrange
		d := "a wine glass"

		// Act
		_, err := app.Interpreter.Read(d)

		// Assert
		Expect(err).To(MatchError(HavePrefix("Stems (goblets, wine glasses) aren't supported yet.")))
	})
})
