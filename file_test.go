package omf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenSquareOMF(t *testing.T) {
	if _, err := os.Stat("testdata/square.omf"); os.IsNotExist(err) {
		t.Skip("testdata/square.omf not found")
	}

	p, err := Open("testdata/square.omf")
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Len(t, p.Elements, 1)

	e := p.Elements[0]
	assert.Equal(t, ElementTriSurf, e.Type)
	assert.Equal(t, "square", e.Name)
	assert.Len(t, e.Vertices, 4)
	assert.Equal(t, 2, e.TriangleCount())

	// Verify vertex values
	assert.InDelta(t, 0, e.Vertices[0][0], 0.01)
	assert.InDelta(t, 0, e.Vertices[0][1], 0.01)
	assert.InDelta(t, 1, e.Vertices[3][0], 0.01)
	assert.InDelta(t, 1, e.Vertices[3][1], 0.01)
}

func TestOpenMultiOMF(t *testing.T) {
	if _, err := os.Stat("testdata/multi.omf"); os.IsNotExist(err) {
		t.Skip("testdata/multi.omf not found")
	}

	p, err := Open("testdata/multi.omf")
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Len(t, p.Elements, 3)

	// First element: bigmesh (TriSurf)
	e0 := p.Elements[0]
	assert.Equal(t, ElementTriSurf, e0.Type)
	assert.Equal(t, "bigmesh", e0.Name)
	assert.Equal(t, 28, e0.TriangleCount())

	// Second element: samples (PointSet)
	e1 := p.Elements[1]
	assert.Equal(t, ElementPointSet, e1.Type)
	assert.Equal(t, "samples", e1.Name)
	assert.Len(t, e1.Vertices, 20)

	// Third element: trajectory (PolyLine)
	e2 := p.Elements[2]
	assert.Equal(t, ElementPolyLine, e2.Type)
	assert.Equal(t, "trajectory", e2.Name)
	assert.Len(t, e2.Vertices, 20)
}

func TestRoundTripFile(t *testing.T) {
	if _, err := os.Stat("testdata/square.omf"); os.IsNotExist(err) {
		t.Skip("testdata/square.omf not found")
	}

	orig, err := Open("testdata/square.omf")
	require.NoError(t, err)

	tmp := t.TempDir() + "/roundtrip.omf"
	err = Save(tmp, orig)
	require.NoError(t, err)

	loaded, err := Open(tmp)
	require.NoError(t, err)
	require.Len(t, loaded.Elements, len(orig.Elements))

	for i := range orig.Elements {
		assert.Equal(t, orig.Elements[i].Type, loaded.Elements[i].Type)
		assert.Equal(t, orig.Elements[i].Name, loaded.Elements[i].Name)
		assert.Equal(t, len(orig.Elements[i].Vertices), len(loaded.Elements[i].Vertices))
		assert.Equal(t, len(orig.Elements[i].Indices), len(loaded.Elements[i].Indices))
	}
}

func TestMultiRoundTrip(t *testing.T) {
	if _, err := os.Stat("testdata/multi.omf"); os.IsNotExist(err) {
		t.Skip("testdata/multi.omf not found")
	}

	orig, err := Open("testdata/multi.omf")
	require.NoError(t, err)

	tmp := t.TempDir() + "/multi_rt.omf"
	err = Save(tmp, orig)
	require.NoError(t, err)

	loaded, err := Open(tmp)
	require.NoError(t, err)
	require.Len(t, loaded.Elements, 3)

	for i, e := range loaded.Elements {
		assert.Equal(t, orig.Elements[i].TriangleCount(), e.TriangleCount(), "element %d", i)
	}
}
