package vessel

import "github.com/kinglet-dev/saymake/domain/param"

var specs = param.Set{
	{Name: "height", Label: "Height", Unit: "mm", Min: 20, Max: 400, Default: 150, Step: 1},
	{Name: "diameter", Label: "Diameter", Unit: "mm", Min: 20, Max: 300, Default: 100, Step: 1},
	{Name: "base", Label: "Base width", Unit: "×", Min: 0.3, Max: 1.5, Default: 0.8, Step: 0.01},
	{Name: "top", Label: "Top width", Unit: "×", Min: 0.2, Max: 2, Default: 0.9, Step: 0.01},
	{Name: "bulge", Label: "Bulge", Unit: "×", Min: 0, Max: 0.6, Default: 0.1, Step: 0.01},
	{Name: "belly", Label: "Bulge height", Unit: "×", Min: 0.2, Max: 0.8, Default: 0.4, Step: 0.01},
	{Name: "neck", Label: "Neck pinch", Unit: "×", Min: 0, Max: 0.8, Default: 0, Step: 0.01},
	{Name: "neckAt", Label: "Neck height", Unit: "×", Min: 0.6, Max: 0.95, Default: 0.8, Step: 0.01},
	{Name: "sides", Label: "Sides (0 = round)", Min: 0, Max: 12, Default: 0, Step: 1, Whole: true},
	{Name: "flutes", Label: "Flutes", Min: 0, Max: 32, Default: 0, Step: 1, Whole: true},
	{Name: "fluteDepth", Label: "Flute depth", Unit: "×", Min: 0, Max: 0.2, Default: 0.06, Step: 0.01},
	{Name: "twist", Label: "Twist", Unit: "°", Min: -360, Max: 360, Default: 0, Step: 1},
	{Name: "lean", Label: "Lean", Unit: "°", Min: 0, Max: 30, Default: 0, Step: 1},
	{Name: "wall", Label: "Wall", Unit: "mm", Min: 0.4, Max: 10, Default: 2, Step: 0.1},
	{Name: "floor", Label: "Floor", Unit: "mm", Min: 0.4, Max: 20, Default: 3, Step: 0.1},
	{Name: "drain", Label: "Drainage hole", Unit: "mm", Min: 0, Max: 40, Default: 0, Step: 1},
	{Name: "rimStyle", Label: "Rim (0 flat, 1 chamfer, 2 round)", Min: 0, Max: 2, Default: 0, Step: 1, Whole: true},
	{Name: "rim", Label: "Rim edge size", Unit: "mm", Min: 0.2, Max: 5, Default: 1, Step: 0.1},
	{Name: "slant", Label: "Rim slant", Unit: "°", Min: 0, Max: 45, Default: 0, Step: 1},
	{Name: "foot", Label: "Raised foot", Unit: "mm", Min: 0, Max: 15, Default: 0, Step: 0.5},
	{Name: "handle", Label: "Handle height (0 = none)", Unit: "×", Min: 0, Max: 0.9, Default: 0, Step: 0.01},
	{Name: "handleReach", Label: "Handle reach", Unit: "mm", Min: 10, Max: 80, Default: 25, Step: 1},
	{Name: "handleThick", Label: "Handle thickness", Unit: "mm", Min: 4, Max: 20, Default: 8, Step: 0.5},
}

// settings are a vessel's values, read once.
type settings struct {
	height, radius, base, top, bulge, belly, neck, neckAt float64
	sides, flutes, rimStyle                               int
	fluteDepth, twist, wall, floor, drain                 float64
	rim, foot, handle, handleReach, handleThick           float64
	slant, lean                                           float64
}

func read(v param.Values) (settings, error) {
	if err := specs.Require(v); err != nil {
		return settings{}, err
	}
	return settings{
		height: v["height"], radius: v["diameter"] / 2, base: v["base"], top: v["top"],
		bulge: v["bulge"], belly: v["belly"], neck: v["neck"], neckAt: v["neckAt"],
		sides: int(v["sides"]), flutes: int(v["flutes"]), fluteDepth: v["fluteDepth"],
		twist: v["twist"], wall: v["wall"], floor: v["floor"], drain: v["drain"],
		rimStyle: int(v["rimStyle"]), rim: v["rim"], foot: v["foot"],
		handle: v["handle"], handleReach: v["handleReach"], handleThick: v["handleThick"],
		slant: v["slant"], lean: v["lean"],
	}, nil
}
