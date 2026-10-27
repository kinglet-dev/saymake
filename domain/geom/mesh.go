package geom

// Mesh is a triangle mesh: vertices, and triangles as triples of vertex
// numbers, wound counterclockwise when seen from outside.
type Mesh struct {
	Vertices  []Vec3
	Triangles [][3]int
}

// Add appends a vertex and returns its number.
func (m *Mesh) Add(v Vec3) int {
	m.Vertices = append(m.Vertices, v)
	return len(m.Vertices) - 1
}

// Tri appends the triangle a, b, c.
func (m *Mesh) Tri(a, b, c int) { m.Triangles = append(m.Triangles, [3]int{a, b, c}) }

// Quad appends the quad a, b, c, d as two triangles with the same winding.
func (m *Mesh) Quad(a, b, c, d int) {
	m.Tri(a, b, c)
	m.Tri(a, c, d)
}

// Outline is a closed polygon; the last point joins the first.
type Outline []Vec2

// Area returns the signed area (shoelace formula): positive when the
// points run counterclockwise.
func (o Outline) Area() float64 {
	sum := 0.0
	for i, p := range o {
		sum += p.Cross(o[(i+1)%len(o)])
	}
	return sum / 2
}

// Finite reports whether every point is finite.
func (o Outline) Finite() bool {
	for _, p := range o {
		if !p.Finite() {
			return false
		}
	}
	return true
}

// Bounds returns the lower-left and upper-right corners of the bounding box.
// The outline must not be empty.
func (o Outline) Bounds() (lo, hi Vec2) {
	lo, hi = o[0], o[0]
	for _, p := range o[1:] {
		lo = Vec2{min(lo.X, p.X), min(lo.Y, p.Y)}
		hi = Vec2{max(hi.X, p.X), max(hi.Y, p.Y)}
	}
	return lo, hi
}
