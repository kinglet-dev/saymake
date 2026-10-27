package vocab_test

import (
	"testing"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/param"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVocab(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vocab Suite")
}

// fakeFamily is a test double for the Family port.
type fakeFamily struct {
	id     string
	params param.Set
	prop   []string
}

func (f fakeFamily) ID() string                               { return f.id }
func (f fakeFamily) Label() string                            { return f.id }
func (f fakeFamily) Params() param.Set                        { return f.params }
func (f fakeFamily) MainSize() string                         { return f.params[0].Name }
func (f fakeFamily) Proportional() []string                   { return f.prop }
func (f fakeFamily) Build(param.Values) (family.Shape, error) { return family.Shape{}, nil }

func registry() *family.Registry {
	r, err := family.NewRegistry(
		fakeFamily{id: "vessel", params: param.Set{
			{Name: "height", Label: "Height", Unit: "mm", Min: 10, Max: 400, Default: 150, Step: 1},
			{Name: "flutes", Label: "Flutes", Min: 0, Max: 48, Default: 0, Step: 1, Whole: true},
		}},
		fakeFamily{id: "tray", params: param.Set{
			{Name: "length", Label: "Length", Unit: "mm", Min: 10, Max: 400, Default: 200, Step: 1},
		}},
	)
	Expect(err).NotTo(HaveOccurred())
	return r
}
