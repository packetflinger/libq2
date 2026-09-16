package bsp

import "math"

// vec3 is a float64 3-vector used for the geometric math involved in
// converting between planes (normal + distance) and the 3 arbitrary
// points a .map brush face is defined by. float64 is used regardless of
// the proto's float32 fields to keep intermediate math precise.
type vec3 struct{ x, y, z float64 }

func (a vec3) add(b vec3) vec3      { return vec3{a.x + b.x, a.y + b.y, a.z + b.z} }
func (a vec3) sub(b vec3) vec3      { return vec3{a.x - b.x, a.y - b.y, a.z - b.z} }
func (a vec3) scale(s float64) vec3 { return vec3{a.x * s, a.y * s, a.z * s} }
func (a vec3) dot(b vec3) float64   { return a.x*b.x + a.y*b.y + a.z*b.z }

func (a vec3) cross(b vec3) vec3 {
	return vec3{
		a.y*b.z - a.z*b.y,
		a.z*b.x - a.x*b.z,
		a.x*b.y - a.y*b.x,
	}
}

func (a vec3) length() float64 {
	return math.Sqrt(a.dot(a))
}

func (a vec3) normalize() vec3 {
	l := a.length()
	if l == 0 {
		return a
	}
	return a.scale(1 / l)
}

// arbitraryPerpendicular returns a vector perpendicular to n by crossing
// it with whichever world axis n is least aligned with, which keeps the
// cross product well-conditioned regardless of n's direction.
func arbitraryPerpendicular(n vec3) vec3 {
	ax, ay, az := math.Abs(n.x), math.Abs(n.y), math.Abs(n.z)
	var ref vec3
	switch {
	case ax <= ay && ax <= az:
		ref = vec3{x: 1}
	case ay <= ax && ay <= az:
		ref = vec3{y: 1}
	default:
		ref = vec3{z: 1}
	}
	return n.cross(ref)
}

// facePlanePoints synthesizes 3 non-collinear points lying on the given
// plane, ordered so that a .map-reading editor recomputing the plane from
// them (normal = cross(p2-p1, p0-p1), the standard convention used by id
// Software's qbsp and by TrenchBroom) reproduces the same normal
// direction. The points don't correspond to any real vertex of the
// original brush - the compiled .bsp format doesn't retain those - they
// only exist to pin down the plane equation; the editor derives the
// brush's actual visible geometry itself as the intersection of all of a
// brush's face planes.
func facePlanePoints(n vec3, dist float64) (p0, p1, p2 vec3) {
	const scale = 128.0

	n = n.normalize()
	base := n.scale(dist)

	u := arbitraryPerpendicular(n).normalize()
	v := n.cross(u)

	p0 = base
	p1 = base.add(u.scale(scale))
	p2 = base.add(v.scale(scale))
	return p0, p1, p2
}

// planeFromPoints computes a plane's normal and distance from 3 points on
// it, using the same formula as id Software's qbsp (normal = cross(p2-p1,
// p0-p1)) so that it's the exact inverse of facePlanePoints.
func planeFromPoints(p0, p1, p2 vec3) (normal vec3, dist float64) {
	t1 := p0.sub(p1)
	t2 := p2.sub(p1)
	normal = t2.cross(t1).normalize()
	dist = p0.dot(normal)
	return normal, dist
}
