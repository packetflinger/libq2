package bsp

import (
	"fmt"
	"math"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// planeInterner deduplicates (normal, distance) pairs into a single
// global plane table, matching real .bsp files where many brush sides
// across the map commonly share an identical plane (e.g. a floor plane
// reused by every brush that sits on it).
type planeInterner struct {
	list  []*bpb.BSPPlane
	norms []vec3 // normal for list[i], kept as float64 for classification math
	index map[[4]int32]int32
}

func newPlaneInterner() *planeInterner {
	return &planeInterner{index: map[[4]int32]int32{}}
}

func (pi *planeInterner) intern(n vec3, dist float64) int32 {
	key := [4]int32{quantize(n.x), quantize(n.y), quantize(n.z), quantize(dist)}
	if idx, ok := pi.index[key]; ok {
		return idx
	}
	idx := int32(len(pi.list))
	pi.list = append(pi.list, &bpb.BSPPlane{
		Normal:   &bpb.Vector3{X: float32(n.x), Y: float32(n.y), Z: float32(n.z)},
		Distance: float32(dist),
		Type:     planeType(n),
	})
	pi.norms = append(pi.norms, n)
	pi.index[key] = idx
	return idx
}

func (pi *planeInterner) normal(idx int32) vec3 {
	return pi.norms[idx]
}

func (pi *planeInterner) dist(idx int32) float64 {
	return float64(pi.list[idx].GetDistance())
}

// planeType classifies a plane the way Quake 2 does: PLANE_X/Y/Z (0-2)
// for axial planes, PLANE_ANYX/Y/Z (3-5) for the nearest non-axial family,
// used by the engine to speed up common box-vs-plane tests.
func planeType(n vec3) int32 {
	const axialThreshold = 0.9999
	ax, ay, az := math.Abs(n.x), math.Abs(n.y), math.Abs(n.z)
	switch {
	case ax > axialThreshold:
		return 0
	case ay > axialThreshold:
		return 1
	case az > axialThreshold:
		return 2
	case ax >= ay && ax >= az:
		return 3
	case ay >= ax && ay >= az:
		return 4
	default:
		return 5
	}
}

// texinfoInterner deduplicates texture alignment info, keyed on every
// field that would make two faces' alignment meaningfully different.
type texinfoInterner struct {
	list  []*bpb.BSPTexInfo
	index map[texinfoKey]int32
}

type texinfoKey struct {
	sx, sy, sz, so int32
	tx, ty, tz, to int32
	flags, value   int32
	texture        string
}

func newTexinfoInterner() *texinfoInterner {
	return &texinfoInterner{index: map[texinfoKey]int32{}}
}

func (ti *texinfoInterner) intern(f *MapBrushFace) int32 {
	key := texinfoKey{
		sx: quantize(f.SAxis.x), sy: quantize(f.SAxis.y), sz: quantize(f.SAxis.z), so: quantize(f.SOffset),
		tx: quantize(f.TAxis.x), ty: quantize(f.TAxis.y), tz: quantize(f.TAxis.z), to: quantize(f.TOffset),
		flags: f.Flags, value: f.Value, texture: f.Material,
	}
	if idx, ok := ti.index[key]; ok {
		return idx
	}
	idx := int32(len(ti.list))
	ti.list = append(ti.list, &bpb.BSPTexInfo{
		SAxis:       &bpb.Vector3{X: float32(f.SAxis.x), Y: float32(f.SAxis.y), Z: float32(f.SAxis.z)},
		SOffset:     float32(f.SOffset),
		TAxis:       &bpb.Vector3{X: float32(f.TAxis.x), Y: float32(f.TAxis.y), Z: float32(f.TAxis.z)},
		TOffset:     float32(f.TOffset),
		Flags:       f.Flags,
		Value:       f.Value,
		Texture:     f.Material,
		NextTexinfo: -1,
	})
	ti.index[key] = idx
	return idx
}

// builtSide is one contributing (non-degenerate) face of a compiled
// brush: its interned plane/texinfo and the polygon bounding its visible
// extent, derived by clipping a huge quad on its plane against every
// other face of the same brush.
type builtSide struct {
	planeIdx, texinfoIdx int32
	poly                 []vec3
}

// builtBrush is the result of turning one MapBrush into interned
// plane/texinfo references plus per-side bounded polygons, ready to
// become a dbrush_t/dbrushside_t run and to feed the BSP tree builder.
type builtBrush struct {
	contents int32
	sides    []builtSide
	verts    []vec3 // union of every side's polygon vertices, for tree classification
}

// buildBrush computes each face's plane, then its bounded polygon (via
// clipping against the brush's other faces), dropping any face that
// clips away to nothing - a non-contributing "bevel" plane, redundant
// given the brush's other faces. The brush's contents is the union
// (bitwise OR) of every face's declared contents, defaulting to
// CONTENTS_SOLID (1) if the .map didn't specify any.
func buildBrush(mb *MapBrush, planes *planeInterner, texinfos *texinfoInterner) (*builtBrush, error) {
	type facePlane struct {
		n    vec3
		dist float64
	}
	fps := make([]facePlane, len(mb.Faces))
	for i, f := range mb.Faces {
		n, dist := planeFromPoints(f.P0, f.P1, f.P2)
		if n == (vec3{}) {
			return nil, fmt.Errorf("face %d: degenerate plane (its 3 points are collinear)", i)
		}
		fps[i] = facePlane{n, dist}
	}

	built := &builtBrush{}
	haveContents := false
	for i, f := range mb.Faces {
		poly := facePolygonOnPlane(fps[i].n, fps[i].dist)
		for j, other := range fps {
			if j == i {
				continue
			}
			poly = clipPolygon(poly, other.n, other.dist)
			if len(poly) == 0 {
				break
			}
		}
		if len(poly) < 3 {
			continue
		}

		planeIdx := planes.intern(fps[i].n, fps[i].dist)
		texinfoIdx := texinfos.intern(f)
		built.sides = append(built.sides, builtSide{planeIdx, texinfoIdx, poly})
		built.verts = append(built.verts, poly...)

		if f.HasSurfaceAttribs {
			if !haveContents {
				built.contents = f.Contents
				haveContents = true
			} else {
				built.contents |= f.Contents
			}
		}
	}
	if !haveContents {
		built.contents = 1 // CONTENTS_SOLID
	}
	if len(built.sides) < 4 {
		return nil, fmt.Errorf("only %d of %d faces bound a real volume (need at least 4); brush is degenerate", len(built.sides), len(mb.Faces))
	}
	return built, nil
}
