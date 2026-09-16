package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// maxTreeDepth caps BSP tree recursion. Real maps stay well under this;
// it exists so a pathological or numerically unstable input fails with a
// clean error instead of a stack-overflow crash.
const maxTreeDepth = 20000

// maxSplitCandidates bounds how many of the remaining faces are scored as
// candidate splitters at each node, trading optimality for speed on
// large face lists (matching the spirit, if not the exact heuristic, of
// real qbsp's SelectSplit).
const maxSplitCandidates = 32

// buildFace is a face-in-progress during tree construction: a brush
// side's plane/texinfo plus whatever's left of its bounded polygon after
// being clipped by ancestor splitting planes on the way down the tree.
type buildFace struct {
	planeIdx, texinfoIdx int32
	poly                 []vec3
}

// brushCandidate is a brush that might still occupy the region currently
// being partitioned, tracked by its global brush index (for the eventual
// leafbrushes entry) and the vertices of its sides (for classifying it
// against further splitting planes without needing to re-clip its
// geometry).
type brushCandidate struct {
	index int32
	verts []vec3
}

// buildModel turns one entity's brushes into a dmodel_t: it builds every
// brush's dbrush_t/dbrushside_t run (appended to the compiler's global
// tables) and then a BSP tree over their faces.
func (c *compiler) buildModel(mapBrushes []*MapBrush) (*bpb.BSPModel, error) {
	var faces []buildFace
	var candidates []brushCandidate
	var allVerts []vec3

	for _, mb := range mapBrushes {
		built, err := buildBrush(mb, c.planes, c.texinfos)
		if err != nil {
			return nil, err
		}

		brushIdx := int32(len(c.brushes))
		firstSide := int32(len(c.brushSides))
		for _, s := range built.sides {
			c.brushSides = append(c.brushSides, &bpb.BSPBrushSide{PlaneNum: uint32(s.planeIdx), Texinfo: s.texinfoIdx})
			faces = append(faces, buildFace{planeIdx: s.planeIdx, texinfoIdx: s.texinfoIdx, poly: s.poly})
		}
		c.brushes = append(c.brushes, &bpb.BSPBrush{
			FirstSide: firstSide,
			NumSides:  int32(len(built.sides)),
			Contents:  built.contents,
		})

		candidates = append(candidates, brushCandidate{index: brushIdx, verts: built.verts})
		allVerts = append(allVerts, built.verts...)
	}

	var mins, maxs vec3
	if len(allVerts) > 0 {
		mins, maxs = polygonBounds(allVerts)
	}

	tb := newTreeBuilder(c, mins, maxs)
	startFace := int32(len(c.faces))
	headNode, err := tb.build(faces, candidates, 0)
	if err != nil {
		return nil, err
	}
	tb.finish()

	return &bpb.BSPModel{
		Bounds:    &bpb.BoundingBox{Mins: vecToProto(mins), Maxs: vecToProto(maxs)},
		HeadNode:  headNode,
		FirstFace: startFace,
		NumFaces:  int32(len(c.faces)) - startFace,
	}, nil
}

// treeBuilder builds one model's BSP tree. Splitting planes are chosen by
// scoring a bounded sample of the remaining faces (fewest resulting
// spanning faces, tie-broken by front/back balance) - simple compared to
// a tuned compiler's heuristics, but enough to avoid the pathological,
// stack-exhausting recursion depth that always splitting on the first
// available face can produce on real, complex geometry.
//
// Every leaf's Bounds is set to the whole model's bounding box rather
// than the leaf's own tight extent - a safe, conservative
// over-approximation, not a tight fit; nothing here relies on leaf
// bounds being minimal.
type treeBuilder struct {
	c                    *compiler
	worldMins, worldMaxs vec3
	leafRangeStart       int32
	leafFaceAccum        map[int32][]uint32
}

func newTreeBuilder(c *compiler, mins, maxs vec3) *treeBuilder {
	return &treeBuilder{
		c:              c,
		worldMins:      mins,
		worldMaxs:      maxs,
		leafRangeStart: int32(len(c.leaves)),
		leafFaceAccum:  map[int32][]uint32{},
	}
}

// pickSplitter scores up to maxSplitCandidates of the remaining faces as
// potential splitting planes and returns the index of the best one.
func pickSplitter(faces []buildFace, planes *planeInterner) int {
	limit := len(faces)
	if limit > maxSplitCandidates {
		limit = maxSplitCandidates
	}

	best, bestScore := 0, -1
	for i := 0; i < limit; i++ {
		n := planes.normal(faces[i].planeIdx)
		d := planes.dist(faces[i].planeIdx)

		var front, back, spans int
		for _, f := range faces {
			min, max := pointSpread(f.poly, n, d)
			switch {
			case min >= -clipEpsilon && max <= clipEpsilon:
				// coplanar: free, not a split either way
			case min >= -clipEpsilon:
				front++
			case max <= clipEpsilon:
				back++
			default:
				spans++
			}
		}

		balance := front - back
		if balance < 0 {
			balance = -balance
		}
		score := spans*1000 + balance
		if bestScore == -1 || score < bestScore {
			bestScore = score
			best = i
		}
	}
	return best
}

// build partitions faces/brushes into a subtree and returns the value to
// store in the parent's FrontChild/BackChild (a node index, or an encoded
// -(leafIndex+1) if this call produced a leaf directly).
func (tb *treeBuilder) build(faces []buildFace, brushes []brushCandidate, depth int) (int32, error) {
	if len(faces) == 0 {
		return tb.makeLeaf(brushes), nil
	}
	if depth >= maxTreeDepth {
		return 0, fmt.Errorf("BSP tree exceeded maximum depth of %d; input geometry is likely degenerate or numerically unstable", maxTreeDepth)
	}
	splitterIdx := pickSplitter(faces, tb.c.planes)
	splitter := faces[splitterIdx]
	splitN := tb.c.planes.normal(splitter.planeIdx)
	splitD := tb.c.planes.dist(splitter.planeIdx)

	// The splitter is always its own node's onPlane face, identified by
	// index rather than by re-testing its polygon against its own plane:
	// a fragment that's been through many generations of clipping can
	// accumulate enough floating point drift that it fails a fresh
	// coplanarity test against the very plane it was cut from, which
	// would otherwise leave it endlessly bounced between front/back
	// without ever resolving.
	onPlane := []buildFace{splitter}
	var front, back []buildFace
	for i, f := range faces {
		if i == splitterIdx {
			continue
		}
		min, max := pointSpread(f.poly, splitN, splitD)
		switch {
		case min >= -clipEpsilon && max <= clipEpsilon:
			if tb.c.planes.normal(f.planeIdx).dot(splitN) > 0 {
				onPlane = append(onPlane, f)
			} else {
				back = append(back, f)
			}
		case min >= -clipEpsilon:
			front = append(front, f)
		case max <= clipEpsilon:
			back = append(back, f)
		default:
			frontPoly := clipPolygon(f.poly, splitN.scale(-1), -splitD)
			backPoly := clipPolygon(f.poly, splitN, splitD)
			if len(frontPoly) >= 3 {
				front = append(front, buildFace{f.planeIdx, f.texinfoIdx, frontPoly})
			}
			if len(backPoly) >= 3 {
				back = append(back, buildFace{f.planeIdx, f.texinfoIdx, backPoly})
			}
		}
	}
	frontBrushes, backBrushes := splitBrushCandidates(brushes, splitN, splitD)

	frontLeafStart := int32(len(tb.c.leaves))
	frontChild, err := tb.build(front, frontBrushes, depth+1)
	if err != nil {
		return 0, err
	}
	frontLeaves := tb.c.leaves[frontLeafStart:]

	backChild, err := tb.build(back, backBrushes, depth+1)
	if err != nil {
		return 0, err
	}

	firstFace := int32(len(tb.c.faces))
	for _, f := range onPlane {
		tb.c.addFace(f.planeIdx, f.texinfoIdx, f.poly)
	}
	numFaces := uint32(len(onPlane))

	// The onPlane faces are only visible from the front (the side their
	// shared normal points towards, by construction above); register them
	// against every non-solid leaf created in the front subtree so a
	// renderer walking that leaf's leaffaces finds them.
	if numFaces > 0 {
		for i, leaf := range frontLeaves {
			if leaf.GetContents()&1 != 0 {
				continue
			}
			leafIdx := frontLeafStart + int32(i)
			for fi := firstFace; fi < firstFace+int32(numFaces); fi++ {
				tb.leafFaceAccum[leafIdx] = append(tb.leafFaceAccum[leafIdx], uint32(fi))
			}
		}
	}

	nodeIdx := int32(len(tb.c.nodes))
	tb.c.nodes = append(tb.c.nodes, &bpb.BSPNode{
		PlaneNum:   splitter.planeIdx,
		FrontChild: frontChild,
		BackChild:  backChild,
		Bounds:     &bpb.BoundingBox{Mins: vecToProto(tb.worldMins), Maxs: vecToProto(tb.worldMaxs)},
		FirstFace:  uint32(firstFace),
		NumFaces:   numFaces,
	})
	return nodeIdx, nil
}

func (tb *treeBuilder) makeLeaf(brushes []brushCandidate) int32 {
	var contents int32
	brushIdxs := make([]uint32, 0, len(brushes))
	for _, b := range brushes {
		contents |= tb.c.brushes[b.index].GetContents()
		brushIdxs = append(brushIdxs, uint32(b.index))
	}

	cluster := int32(0) // shared "always visible" cluster
	if contents&1 != 0 {
		cluster = -1 // CONTENTS_SOLID: unreachable, needs no visibility info
	}

	firstLeafBrush := uint32(len(tb.c.leafBrushes))
	tb.c.leafBrushes = append(tb.c.leafBrushes, brushIdxs...)

	leafIdx := int32(len(tb.c.leaves))
	tb.c.leaves = append(tb.c.leaves, &bpb.BSPLeaf{
		Contents:       contents,
		Cluster:        cluster,
		Area:           1,
		Bounds:         &bpb.BoundingBox{Mins: vecToProto(tb.worldMins), Maxs: vecToProto(tb.worldMaxs)},
		FirstLeafBrush: firstLeafBrush,
		NumLeafBrushes: uint32(len(brushIdxs)),
	})
	return -(leafIdx + 1)
}

// finish flushes the accumulated per-leaf face lists built up over the
// whole build() recursion into the compiler's global leaffaces table.
func (tb *treeBuilder) finish() {
	for leafIdx := tb.leafRangeStart; int(leafIdx) < len(tb.c.leaves); leafIdx++ {
		faces, ok := tb.leafFaceAccum[leafIdx]
		if !ok {
			continue
		}
		leaf := tb.c.leaves[leafIdx]
		leaf.FirstLeafFace = uint32(len(tb.c.leafFaces))
		leaf.NumLeafFaces = uint32(len(faces))
		tb.c.leafFaces = append(tb.c.leafFaces, faces...)
	}
}

func splitBrushCandidates(brushes []brushCandidate, n vec3, dist float64) (front, back []brushCandidate) {
	for _, b := range brushes {
		min, max := pointSpread(b.verts, n, dist)
		switch {
		case min >= -clipEpsilon:
			front = append(front, b)
		case max <= clipEpsilon:
			back = append(back, b)
		default:
			front = append(front, b)
			back = append(back, b)
		}
	}
	return front, back
}

func vecToProto(v vec3) *bpb.Vector3 {
	return &bpb.Vector3{X: float32(v.x), Y: float32(v.y), Z: float32(v.z)}
}
