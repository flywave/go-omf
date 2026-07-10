package omf

import (
	"bytes"
	"strings"
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

func TestWriteRead(t *testing.T) {
	p := &Project{Name: "test"}
	p.AddTriSurf("mesh", []Vector3{
		{0, 0, 0}, {1, 0, 0}, {1, 1, 0},
	}, []uint32{0, 1, 2})

	var buf bytes.Buffer
	err := Write(&buf, p)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "OMF:TRISURF") {
		t.Errorf("missing TRISURF marker")
	}
	if !strings.Contains(output, "verts=3") {
		t.Errorf("missing vertex count")
	}
}

func TestAddPointSet(t *testing.T) {
	p := &Project{Name: "points"}
	p.AddPointSet("samples", []Vector3{
		{100, 200, 300}, {400, 500, 600},
	})
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
