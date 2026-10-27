package interpret_test

import (
	"testing"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/domain/vocab"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInterpret(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interpret Suite")
}

// fakeFamily is a test double for the Family port.
type fakeFamily struct {
	id, main string
	params   param.Set
	prop     []string
}

func (f fakeFamily) ID() string                               { return f.id }
func (f fakeFamily) Label() string                            { return f.id }
func (f fakeFamily) Params() param.Set                        { return f.params }
func (f fakeFamily) MainSize() string                         { return f.main }
func (f fakeFamily) Proportional() []string                   { return f.prop }
func (f fakeFamily) Build(param.Values) (family.Shape, error) { return family.Shape{}, nil }

func mm(name, label string, lo, hi, def float64) param.Spec {
	return param.Spec{Name: name, Label: label, Unit: "mm", Min: lo, Max: hi, Default: def, Step: 1}
}

func registry() *family.Registry {
	r, err := family.NewRegistry(
		fakeFamily{id: "vessel", main: "height", params: param.Set{
			mm("height", "Height", 20, 400, 150),
			mm("diameter", "Diameter", 20, 300, 100),
			{Name: "flutes", Label: "Flutes", Min: 0, Max: 48, Default: 0, Step: 1, Whole: true},
		}},
		fakeFamily{id: "tray", main: "length", prop: []string{"width"}, params: param.Set{
			mm("length", "Length", 20, 500, 250),
			mm("width", "Width", 20, 500, 180),
		}},
	)
	Expect(err).NotTo(HaveOccurred())
	return r
}

func vocabulary() vocab.Vocabulary {
	return vocab.Vocabulary{
		Concepts: []vocab.Concept{
			{ID: "vessel", Label: "Vessel", Family: "vessel", Words: []string{"vessel"}},
			{ID: "vase", Label: "Vase", Parent: "vessel", Words: []string{"vase", "urn"}, Defaults: param.Values{"height": 200}},
			{ID: "bud-vase", Label: "Bud vase", Parent: "vase", Words: []string{"bud vase"}, Defaults: param.Values{"height": 120}},
			{ID: "bowl", Label: "Bowl", Parent: "vessel", Words: []string{"bowl"}, Defaults: param.Values{"height": 80, "diameter": 180}},
			{ID: "tray", Label: "Tray", Family: "tray", Words: []string{"tray"}},
		},
		Modifiers: []vocab.Modifier{
			{ID: "fluted", Label: "fluted", Words: []string{"fluted", "ribbed", "flutes"}, Families: []string{"vessel"},
				Ops: []vocab.Op{{Param: "flutes", Kind: vocab.Set, Value: 12}}},
			{ID: "tall", Label: "tall", Words: []string{"tall"}, Ops: []vocab.Op{{Param: "height", Kind: vocab.Scale, Value: 1.4}}},
			{ID: "wider", Label: "wider", Words: []string{"wider", "wide"}, Ops: []vocab.Op{
				{Param: "diameter", Kind: vocab.Add, Value: 30}, {Param: "width", Kind: vocab.Add, Value: 30}}},
			{ID: "narrow", Label: "narrow", Words: []string{"thin"}, Ops: []vocab.Op{{Param: "diameter", Kind: vocab.Scale, Value: 0.7}}},
			{ID: "thick", Label: "thick walls", Words: []string{"thick"}, Ops: []vocab.Op{{Param: "diameter", Kind: vocab.Scale, Value: 1.2}}},
			{ID: "thick-rim", Label: "thick rim", Words: []string{"thick rim"}, Ops: []vocab.Op{{Param: "diameter", Kind: vocab.Scale, Value: 1.1}}},
			{ID: "thin-rim", Label: "thin rim", Words: []string{"thin rim"}, Ops: []vocab.Op{{Param: "diameter", Kind: vocab.Scale, Value: 0.9}}},
			{ID: "big-wide-thick", Label: "big wide thick", Words: []string{"big wide thick"}, Ops: []vocab.Op{{Param: "height", Kind: vocab.Set, Value: 33}}},
			{ID: "big-wide-thin", Label: "big wide thin", Words: []string{"big wide thin"}, Ops: []vocab.Op{{Param: "height", Kind: vocab.Set, Value: 44}}},
			{ID: "big", Label: "big", Words: []string{"big"}, Ops: []vocab.Op{{Param: "height", Kind: vocab.Set, Value: 300}}},
			{ID: "rim", Label: "rim", Words: []string{"rim"}, Ops: []vocab.Op{{Param: "flutes", Kind: vocab.Set, Value: 2}}},
			{ID: "flute-count", Label: "flutes", Words: []string{"flutes", "grooves"}, Families: []string{"vessel"}, Count: "flutes"},
		},
		Sizes: map[string]map[string]string{
			"tall": {"vessel": "height"}, "height": {"vessel": "height"},
			"wide": {"vessel": "diameter"}, "long": {"tray": "length"}, "base": {"vessel": "height"},
		},
		Ignore:      []string{"a", "with", "of", "and", "the"},
		Unsupported: []vocab.Unsupported{{Words: []string{"text", "word", "engraved name"}, Message: "Text isn't supported yet."}},
	}
}
