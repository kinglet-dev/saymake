package geom

// Band joins two rings of the same length with quads, wrapping around;
// flip reverses the faces.
func (m *Mesh) Band(a, b []int, flip bool) {
	n := len(a)
	for i := range n {
		k := (i + 1) % n
		if flip {
			m.Quad(a[i], b[i], b[k], a[k])
		} else {
			m.Quad(a[i], a[k], b[k], b[i])
		}
	}
}

// Fan joins a ring to a center vertex; down makes the faces point the
// other way.
func (m *Mesh) Fan(center int, ring []int, down bool) {
	n := len(ring)
	for i := range n {
		k := (i + 1) % n
		if down {
			m.Tri(center, ring[k], ring[i])
		} else {
			m.Tri(center, ring[i], ring[k])
		}
	}
}

// Ring adds an outline's points at height z and returns their numbers.
func (m *Mesh) Ring(o Outline, z float64) []int {
	ids := make([]int, len(o))
	for i, p := range o {
		ids[i] = m.Add(Vec3{X: p.X, Y: p.Y, Z: z})
	}
	return ids
}
