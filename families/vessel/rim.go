package vessel

import "math"

// roundSteps is how many rings draw a rounded rim's quarter circle.
const roundSteps = 6

// level is a ring's radius (before the pattern) and height.
type level struct{ r, z float64 }

// rimTop is the height of the rim at angle t. A slanted rim is cut by a
// plane through the top at angle 0 (the handle's side), tilted by the
// slant: exact for a round, straight top, and a smooth close fit to a plane
// where the top is patterned or curved.
func (c settings) rimTop(t float64) float64 {
	return c.height - math.Tan(c.slant*math.Pi/180)*c.profile(1)*(1-math.Cos(t))
}

// rimEdge returns the rings that take the outer wall from its rim edge up
// to the rim at the given height: none for a flat rim, one for a 45°
// chamfer, a quarter circle for a rounded one.
func (c settings) rimEdge(height float64) []level {
	top := height - c.rimDrop()
	r := c.profile(top / c.height)
	switch c.rimStyle {
	case 1:
		return []level{{r - c.rim, height}}
	case 2:
		var out []level
		for k := 1; k <= roundSteps; k++ {
			a := math.Pi / 2 * float64(k) / roundSteps
			out = append(out, level{r - c.rim + c.rim*math.Cos(a), top + c.rim*math.Sin(a)})
		}
		return out
	}
	return nil
}
