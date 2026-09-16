package bsp

import "math"

// clipEpsilon is the distance tolerance used throughout brush/tree
// construction to decide whether a point lies "on" a plane.
const clipEpsilon = 1.0 / 32.0

// bigExtent is the half-width of the huge quad facePolygonOnPlane starts
// from before clipping it down to a brush's real bounded face. It must be
// larger than any realistic map's extent.
const bigExtent = 1 << 20

// facePolygonOnPlane returns a huge quad lying exactly on the given
// plane, wound counter-clockwise as seen from the side the normal points
// to (the standard convention: for consecutive polygon vertices, the
// cross product of adjacent edges points along the normal).
func facePolygonOnPlane(n vec3, dist float64) []vec3 {
	n = n.normalize()
	base := n.scale(dist)
	u := arbitraryPerpendicular(n).normalize()
	v := n.cross(u)

	return []vec3{
		base.sub(u.scale(bigExtent)).sub(v.scale(bigExtent)),
		base.add(u.scale(bigExtent)).sub(v.scale(bigExtent)),
		base.add(u.scale(bigExtent)).add(v.scale(bigExtent)),
		base.sub(u.scale(bigExtent)).add(v.scale(bigExtent)),
	}
}

// clipPolygon clips a polygon to the half-space dot(p,n) <= dist,
// discarding the portion in front of the plane (Sutherland-Hodgman, for
// one clipping plane).
func clipPolygon(poly []vec3, n vec3, dist float64) []vec3 {
	if len(poly) == 0 {
		return nil
	}
	var out []vec3
	for i := range poly {
		cur := poly[i]
		prev := poly[(i-1+len(poly))%len(poly)]
		curDist := cur.dot(n) - dist
		prevDist := prev.dot(n) - dist
		curIn := curDist <= clipEpsilon
		prevIn := prevDist <= clipEpsilon
		if curIn != prevIn {
			t := prevDist / (prevDist - curDist)
			out = append(out, prev.add(cur.sub(prev).scale(t)))
		}
		if curIn {
			out = append(out, cur)
		}
	}
	return out
}

// pointSpread returns the minimum and maximum signed distance from a set
// of points to a plane, used to classify a polygon or brush as entirely
// in front of, entirely behind, or spanning a splitting plane.
func pointSpread(pts []vec3, n vec3, dist float64) (min, max float64) {
	min, max = math.Inf(1), math.Inf(-1)
	for _, p := range pts {
		d := p.dot(n) - dist
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}
	return min, max
}

// polygonBounds returns the axis-aligned bounding box across every point
// of every given polygon.
func polygonBounds(polys ...[]vec3) (mins, maxs vec3) {
	mins = vec3{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs = vec3{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, poly := range polys {
		for _, p := range poly {
			mins = vec3{math.Min(mins.x, p.x), math.Min(mins.y, p.y), math.Min(mins.z, p.z)}
			maxs = vec3{math.Max(maxs.x, p.x), math.Max(maxs.y, p.y), math.Max(maxs.z, p.z)}
		}
	}
	return mins, maxs
}
