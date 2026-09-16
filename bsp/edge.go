package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// edgeSize is the on-disk size of a dedge_t: two unsigned short vertex
// indices.
const edgeSize = 4

// surfEdgeSize is the on-disk size of one surfedges entry: a signed int32
// index into the edges table (negative means traverse it in reverse).
const surfEdgeSize = 4

func parseEdges(data []byte) ([]*bpb.BSPEdge, error) {
	if len(data)%edgeSize != 0 {
		return nil, fmt.Errorf("edges lump size %d is not a multiple of %d", len(data), edgeSize)
	}
	count := len(data) / edgeSize
	edges := make([]*bpb.BSPEdge, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		edges = append(edges, &bpb.BSPEdge{
			Vertex1: r.uint16(),
			Vertex2: r.uint16(),
		})
	}
	return edges, r.err
}

func marshalEdges(edges []*bpb.BSPEdge) []byte {
	w := &writer{}
	for _, e := range edges {
		w.uint16(e.GetVertex1())
		w.uint16(e.GetVertex2())
	}
	return w.buf.Bytes()
}

func parseSurfEdges(data []byte) ([]int32, error) {
	if len(data)%surfEdgeSize != 0 {
		return nil, fmt.Errorf("surfedges lump size %d is not a multiple of %d", len(data), surfEdgeSize)
	}
	count := len(data) / surfEdgeSize
	edges := make([]int32, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		edges = append(edges, r.int32Val())
	}
	return edges, r.err
}

func marshalSurfEdges(edges []int32) []byte {
	w := &writer{}
	for _, e := range edges {
		w.int32Val(e)
	}
	return w.buf.Bytes()
}
