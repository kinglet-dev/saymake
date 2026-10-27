package build_test

import (
	"errors"
	"io"
	"testing"

	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	"github.com/kinglet-dev/saymake/domain/param"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBuild(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Build Suite")
}

// fakeFamily is a test double for the Family port: it builds a shape with a
// minimum wall equal to its "size" setting, or fails when size is 13.
type fakeFamily struct{}

func (fakeFamily) ID() string             { return "box" }
func (fakeFamily) Label() string          { return "Box" }
func (fakeFamily) MainSize() string       { return "size" }
func (fakeFamily) Proportional() []string { return nil }
func (fakeFamily) Params() param.Set {
	return param.Set{{Name: "size", Label: "Size", Unit: "mm", Min: 1, Max: 20, Default: 10, Step: 1}}
}
func (fakeFamily) Build(v param.Values) (family.Shape, error) {
	if v["size"] == 13 {
		return family.Shape{}, errors.New("Unlucky size. Pick another.")
	}
	return family.Shape{Mesh: &geom.Mesh{}, MinWall: v["size"]}, nil
}

// fakeCheck is a test double for the Check port: it reports a finding of
// its level when the wall is under its limit.
type fakeCheck struct {
	level check.Level
	under float64
}

func (c fakeCheck) Run(s family.Shape, ctx check.Context) []check.Finding {
	if ctx.Printer[0] == 1 {
		return []check.Finding{{Level: check.Warning, Message: "context seen"}}
	}
	if s.MinWall < c.under {
		return []check.Finding{{Level: c.level, Message: string(c.level) + " found"}}
	}
	return nil
}

// fakeExporter is a test double for the Exporter port.
type fakeExporter struct {
	id       string
	supports bool
	fail     error
}

func (e fakeExporter) ID() string                 { return e.id }
func (e fakeExporter) Label() string              { return "Format " + e.id }
func (e fakeExporter) Extension() string          { return "." + e.id }
func (e fakeExporter) MediaType() string          { return "x/" + e.id }
func (e fakeExporter) Supports(family.Shape) bool { return e.supports }
func (e fakeExporter) Write(w io.Writer, s family.Shape) error {
	if e.fail != nil {
		return e.fail
	}
	_, err := io.WriteString(w, e.id+" data")
	return err
}

func registry() *family.Registry {
	r, err := family.NewRegistry(fakeFamily{})
	Expect(err).NotTo(HaveOccurred())
	return r
}
