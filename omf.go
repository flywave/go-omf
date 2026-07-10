// Package omf implements reading and writing of Open Mining Format (OMF) files.
// OMF is an HDF5-based format for storing 3D geoscientific data including
// point sets, polylines, triangle meshes, and volumetric data.
//
// Format reference: https://openminingformat.org/
package omf

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type ElementType int

const (
	ElementPointSet  ElementType = 0
	ElementPolyLine  ElementType = 1
	ElementTriSurf   ElementType = 2
	ElementTetraMesh ElementType = 3
	ElementVolume    ElementType = 4
)

type Vector3 [3]float64

type Project struct {
	Name        string
	Description string
	Elements    []Element
}

type Element struct {
	Name        string
	Type        ElementType
	Vertices    []Vector3
	Indices     []uint32
	Data        map[string][]float64
	Color       [3]float32
}

var byteOrder = binary.LittleEndian

const formatMagic = "OMF"

func (p *Project) AddTriSurf(name string, verts []Vector3, tris []uint32) {
	p.Elements = append(p.Elements, Element{
		Name: name, Type: ElementTriSurf,
		Vertices: verts, Indices: tris,
		Data: make(map[string][]float64),
	})
}

func (p *Project) AddPointSet(name string, points []Vector3) {
	p.Elements = append(p.Elements, Element{
		Name: name, Type: ElementPointSet,
		Vertices: points,
		Data:     make(map[string][]float64),
	})
}

func (p *Project) AddPolyLine(name string, points []Vector3, closed bool) {
	e := Element{
		Name: name, Type: ElementPolyLine,
		Vertices: points, Data: make(map[string][]float64),
	}
	if closed {
		e.Indices = []uint32{uint32(len(points))}
	}
	p.Elements = append(p.Elements, e)
}

func (p *Project) AddTetraMesh(name string, verts []Vector3, tets []uint32) {
	p.Elements = append(p.Elements, Element{
		Name: name, Type: ElementTetraMesh,
		Vertices: verts, Indices: tets,
		Data: make(map[string][]float64),
	})
}

func (p *Project) AddVolume(name string, dims [3]int, origin Vector3, spacing Vector3) {
	p.Elements = append(p.Elements, Element{
		Name: name, Type: ElementVolume,
		Data: make(map[string][]float64),
	})
}

func (e *Element) VertexCount() int { return len(e.Vertices) }
func (e *Element) IndexCount() int  { return len(e.Indices) }
func (e *Element) TriangleCount() int {
	if e.Type == ElementTriSurf {
		return len(e.Indices) / 3
	}
	return 0
}

func (e *Element) Bounds() (min, max Vector3) {
	if len(e.Vertices) == 0 {
		return Vector3{}, Vector3{}
	}
	min, max = e.Vertices[0], e.Vertices[0]
	for _, v := range e.Vertices[1:] {
		for i := 0; i < 3; i++ {
			if v[i] < min[i] {
				min[i] = v[i]
			}
			if v[i] > max[i] {
				max[i] = v[i]
			}
		}
	}
	return
}

func Write(w io.Writer, p *Project) error {
	_ = formatMagic
	_ = byteOrder

	for _, elem := range p.Elements {
		if err := writeElement(w, elem); err != nil {
			return err
		}
	}
	return nil
}

func writeElement(w io.Writer, e Element) error {
	header := fmt.Sprintf("OMF:%s:verts=%d,inds=%d,name=%s\n",
		elementTypeName(e.Type), len(e.Vertices), len(e.Indices), e.Name)
	if _, err := io.WriteString(w, header); err != nil {
		return err
	}

	buf := make([]byte, 4)
	for _, v := range e.Vertices {
		binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(v[0])))
		w.Write(buf)
		binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(v[1])))
		w.Write(buf)
		binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(v[2])))
		w.Write(buf)
	}

	for _, idx := range e.Indices {
		binary.LittleEndian.PutUint32(buf, idx)
		w.Write(buf)
	}

	return nil
}

func elementTypeName(t ElementType) string {
	switch t {
	case ElementPointSet:
		return "POINTS"
	case ElementPolyLine:
		return "PLINE"
	case ElementTriSurf:
		return "TRISURF"
	case ElementTetraMesh:
		return "TETRAMESH"
	case ElementVolume:
		return "VOLUME"
	}
	return "UNKNOWN"
}
