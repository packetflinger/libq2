package bsp

import (
	"math"
	"strings"
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestDecompileHeader(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Decompile(bsp)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"// Game: Quake 2\n", "// Format: Quake2 (Valve)\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing header line %q", want)
		}
	}
}

// TestDecompileRoundTrip decompiles the real test map, parses the result
// back with ParseMap, and checks the parsed data against the original
// BSPFile: entity count/properties, every brush accounted for exactly
// once, and - critically - that reconstructing each face's plane from its
// 3 emitted points reproduces the original plane's normal and distance.
// That last check is the one that would catch a winding-order mistake in
// facePlanePoints.
func TestDecompileRoundTrip(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Decompile(bsp)
	if err != nil {
		t.Fatal(err)
	}

	mf, err := ParseMap([]byte(out))
	if err != nil {
		t.Fatalf("re-parsing decompiled map: %v", err)
	}

	if len(mf.Entities) != len(bsp.GetEntities()) {
		t.Fatalf("entity count - have: %d, want: %d", len(mf.Entities), len(bsp.GetEntities()))
	}

	// worldspawn properties (minus "model", which we intentionally drop)
	// round trip exactly.
	wantProps := bsp.GetEntities()[0].GetProperties()
	gotProps := mf.Entities[0].Properties
	for k, v := range wantProps {
		if k == "model" {
			continue
		}
		if gotProps[k] != v {
			t.Errorf("worldspawn property %q - have: %q, want: %q", k, gotProps[k], v)
		}
	}

	// Every brush in the .bsp should appear in exactly one entity's brush
	// list, since every real brush belongs to exactly one model.
	totalBrushes := 0
	for _, ent := range mf.Entities {
		totalBrushes += len(ent.Brushes)
	}
	if totalBrushes != len(bsp.GetBrushes()) {
		t.Errorf("total brush count across all entities - have: %d, want: %d", totalBrushes, len(bsp.GetBrushes()))
	}

	// Spot-check every brush side's plane and texinfo actually round trip:
	// walk model 0 (worldspawn) the same way Decompile does, and compare
	// each face's re-derived plane and texture info against the source.
	modelBrushes, err := modelBrushIndices(bsp)
	if err != nil {
		t.Fatal(err)
	}
	worldBrushIndices := modelBrushes[0]
	worldBrushes := mf.Entities[0].Brushes
	if len(worldBrushes) != len(worldBrushIndices) {
		t.Fatalf("world brush count - have: %d, want: %d", len(worldBrushes), len(worldBrushIndices))
	}

	const eps = 1e-3
	checked := 0
	for i, brushIdx := range worldBrushIndices {
		brush := bsp.GetBrushes()[brushIdx]
		parsedBrush := worldBrushes[i]
		if len(parsedBrush.Faces) != int(brush.GetNumSides()) {
			t.Fatalf("brush %d face count - have: %d, want: %d", brushIdx, len(parsedBrush.Faces), brush.GetNumSides())
		}
		for j, face := range parsedBrush.Faces {
			side := bsp.GetBrushSides()[brush.GetFirstSide()+int32(j)]
			wantPlane := bsp.GetPlanes()[side.GetPlaneNum()]
			wantNormal := vec3{
				float64(wantPlane.GetNormal().GetX()),
				float64(wantPlane.GetNormal().GetY()),
				float64(wantPlane.GetNormal().GetZ()),
			}

			gotNormal, gotDist := planeFromPoints(face.P0, face.P1, face.P2)
			if math.Abs(gotNormal.dot(wantNormal)-1) > eps {
				t.Errorf("brush %d face %d: normal mismatch - have: %+v, want: %+v", brushIdx, j, gotNormal, wantNormal)
			}
			if math.Abs(gotDist-float64(wantPlane.GetDistance())) > eps {
				t.Errorf("brush %d face %d: distance - have: %v, want: %v", brushIdx, j, gotDist, wantPlane.GetDistance())
			}

			if side.GetTexinfo() >= 0 {
				ti := bsp.GetTextureInfo()[side.GetTexinfo()]
				if face.Material != ti.GetTexture() {
					t.Errorf("brush %d face %d: material - have: %q, want: %q", brushIdx, j, face.Material, ti.GetTexture())
				}
				if float32(face.SOffset) != ti.GetSOffset() || float32(face.TOffset) != ti.GetTOffset() {
					t.Errorf("brush %d face %d: offsets - have: (%v,%v), want: (%v,%v)", brushIdx, j, face.SOffset, face.TOffset, ti.GetSOffset(), ti.GetTOffset())
				}
				if !face.HasSurfaceAttribs {
					t.Errorf("brush %d face %d: expected surface attributes to be present", brushIdx, j)
				} else if face.Flags != ti.GetFlags() || face.Value != ti.GetValue() {
					t.Errorf("brush %d face %d: flags/value - have: (%d,%d), want: (%d,%d)", brushIdx, j, face.Flags, face.Value, ti.GetFlags(), ti.GetValue())
				}
				if face.Contents != brush.GetContents() {
					t.Errorf("brush %d face %d: contents - have: %d, want: %d", brushIdx, j, face.Contents, brush.GetContents())
				}
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("no faces were checked")
	}
}

func TestDecompileBrushEntity(t *testing.T) {
	// Build a tiny synthetic BSP: worldspawn (model 0, empty) plus one
	// func_door brush entity (model 1, a single unit cube), to check the
	// "model" "*N" -> entity brush association without depending on the
	// specifics of the real test map.
	cube := unitCubeBrush(t, 0, 0)

	bsp := &bpb.BSPFile{
		Entities: []*bpb.BSPEntity{
			{ClassName: "worldspawn", Properties: map[string]string{"classname": "worldspawn"}},
			{ClassName: "func_door", Properties: map[string]string{"classname": "func_door", "model": "*1"}},
		},
		Planes:      cube.planes,
		TextureInfo: cube.texinfo,
		Brushes:     []*bpb.BSPBrush{cube.brush},
		BrushSides:  cube.sides,
		Nodes:       []*bpb.BSPNode{},
		Leaves: []*bpb.BSPLeaf{
			{FirstLeafBrush: 0, NumLeafBrushes: 1},
		},
		LeafBrushes: []uint32{0},
		Models: []*bpb.BSPModel{
			{HeadNode: -1}, // model 0 (world): leaf -(-1+1)=0, which has no brushes here
			{HeadNode: -1}, // model 1 (func_door): same single leaf, which owns the cube brush
		},
	}
	// Give model 0 an empty leaf and model 1 the leaf with the brush by
	// using distinct leaves.
	bsp.Leaves = []*bpb.BSPLeaf{
		{FirstLeafBrush: 0, NumLeafBrushes: 0}, // leaf 0: empty, for model 0
		{FirstLeafBrush: 0, NumLeafBrushes: 1}, // leaf 1: has the cube brush, for model 1
	}
	bsp.Models[0].HeadNode = -1 // -(-1+1) = leaf 0
	bsp.Models[1].HeadNode = -2 // -(-2+1) = leaf 1

	out, err := Decompile(bsp)
	if err != nil {
		t.Fatal(err)
	}
	mf, err := ParseMap([]byte(out))
	if err != nil {
		t.Fatalf("re-parsing decompiled map: %v", err)
	}
	if len(mf.Entities) != 2 {
		t.Fatalf("entity count - have: %d, want: 2", len(mf.Entities))
	}
	if len(mf.Entities[0].Brushes) != 0 {
		t.Errorf("worldspawn brush count - have: %d, want: 0", len(mf.Entities[0].Brushes))
	}
	if len(mf.Entities[1].Brushes) != 1 {
		t.Fatalf("func_door brush count - have: %d, want: 1", len(mf.Entities[1].Brushes))
	}
	if _, ok := mf.Entities[1].Properties["model"]; ok {
		t.Error("expected the stale \"model\" property to be dropped from decompiled output")
	}
	if len(mf.Entities[1].Brushes[0].Faces) != 6 {
		t.Errorf("cube face count - have: %d, want: 6", len(mf.Entities[1].Brushes[0].Faces))
	}
}

// cubeData holds the plane/texinfo/brush/brushside data for a synthetic
// axis-aligned unit brush, for use in tests that don't want to depend on
// the specifics of testdata/backup.bsp.
type cubeData struct {
	planes  []*bpb.BSPPlane
	texinfo []*bpb.BSPTexInfo
	brush   *bpb.BSPBrush
	sides   []*bpb.BSPBrushSide
}

// unitCubeBrush builds the 6 axial planes of a brush spanning
// [mins,mins+64] on each axis, all using a single shared texinfo.
func unitCubeBrush(t *testing.T, minX, minY float32) cubeData {
	t.Helper()
	const size = 64
	minZ := float32(0)

	type pd struct {
		normal   *bpb.Vector3
		distance float32
	}
	defs := []pd{
		{&bpb.Vector3{X: -1}, -minX},
		{&bpb.Vector3{X: 1}, minX + size},
		{&bpb.Vector3{Y: -1}, -minY},
		{&bpb.Vector3{Y: 1}, minY + size},
		{&bpb.Vector3{Z: -1}, -minZ},
		{&bpb.Vector3{Z: 1}, minZ + size},
	}

	var planes []*bpb.BSPPlane
	var sides []*bpb.BSPBrushSide
	for i, d := range defs {
		planes = append(planes, &bpb.BSPPlane{Normal: d.normal, Distance: d.distance, Type: int32(i % 3)})
		sides = append(sides, &bpb.BSPBrushSide{PlaneNum: uint32(i), Texinfo: 0})
	}

	texinfo := []*bpb.BSPTexInfo{
		{
			SAxis:   &bpb.Vector3{X: 1},
			TAxis:   &bpb.Vector3{Y: -1},
			Texture: "walls/brick1",
		},
	}

	brush := &bpb.BSPBrush{FirstSide: 0, NumSides: int32(len(sides)), Contents: 1}

	return cubeData{planes: planes, texinfo: texinfo, brush: brush, sides: sides}
}
