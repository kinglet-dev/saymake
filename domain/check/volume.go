package check

import (
	"fmt"
	"strconv"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/meshcheck"
)

// BuildVolume warns when a solid doesn't fit the printer, as it is or
// turned a quarter on the bed. Flat shapes, and requests without a printer
// size, aren't checked.
type BuildVolume struct{}

// Run implements Check.
func (BuildVolume) Run(s family.Shape, ctx Context) []Finding {
	p := ctx.Printer
	if s.Mesh == nil || p[0] <= 0 {
		return nil
	}
	r := meshcheck.Inspect(*s.Mesh)
	x, y, z := r.Max.X-r.Min.X, r.Max.Y-r.Min.Y, r.Max.Z-r.Min.Z
	fits := z <= p[2] && (x <= p[0] && y <= p[1] || x <= p[1] && y <= p[0])
	if fits {
		return nil
	}
	return []Finding{{Warning, fmt.Sprintf("At %s × %s × %s mm, it doesn't fit a %s × %s × %s mm printer, even turned. Make it smaller, or change the printer size.",
		mm(x), mm(y), mm(z), mm(p[0]), mm(p[1]), mm(p[2]))}}
}

func mm(v float64) string { return strconv.FormatFloat(float64(int64(v*10+0.5))/10, 'f', -1, 64) }
