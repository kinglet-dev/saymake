package interpret_test

import (
	"github.com/kinglet-dev/saymake/app/interpret"
	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/domain/vocab"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const tol = 1e-9

var _ = Describe("Interpreter.Read: objects", func() {
	var in *interpret.Interpreter

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		in, err = interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())
	})

	It("reads an object word and starts from its defaults over the family's", func() {
		// Arrange
		d := "a vase"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("vase"))
		Expect(got.Family).To(Equal("vessel"))
		Expect(got.Path).To(Equal([]string{"Vessel", "Vase"}))
		Expect(got.Values).To(Equal(param.Values{"height": 200, "diameter": 100, "flutes": 0}))
		Expect(got.Uses).To(ConsistOf(interpret.Use{Typed: "vase", Meaning: "Vase", Kind: interpret.Object}))
	})

	It("reads an object word with a typo", func() {
		// Arrange
		d := "vaze"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("vase"))
		Expect(got.Uses[0].Typed).To(Equal("vaze"))
	})

	It("reads a two-word object", func() {
		// Arrange
		d := "bud vase"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("bud-vase"))
		Expect(got.Values["height"]).To(BeNumerically("~", 120, tol))
	})

	It("prefers the more specific of two related objects", func() {
		// Arrange
		d := "vase vessel"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("vase"))
		Expect(got.Notes).To(BeEmpty())
	})

	It("uses the last of two unrelated objects and says so", func() {
		// Arrange
		d := "vase tray"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("tray"))
		Expect(got.Notes).To(ContainElement(`"vase" also names an object; making a Tray, the last one named.`))
	})

	It("lists words it didn't understand", func() {
		// Arrange
		d := "a zorgle vase with glitter"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Unknown).To(Equal([]string{"zorgle", "glitter"}))
	})

	It("refuses a description that names no object, and suggests some", func() {
		// Arrange
		d := "a tall blorp"

		// Act
		_, err := in.Read(d)

		// Assert
		Expect(err).To(MatchError(`I couldn't tell what to make. Name an object, such as vessel, vase, bud vase, bowl or tray.`))
	})

	It("passes on a description it can't read", func() {
		// Arrange
		d := "vase 123456 mm"

		// Act
		_, err := in.Read(d)

		// Assert
		Expect(err).To(MatchError(ContainSubstring("too many digits")))
	})

	It("refuses a vocabulary that doesn't fit the families", func() {
		// Arrange
		v := vocabulary()
		v.Concepts[0].Family = "lamp"

		// Act
		_, err := interpret.New(v, registry())

		// Assert
		Expect(err).To(MatchError(ContainSubstring(`family "lamp" is not registered`)))
	})
})

var _ = Describe("Interpreter.New", func() {
	It("lists at most eight object suggestions", func() {
		// Arrange
		v := vocabulary()
		for _, w := range []string{"jug", "pot", "jar", "cup", "mug"} {
			v.Concepts = append(v.Concepts, vocab.Concept{ID: w, Label: w, Parent: "vessel", Words: []string{w}})
		}
		in, err := interpret.New(v, registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		_, err = in.Read("nothing")

		// Assert
		Expect(err).To(MatchError(`I couldn't tell what to make. Name an object, such as vessel, vase, bud vase, bowl, tray, jug, pot or jar.`))
	})
})
