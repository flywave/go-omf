package omf

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestAddTriSurf(t *testing.T) {
	p := &Project{Name: "test"}
	p.AddTriSurf("surface", []Vector3{
		{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0},
	}, []uint32{0, 1, 2, 0, 2, 3})

	if len(p.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(p.Elements))
	}
	e := p.Elements[0]
	if e.Type != ElementTriSurf {
		t.Errorf("expected TriSurf type")
	}
	if e.TriangleCount() != 2 {
		t.Errorf("expected 2 triangles, got %d", e.TriangleCount())
	}
}

func TestBounds(t *testing.T) {
	e := Element{
		Vertices: []Vector3{{0, 0, 0}, {5, 10, 15}, {-2, 3, 8}},
	}
	min, max := e.Bounds()
	if min[0] != -2 || max[0] != 5 {
		t.Errorf("X bounds: [%f, %f]", min[0], max[0])
	}
	if min[1] != 0 || max[1] != 10 {
		t.Errorf("Y bounds: [%f, %f]", min[1], max[1])
	}
	if min[2] != 0 || max[2] != 15 {
		t.Errorf("Z bounds: [%f, %f]", min[2], max[2])
	}
}

func TestAddPointSet(t *testing.T) {
	p := &Project{Name: "points"}
	p.AddPointSet("samples", []Vector3{{100, 200, 300}, {400, 500, 600}})
	if len(p.Elements) != 1 {
		t.Fatalf("expected 1 element")
	}
	if p.Elements[0].VertexCount() != 2 {
		t.Errorf("expected 2 vertices")
	}
}

func TestEmptyBounds(t *testing.T) {
	e := Element{}
	min, max := e.Bounds()
	if min != (Vector3{}) || max != (Vector3{}) {
		t.Errorf("expected zero bounds for empty element")
	}
}

func TestHDF5RoundTrip(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.omf")

	orig := &Project{Name: "roundtrip"}
	orig.AddTriSurf("mesh", []Vector3{
		{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0},
	}, []uint32{0, 1, 2, 0, 2, 3})
	orig.AddPointSet("points", []Vector3{{10, 20, 30}, {40, 50, 60}})

	if err := Save(tmp, orig); err != nil {
		t.Fatal(err)
	}

	loaded, err := Open(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(loaded.Elements))
	}

	e0 := loaded.Elements[0]
	if e0.TriangleCount() != 2 {
		t.Errorf("expected 2 triangles, got %d", e0.TriangleCount())
	}
	if len(e0.Vertices) != 4 {
		t.Errorf("expected 4 vertices in mesh, got %d", len(e0.Vertices))
	}

	e1 := loaded.Elements[1]
	if len(e1.Vertices) != 2 {
		t.Errorf("expected 2 vertices in pointset, got %d", len(e1.Vertices))
	}
}

func TestHDF5FileExists(t *testing.T) {
	if FileExists("/nonexistent/path.omf") {
		t.Error("should return false for nonexistent file")
	}
}

func TestHDF5LargeMesh(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "large.omf")

	n := 1000
	verts := make([]Vector3, n)
	for i := range verts {
		verts[i] = Vector3{float64(i), float64(i) * 2, math.Sqrt(float64(i))}
	}
	tris := make([]uint32, (n-2)*3)
	for i := 0; i < n-2; i++ {
		tris[i*3+0] = uint32(i)
		tris[i*3+1] = uint32(i + 1)
		tris[i*3+2] = uint32(i + 2)
	}

	p := &Project{Name: "large"}
	p.AddTriSurf("bigmesh", verts, tris)

	if err := Save(tmp, p); err != nil {
		t.Fatal(err)
	}

	stat, _ := os.Stat(tmp)
	t.Logf("OMF file size: %d bytes", stat.Size())

	loaded, err := Open(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(loaded.Elements))
	}
	if loaded.Elements[0].TriangleCount() != n-2 {
		t.Errorf("expected %d triangles, got %d", n-2, loaded.Elements[0].TriangleCount())
	}
}
