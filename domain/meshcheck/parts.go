package meshcheck

import "github.com/kinglet-dev/saymake/domain/geom"

// parts groups triangles into connected parts (triangles sharing a vertex,
// by union-find with path halving) and counts the parts whose own volume
// isn't positive: an inside-out part would otherwise hide inside a whole
// whose other parts outweigh it. Indices must be valid.
func parts(m geom.Mesh) (count, insideOut int) {
	parent := make([]int, len(m.Vertices))
	for i := range parent {
		parent[i] = i
	}
	find := func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	for _, t := range m.Triangles {
		a := find(t[0])
		parent[find(t[1])], parent[find(t[2])] = a, a
	}
	volume := map[int]float64{}
	for _, t := range m.Triangles {
		a, b, c := m.Vertices[t[0]], m.Vertices[t[1]], m.Vertices[t[2]]
		volume[find(t[0])] += a.Dot(b.Cross(c)) / 6
	}
	for _, v := range volume {
		if v <= 0 {
			insideOut++
		}
	}
	return len(volume), insideOut
}
