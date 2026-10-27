package vessel

import (
	"errors"
	"fmt"
	"math"
)

// check refuses shapes that can't be made, saying how to fix them.
func (c settings) check() error {
	if c.floor > c.height-1 {
		return errors.New("The floor is too thick for the height. Make the floor thinner or the vessel taller.")
	}
	narrowest := math.Inf(1)
	for j := range rings + 1 {
		narrowest = min(narrowest, c.profile(float64(j)/rings))
	}
	if inner := (narrowest - c.insetFactor()) * c.closest(); inner < 1 {
		return fmt.Errorf("The wall is too thick for the narrowest part (%.1f mm across). Make the wall thinner, or the neck or base wider.", 2*narrowest)
	}
	floorRadius := (c.profile(c.floor/c.height) - c.insetFactor()) * c.closest()
	if c.drain > 0 && c.drain/2 > floorRadius-1 {
		return errors.New("The drainage hole is too wide for the floor. Make the hole smaller or the base wider.")
	}
	for _, check := range []func() error{c.checkRim, c.checkFoot, c.checkHandle} {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

// checkRim needs a slanted rim's low side to stay above the floor, and the
// rim's top to stay outside the inner wall all the way round (the profile
// differs between a slanted rim's high and low sides).
func (c settings) checkRim() error {
	low := c.rimTop(math.Pi)
	if low-c.rimDrop() < c.floor+1 {
		return errors.New("The rim slants too steeply for this vessel: its low side would reach the floor. Make the slant smaller or the vessel taller.")
	}
	for k := range rings + 1 {
		h := low + (c.height-low)*float64(k)/rings
		if c.rimStyle > 0 && c.profile((h-c.rimDrop())/c.height)-c.rim <= c.profile(h/c.height)-c.insetFactor()+0.1 {
			return errors.New("The rim's edge is as wide as the wall. Make the rim smaller or the wall thicker.")
		}
	}
	return nil
}

func (c settings) checkFoot() error {
	switch {
	case c.foot <= 0:
		return nil
	case c.floor < c.foot+1:
		return errors.New("The floor must be at least 1 mm thicker than the foot is tall. Make the floor thicker or the foot lower.")
	case c.footRadius() < 2:
		return errors.New("The base is too narrow for a foot ring. Make the base wider or remove the foot.")
	case c.drain/2 > c.footRadius()-1:
		return errors.New("The drainage hole is too wide for the foot. Make the hole smaller or remove the foot.")
	}
	return nil
}

func (c settings) checkHandle() error {
	bottom, top := c.handleSpan()
	a := c.handleThick / 2
	if c.handle > 0 && (top+a > c.height-c.rimDrop()-1 || bottom-a < c.foot+1) {
		return errors.New("The handle is too tall for this vessel. Make it shorter or the vessel taller.")
	}
	return nil
}

// closest is the smallest the pattern makes the radius: a polygon's sides
// are cos(π/n) as far as its corners, and flutes dip by their depth.
func (c settings) closest() float64 {
	f := 1.0
	if c.sides >= 3 {
		f = math.Cos(math.Pi / float64(c.sides))
	}
	if c.flutes > 0 {
		f *= 1 - c.fluteDepth
	}
	return f
}
