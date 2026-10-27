package vessel

import "math"

// The profile is a radius r(s) for s = z/height in [0, 1], built from smooth
// terms: a linear taper from base to top, a sine bulge peaking at the belly,
// and a neck pinch under a smooth window. Around the axis the radius is
// modulated by a regular polygon's polar form and by flutes,
// r·(1 + d·cos kθ), with the pattern turned by the twist as s rises.

const neckWidth = 0.15 // half-width of the neck window, as a fraction of height

func (c settings) profile(s float64) float64 {
	taper := c.base + (c.top-c.base)*s
	return c.radius * (taper + c.bulge*bump(s, c.belly) - c.neck*window(s, c.neckAt))
}

// bump rises from 0 at s = 0 to 1 at s = peak and falls back to 0 at s = 1.
func bump(s, peak float64) float64 {
	u := s / (2 * peak)
	if s > peak {
		u = 0.5 + (s-peak)/(2*(1-peak))
	}
	return math.Sin(math.Pi * u)
}

// window is 1 at s = at and falls smoothly to 0 at a distance of neckWidth.
func window(s, at float64) float64 {
	x := (s - at) / neckWidth
	if math.Abs(x) >= 1 {
		return 0
	}
	return (1 - x*x) * (1 - x*x)
}

// pattern is the factor the radius is multiplied by at angle theta and
// height s: polygon sides, flutes and twist.
func (c settings) pattern(theta, s float64) float64 {
	t := theta - c.twist*math.Pi/180*s
	m := 1.0
	if c.sides >= 3 {
		n := float64(c.sides)
		seg := 2 * math.Pi / n
		m = math.Cos(math.Pi/n) / math.Cos(math.Mod(t+8*math.Pi, seg)-math.Pi/n)
	}
	if c.flutes > 0 {
		m *= 1 + c.fluteDepth*math.Cos(float64(c.flutes)*t)
	}
	return m
}

// insetFactor is how much less the inner radius is: the wall across a flat
// side of a polygon is the wall setting, so its corners are set in further.
func (c settings) insetFactor() float64 {
	if c.sides >= 3 {
		return c.wall / math.Cos(math.Pi/float64(c.sides))
	}
	return c.wall
}

// overhang returns the steepest slope of the profile from vertical, in
// degrees, over the given rings.
func (c settings) overhang(rings int) float64 {
	steepest := 0.0
	dz := c.height / float64(rings)
	for j := range rings {
		s0, s1 := float64(j)/float64(rings), float64(j+1)/float64(rings)
		slope := math.Abs(c.profile(s1)-c.profile(s0)) / dz
		steepest = max(steepest, math.Atan(slope)*180/math.Pi)
	}
	return steepest
}
