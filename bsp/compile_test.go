package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// axisBoxBrush builds a MapBrush for an axis-aligned box directly (rather
// than through .map text), using facePolygonOnPlane to get correctly
// wound 3-point face definitions for each of the 6 sides.
func axisBoxBrush(mins, maxs vec3, texture string) *MapBrush {
	defs := []struct {
		n vec3
		d float64
	}{
		{vec3{x: -1}, -mins.x},
		{vec3{x: 1}, maxs.x},
		{vec3{y: -1}, -mins.y},
		{vec3{y: 1}, maxs.y},
		{vec3{z: -1}, -mins.z},
		{vec3{z: 1}, maxs.z},
	}
	brush := &MapBrush{}
	for _, d := range defs {
		poly := facePolygonOnPlane(d.n, d.d)
		brush.Faces = append(brush.Faces, &MapBrushFace{
			P0: poly[0], P1: poly[1], P2: poly[2],
			Material:          texture,
			SAxis:             vec3{x: 1},
			TAxis:             vec3{y: -1},
			ScaleX:            1,
			ScaleY:            1,
			HasSurfaceAttribs: true,
			Contents:          1, // CONTENTS_SOLID
		})
	}
	return brush
}

// classifyPoint walks a model's BSP tree and returns the contents of the
// leaf containing p.
func classifyPoint(t *testing.T, bsp *bpb.BSPFile, headNode int32, p vec3) int32 {
	t.Helper()
	node := headNode
	for node >= 0 {
		if int(node) >= len(bsp.GetNodes()) {
			t.Fatalf("node index %d out of range (have %d)", node, len(bsp.GetNodes()))
		}
		n := bsp.GetNodes()[node]
		plane := bsp.GetPlanes()[n.GetPlaneNum()]
		normal := vec3{float64(plane.GetNormal().GetX()), float64(plane.GetNormal().GetY()), float64(plane.GetNormal().GetZ())}
		if p.dot(normal)-float64(plane.GetDistance()) >= 0 {
			node = n.GetFrontChild()
		} else {
			node = n.GetBackChild()
		}
	}
	leafIdx := -(node + 1)
	if leafIdx < 0 || int(leafIdx) >= len(bsp.GetLeaves()) {
		t.Fatalf("leaf index %d out of range (have %d)", leafIdx, len(bsp.GetLeaves()))
	}
	return bsp.GetLeaves()[leafIdx].GetContents()
}

func TestCompileSingleCube(t *testing.T) {
	mf := &MapFile{
		Entities: []*MapEntity{
			{
				Properties: map[string]string{"classname": "worldspawn"},
				Brushes:    []*MapBrush{axisBoxBrush(vec3{}, vec3{64, 64, 64}, "walls/brick1")},
			},
		},
	}

	bsp, err := Compile(mf)
	if err != nil {
		t.Fatal(err)
	}

	if len(bsp.GetModels()) != 1 {
		t.Fatalf("model count - have: %d, want: 1", len(bsp.GetModels()))
	}
	if len(bsp.GetBrushes()) != 1 {
		t.Fatalf("brush count - have: %d, want: 1", len(bsp.GetBrushes()))
	}
	if got := bsp.GetBrushes()[0].GetContents(); got != 1 {
		t.Errorf("brush contents - have: %d, want: 1", got)
	}
	if got, want := int(bsp.GetModels()[0].GetNumFaces()), 6; got != want {
		t.Errorf("face count - have: %d, want: %d", got, want)
	}

	data, err := Marshal(bsp)
	if err != nil {
		t.Fatalf("marshaling compiled bsp: %v", err)
	}
	if _, err := Unmarshal(data); err != nil {
		t.Fatalf("re-parsing compiled bsp: %v", err)
	}

	headNode := bsp.GetModels()[0].GetHeadNode()
	if got := classifyPoint(t, bsp, headNode, vec3{32, 32, 32}); got&1 == 0 {
		t.Errorf("center of cube - contents %d, want CONTENTS_SOLID bit set", got)
	}
	if got := classifyPoint(t, bsp, headNode, vec3{1000, 1000, 1000}); got&1 != 0 {
		t.Errorf("far outside cube - contents %d, want CONTENTS_SOLID bit clear", got)
	}
}

func TestCompileDegenerateBrush(t *testing.T) {
	// A brush with only 3 faces can't bound a finite solid.
	tooFewFaces := &MapBrush{
		Faces: axisBoxBrush(vec3{}, vec3{64, 64, 64}, "x").Faces[:3],
	}
	mf := &MapFile{
		Entities: []*MapEntity{
			{Properties: map[string]string{"classname": "worldspawn"}, Brushes: []*MapBrush{tooFewFaces}},
		},
	}
	if _, err := Compile(mf); err == nil {
		t.Error("expected an error compiling a brush with too few faces")
	}
}

// TestCompileHollowRoom builds a room out of 6 separate wall/floor/ceiling
// brushes (not one solid box) and checks that the interior classifies as
// empty, a wall as solid, and the whole thing survives a marshal/unmarshal
// round trip.
func TestCompileHollowRoom(t *testing.T) {
	const (
		inner = 128.0
		wall  = 16.0
	)
	outer := inner + 2*wall

	brush := func(mins, maxs vec3) *MapBrush {
		return axisBoxBrush(mins, maxs, "walls/brick1")
	}

	brushes := []*MapBrush{
		brush(vec3{0, 0, 0}, vec3{outer, outer, wall}),                                 // floor
		brush(vec3{0, 0, outer - wall}, vec3{outer, outer, outer}),                     // ceiling
		brush(vec3{0, 0, wall}, vec3{wall, outer, outer - wall}),                       // wall -X
		brush(vec3{outer - wall, 0, wall}, vec3{outer, outer, outer - wall}),           // wall +X
		brush(vec3{wall, 0, wall}, vec3{outer - wall, wall, outer - wall}),             // wall -Y
		brush(vec3{wall, outer - wall, wall}, vec3{outer - wall, outer, outer - wall}), // wall +Y
	}

	mf := &MapFile{
		Entities: []*MapEntity{
			{
				Properties: map[string]string{"classname": "worldspawn"},
				Brushes:    brushes,
			},
			{
				Properties: map[string]string{"classname": "info_player_start", "origin": "64 64 64"},
			},
		},
	}

	bsp, err := Compile(mf)
	if err != nil {
		t.Fatal(err)
	}
	if len(bsp.GetEntities()) != 2 {
		t.Fatalf("entity count - have: %d, want: 2", len(bsp.GetEntities()))
	}
	if _, ok := bsp.GetEntities()[1].GetProperties()["model"]; ok {
		t.Error("a point entity shouldn't get a \"model\" property")
	}
	if len(bsp.GetBrushes()) != 6 {
		t.Fatalf("brush count - have: %d, want: 6", len(bsp.GetBrushes()))
	}

	data, err := Marshal(bsp)
	if err != nil {
		t.Fatalf("marshaling compiled bsp: %v", err)
	}
	if _, err := Unmarshal(data); err != nil {
		t.Fatalf("re-parsing compiled bsp: %v", err)
	}

	headNode := bsp.GetModels()[0].GetHeadNode()
	center := vec3{outer / 2, outer / 2, outer / 2}
	if got := classifyPoint(t, bsp, headNode, center); got&1 != 0 {
		t.Errorf("room center - contents %d, want CONTENTS_SOLID bit clear (empty/playable space)", got)
	}
	insideFloor := vec3{outer / 2, outer / 2, wall / 2}
	if got := classifyPoint(t, bsp, headNode, insideFloor); got&1 == 0 {
		t.Errorf("inside floor brush - contents %d, want CONTENTS_SOLID bit set", got)
	}

	// The center leaf should list at least one face (it should be able to
	// see the surrounding walls/floor/ceiling).
	node := headNode
	for node >= 0 {
		n := bsp.GetNodes()[node]
		plane := bsp.GetPlanes()[n.GetPlaneNum()]
		normal := vec3{float64(plane.GetNormal().GetX()), float64(plane.GetNormal().GetY()), float64(plane.GetNormal().GetZ())}
		if center.dot(normal)-float64(plane.GetDistance()) >= 0 {
			node = n.GetFrontChild()
		} else {
			node = n.GetBackChild()
		}
	}
	leaf := bsp.GetLeaves()[-(node + 1)]
	if leaf.GetNumLeafFaces() == 0 {
		t.Error("expected the room's central leaf to reference at least one face")
	}
}

func TestCompileDecompileRoundTrip(t *testing.T) {
	mf := &MapFile{
		Entities: []*MapEntity{
			{
				Properties: map[string]string{"classname": "worldspawn"},
				Brushes:    []*MapBrush{axisBoxBrush(vec3{}, vec3{64, 64, 64}, "walls/brick1")},
			},
			{
				Properties: map[string]string{"classname": "func_wall"},
				Brushes:    []*MapBrush{axisBoxBrush(vec3{128, 0, 0}, vec3{192, 64, 64}, "walls/brick1")},
			},
		},
	}
	bsp, err := Compile(mf)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := bsp.GetEntities()[1].GetProperties()["model"], "*1"; got != want {
		t.Errorf("func_wall model property - have: %q, want: %q", got, want)
	}

	out, err := Decompile(bsp)
	if err != nil {
		t.Fatal(err)
	}
	mf2, err := ParseMap([]byte(out))
	if err != nil {
		t.Fatalf("re-parsing decompiled output: %v", err)
	}
	if len(mf2.Entities) != 2 {
		t.Fatalf("entity count - have: %d, want: 2", len(mf2.Entities))
	}
	if len(mf2.Entities[0].Brushes) != 1 || len(mf2.Entities[1].Brushes) != 1 {
		t.Errorf("brush distribution - have: %d/%d, want: 1/1", len(mf2.Entities[0].Brushes), len(mf2.Entities[1].Brushes))
	}
}

// TestCompileRealMap decompiles the real 248-brush/65-entity test map and
// recompiles the result, exercising the full pipeline against actual
// non-trivial mapper geometry rather than small synthetic shapes. This is
// the test that caught a real bug: an earlier version of the splitting
// heuristic reclassified the chosen splitter face against its own plane
// by re-testing its (possibly drifted, after many clipping generations)
// polygon rather than identifying it by index, which on this map's real
// complexity caused the tree builder to never terminate.
func TestCompileRealMap(t *testing.T) {
	original, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	src, err := Decompile(original)
	if err != nil {
		t.Fatal(err)
	}
	mf, err := ParseMap([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	compiled, err := Compile(mf)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(compiled.GetBrushes()), len(original.GetBrushes()); got != want {
		t.Errorf("brush count - have: %d, want: %d", got, want)
	}
	if got, want := len(compiled.GetModels()), len(original.GetModels()); got != want {
		t.Errorf("model count - have: %d, want: %d", got, want)
	}
	if got, want := len(compiled.GetEntities()), len(original.GetEntities()); got != want {
		t.Errorf("entity count - have: %d, want: %d", got, want)
	}
	if len(compiled.GetNodes()) == 0 || len(compiled.GetLeaves()) == 0 {
		t.Error("expected a non-trivial BSP tree")
	}

	data, err := Marshal(compiled)
	if err != nil {
		t.Fatalf("marshaling recompiled bsp: %v", err)
	}
	if _, err := Unmarshal(data); err != nil {
		t.Fatalf("re-parsing recompiled bsp: %v", err)
	}
}
