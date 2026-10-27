package tray_test

import (
	"testing"

	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/families/tray"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTray(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tray Suite")
}

// box is a straight-walled rectangular tray.
func box() param.Values {
	v := tray.New().Params().Defaults()
	for k, x := range map[string]float64{
		"length": 200, "width": 120, "height": 30, "wall": 2, "floor": 3, "sides": 4, "flare": 0,
	} {
		v[k] = x
	}
	return v
}
