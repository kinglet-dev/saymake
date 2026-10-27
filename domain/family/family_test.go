package family_test

import (
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/param"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fake is a test double for the Family port.
type fake struct {
	id     string
	params param.Set
	main   string
	prop   []string
}

func (f fake) ID() string                               { return f.id }
func (f fake) Label() string                            { return "Fake " + f.id }
func (f fake) Params() param.Set                        { return f.params }
func (f fake) MainSize() string                         { return f.main }
func (f fake) Proportional() []string                   { return f.prop }
func (f fake) Build(param.Values) (family.Shape, error) { return family.Shape{}, nil }

var size = param.Spec{Name: "size", Label: "Size", Unit: "mm", Min: 1, Max: 9, Default: 5, Step: 1}

var _ = Describe("Registry", func() {
	It("finds registered families by id and lists them in order", func() {
		// Arrange
		a, b := fake{id: "b", params: param.Set{size}, main: "size"}, fake{id: "a", params: param.Set{size}, main: "size"}

		// Act
		r, err := family.NewRegistry(a, b)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		got, ok := r.Get("a")
		Expect(ok).To(BeTrue())
		Expect(got.ID()).To(Equal("a"))
		Expect(r.IDs()).To(Equal([]string{"b", "a"}))
	})

	It("reports an id that isn't registered", func() {
		// Arrange
		r, err := family.NewRegistry(fake{id: "a", params: param.Set{size}, main: "size"})
		Expect(err).NotTo(HaveOccurred())

		// Act
		_, ok := r.Get("zzz")

		// Assert
		Expect(ok).To(BeFalse())
	})

	DescribeTable("refuses a family that breaks the contract",
		func(fs []family.Family, want string) {
			// Arrange
			// (families from the table)

			// Act
			_, err := family.NewRegistry(fs...)

			// Assert
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("empty id", []family.Family{fake{id: "", params: param.Set{size}, main: "size"}}, "a family has no id"),
		Entry("same id twice", []family.Family{fake{id: "a", params: param.Set{size}, main: "size"}, fake{id: "a", params: param.Set{size}, main: "size"}}, `family "a" is registered twice`),
		Entry("bad settings", []family.Family{fake{id: "a", params: param.Set{{Name: "x", Min: 2, Max: 1, Step: 1}}, main: "x"}}, `family "a": setting "x"`),
		Entry("proportional setting unknown", []family.Family{fake{id: "a", params: param.Set{size}, main: "size", prop: []string{"girth"}}}, `family "a": "girth" scales with its main size but is not one of its settings`),
		Entry("main size that can be zero", []family.Family{fake{id: "a", params: param.Set{{Name: "s", Min: 0, Max: 9, Default: 5, Step: 1}}, main: "s"}}, `family "a": its main size "s" must have a positive minimum`),
		Entry("main size not a setting", []family.Family{fake{id: "a", params: param.Set{size}, main: "girth"}}, `family "a": its main size "girth" is not one of its settings`),
	)
})
