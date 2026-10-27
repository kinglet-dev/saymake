package vessel_test

import (
	"testing"

	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVessel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vessel Suite")
}

// cylinder is a plain straight-sided cup: no taper, bulge, neck or pattern.
func cylinder() param.Values {
	v := vessel.New().Params().Defaults()
	for k, x := range map[string]float64{
		"height": 100, "diameter": 80, "base": 1, "top": 1, "bulge": 0, "neck": 0,
		"wall": 2, "floor": 3, "sides": 0, "flutes": 0, "twist": 0, "drain": 0,
		"rimStyle": 0, "rim": 1, "foot": 0, "handle": 0, "handleReach": 25, "handleThick": 8,
	} {
		v[k] = x
	}
	return v
}
