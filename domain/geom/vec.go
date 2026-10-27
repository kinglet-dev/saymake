// Package geom holds the plain geometry shapes are built from: 2D and 3D
// vectors, triangle meshes and closed outlines, in millimeters.
package geom

import "math"

// Vec3 is a point or direction in 3D, in millimeters.
type Vec3 struct{ X, Y, Z float64 }

// Sub returns v − w.
func (v Vec3) Sub(w Vec3) Vec3 { return Vec3{v.X - w.X, v.Y - w.Y, v.Z - w.Z} }

// Cross returns the right-handed cross product v × w.
func (v Vec3) Cross(w Vec3) Vec3 {
	return Vec3{v.Y*w.Z - v.Z*w.Y, v.Z*w.X - v.X*w.Z, v.X*w.Y - v.Y*w.X}
}

// Dot returns the dot product v · w.
func (v Vec3) Dot(w Vec3) float64 { return v.X*w.X + v.Y*w.Y + v.Z*w.Z }

// Finite reports whether no component is NaN or infinite.
func (v Vec3) Finite() bool { return finite(v.X) && finite(v.Y) && finite(v.Z) }

// Vec2 is a point or direction in the plane, in millimeters.
type Vec2 struct{ X, Y float64 }

// Add returns v + w.
func (v Vec2) Add(w Vec2) Vec2 { return Vec2{v.X + w.X, v.Y + w.Y} }

// Sub returns v − w.
func (v Vec2) Sub(w Vec2) Vec2 { return Vec2{v.X - w.X, v.Y - w.Y} }

// Scale returns v·s.
func (v Vec2) Scale(s float64) Vec2 { return Vec2{v.X * s, v.Y * s} }

// Cross returns the z component of the 3D cross product of v and w.
func (v Vec2) Cross(w Vec2) float64 { return v.X*w.Y - v.Y*w.X }

// Finite reports whether neither component is NaN or infinite.
func (v Vec2) Finite() bool { return finite(v.X) && finite(v.Y) }

func finite(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) }

// Dot returns the dot product v · w.
func (v Vec2) Dot(w Vec2) float64 { return v.X*w.X + v.Y*w.Y }
