package bsp

import (
	"fmt"
	"math"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// Compile builds a .bsp file from parsed .map source, playing the role of
// id Software's qbsp (turning brushes into a BSP tree of nodes/leaves)
// without the vis or light stages: every leaf shares a single visibility
// cluster (so the compiled map has no PVS culling - correct, just
// unoptimized, exactly like an engine's normal fallback for an unvised
// map) and every face is fullbright (LightOffset -1, no lighting lump).
// Area portals aren't computed either; every leaf is placed in a single
// area.
//
// The BSP tree itself is built with a simple, unoptimized splitting
// heuristic (always split on the first remaining face), which is
// correctness-preserving but can be slow and produce a larger tree than a
// tuned compiler would for big, complex maps.
func Compile(mf *MapFile) (*bpb.BSPFile, error) {
	c := newCompiler()
	bsp := &bpb.BSPFile{
		Header: &bpb.BSPHeader{Magic: Magic, Version: Version},
		Pop:    make([]byte, 256),
		Visibility: &bpb.BSPVisibility{
			ClusterCount: 1,
			Offsets:      []*bpb.BSPVisOffset{{Pvs: 0, Phs: 0}},
		},
		Areas: []*bpb.BSPArea{{}},
	}

	for i, ent := range mf.Entities {
		props := make(map[string]string, len(ent.Properties)+1)
		for k, v := range ent.Properties {
			props[k] = v
		}

		hasModel := i == 0 || len(ent.Brushes) > 0
		if hasModel {
			modelIdx := len(bsp.Models)
			model, err := c.buildModel(ent.Brushes)
			if err != nil {
				return nil, fmt.Errorf("entity %d (%s): %w", i, props["classname"], err)
			}
			if i != 0 {
				props["model"] = fmt.Sprintf("*%d", modelIdx)
			}
			bsp.Models = append(bsp.Models, model)
		}

		bsp.Entities = append(bsp.Entities, &bpb.BSPEntity{
			ClassName:  props["classname"],
			Properties: props,
		})
	}

	bsp.Planes = c.planes.list
	bsp.TextureInfo = c.texinfos.list
	bsp.Vertices = c.vertices
	bsp.Edges = c.edges
	bsp.SurfEdges = c.surfEdges
	bsp.Faces = c.faces
	bsp.Nodes = c.nodes
	bsp.Leaves = c.leaves
	bsp.LeafBrushes = c.leafBrushes
	bsp.LeafFaces = c.leafFaces
	bsp.Brushes = c.brushes
	bsp.BrushSides = c.brushSides

	return bsp, nil
}

// compiler accumulates the global, shared-across-all-models tables that
// make up most of a .bsp file (planes, texinfo, vertices, edges, faces,
// nodes, leaves, and the brush/leaf index tables), plus the interners
// used to deduplicate planes and texinfo as brushes are built.
type compiler struct {
	planes   *planeInterner
	texinfos *texinfoInterner

	vertices  []*bpb.BSPVertex
	vertexIdx map[[3]int32]uint32

	edges   []*bpb.BSPEdge
	edgeIdx map[[2]uint32]int32

	surfEdges []int32
	faces     []*bpb.BSPFace

	nodes  []*bpb.BSPNode
	leaves []*bpb.BSPLeaf

	leafBrushes []uint32
	leafFaces   []uint32

	brushes    []*bpb.BSPBrush
	brushSides []*bpb.BSPBrushSide
}

func newCompiler() *compiler {
	return &compiler{
		planes:    newPlaneInterner(),
		texinfos:  newTexinfoInterner(),
		vertexIdx: map[[3]int32]uint32{},
		edges:     []*bpb.BSPEdge{{}}, // index 0 is reserved/unused
		edgeIdx:   map[[2]uint32]int32{},
	}
}

// quantize snaps a coordinate to a fine grid so that near-identical
// floating point values (from independently-computed brush faces meeting
// at a shared vertex, for instance) intern to the same table entry.
func quantize(v float64) int32 {
	return int32(math.Round(v * 8))
}

func (c *compiler) internVertex(p vec3) uint32 {
	key := [3]int32{quantize(p.x), quantize(p.y), quantize(p.z)}
	if idx, ok := c.vertexIdx[key]; ok {
		return idx
	}
	idx := uint32(len(c.vertices))
	c.vertices = append(c.vertices, &bpb.BSPVertex{Point: &bpb.Vector3{X: float32(p.x), Y: float32(p.y), Z: float32(p.z)}})
	c.vertexIdx[key] = idx
	return idx
}

// internEdge returns a surfedge value for the directed edge v1->v2:
// positive if a new or matching edge stores that exact direction,
// negative if the shared edge entry stores the opposite direction.
func (c *compiler) internEdge(v1, v2 uint32) int32 {
	key := [2]uint32{v1, v2}
	reversed := v1 > v2
	if reversed {
		key = [2]uint32{v2, v1}
	}
	if idx, ok := c.edgeIdx[key]; ok {
		if reversed {
			return -idx
		}
		return idx
	}
	idx := int32(len(c.edges))
	c.edges = append(c.edges, &bpb.BSPEdge{Vertex1: v1, Vertex2: v2})
	c.edgeIdx[key] = idx
	if reversed {
		return -idx
	}
	return idx
}

// addFace appends a face built from a bounded polygon to the global
// faces/surfedges tables and returns its index.
func (c *compiler) addFace(planeIdx, texinfoIdx int32, poly []vec3) int32 {
	firstEdge := int32(len(c.surfEdges))
	n := len(poly)
	for i := 0; i < n; i++ {
		v1 := c.internVertex(poly[i])
		v2 := c.internVertex(poly[(i+1)%n])
		c.surfEdges = append(c.surfEdges, c.internEdge(v1, v2))
	}
	faceIdx := int32(len(c.faces))
	c.faces = append(c.faces, &bpb.BSPFace{
		PlaneNum:    uint32(planeIdx),
		Side:        0,
		FirstEdge:   firstEdge,
		NumEdges:    int32(n),
		Texinfo:     texinfoIdx,
		LightStyles: []uint32{0, 255, 255, 255},
		LightOffset: -1,
	})
	return faceIdx
}
