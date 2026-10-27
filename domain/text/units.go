package text

var units = map[string]float64{
	"mm": 1, "millimeter": 1, "millimeters": 1, "millimetre": 1, "millimetres": 1,
	"cm": 10, "centimeter": 10, "centimeters": 10, "centimetre": 10, "centimetres": 10,
	"m": 1000, "meter": 1000, "meters": 1000, "metre": 1000, "metres": 1000,
	"in": 25.4, "inch": 25.4, "inches": 25.4,
	"ft": 304.8, "foot": 304.8, "feet": 304.8,
}

// Millimeters returns how many millimeters one of the unit is. Only exact,
// lowercase unit words are known (inch and foot are the international inch,
// 25.4 mm exactly, and foot, 304.8 mm).
func Millimeters(unit string) (float64, bool) {
	mm, ok := units[unit]
	return mm, ok
}
