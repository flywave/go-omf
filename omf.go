package omf

import (
	"fmt"
	"os"

	"github.com/flywave/hdf5"
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
	Name     string
	Type     ElementType
	Vertices []Vector3
	Indices  []uint32
	Data     map[string][]float64
	Color    [3]float32
}

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

func (p *Project) AddPolyLine(name string, points []Vector3) {
	p.Elements = append(p.Elements, Element{
		Name: name, Type: ElementPolyLine,
		Vertices: points, Data: make(map[string][]float64),
	})
}

func (p *Project) AddTetraMesh(name string, verts []Vector3, tets []uint32) {
	p.Elements = append(p.Elements, Element{
		Name: name, Type: ElementTetraMesh,
		Vertices: verts, Indices: tets,
		Data: make(map[string][]float64),
	})
}

func (e *Element) VertexCount() int  { return len(e.Vertices) }
func (e *Element) IndexCount() int   { return len(e.Indices) }
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

func Open(path string) (*Project, error) {
	f, err := hdf5.Open(path)
	if err != nil {
		return nil, fmt.Errorf("omf: open %s: %v", path, err)
	}
	defer f.Close()

	p := &Project{Name: path}
	root := f.Root()
	return p, readElements(root, p)
}

func Save(path string, p *Project) error {
	fw, err := hdf5.CreateForWrite(path, hdf5.CreateTruncate)
	if err != nil {
		return fmt.Errorf("omf: create %s: %v", path, err)
	}
	defer fw.Close()

	return writeElements(fw, p)
}

func readElements(parent interface{ Children() []hdf5.Object }, p *Project) error {
	for _, obj := range parent.Children() {
		if g, ok := obj.(*hdf5.Group); ok {
			elem, err := readElement(g)
			if err == nil {
				p.Elements = append(p.Elements, elem)
			}
		}
	}
	return nil
}

func readElement(g *hdf5.Group) (Element, error) {
	e := Element{Data: make(map[string][]float64)}
	e.Name = g.Name()

	if attrs, err := g.Attributes(); err == nil {
		for _, a := range attrs {
			val := hdf5.AttributeStringValue(a)
			switch a.Name {
			case "element_type":
				e.Type = parseElementType(val)
			case "name":
				e.Name = val
			}
		}
	}

	for _, child := range g.Children() {
		if ds, ok := child.(*hdf5.Dataset); ok {
			switch ds.Name() {
			case "vertices":
				if data, err := ds.Read(); err == nil {
					for i := 0; i+2 < len(data); i += 3 {
						e.Vertices = append(e.Vertices, Vector3{
							data[i], data[i+1], data[i+2],
						})
					}
				}
			case "indices":
				if data, err := ds.Read(); err == nil {
					for _, v := range data {
						e.Indices = append(e.Indices, uint32(v))
					}
				}
			}
		}
	}

	return e, nil
}

func writeElements(fw *hdf5.FileWriter, p *Project) error {
	for idx, elem := range p.Elements {
		g, err := fw.CreateGroup(fmt.Sprintf("/element_%d", idx))
		if err != nil {
			return fmt.Errorf("omf: create group: %v", err)
		}
		g.WriteAttribute("element_type", elementTypeName(elem.Type))
		g.WriteAttribute("name", elem.Name)

		if len(elem.Vertices) > 0 {
			flat := make([]float64, len(elem.Vertices)*3)
			for i, v := range elem.Vertices {
				flat[i*3+0] = v[0]
				flat[i*3+1] = v[1]
				flat[i*3+2] = v[2]
			}
			ds, err := fw.CreateDataset(
				fmt.Sprintf("/element_%d/vertices", idx),
				hdf5.Float64, []uint64{uint64(len(flat))})
			if err != nil {
				return fmt.Errorf("omf: create vertices: %v", err)
			}
			if err := ds.Write(flat); err != nil {
				return fmt.Errorf("omf: write vertices: %v", err)
			}
			ds.Close()
		}

		if len(elem.Indices) > 0 {
			ints := make([]float64, len(elem.Indices))
			for i, v := range elem.Indices {
				ints[i] = float64(v)
			}
			ds, err := fw.CreateDataset(
				fmt.Sprintf("/element_%d/indices", idx),
				hdf5.Float64, []uint64{uint64(len(ints))})
			if err != nil {
				return fmt.Errorf("omf: create indices: %v", err)
			}
			if err := ds.Write(ints); err != nil {
				return fmt.Errorf("omf: write indices: %v", err)
			}
			ds.Close()
		}
	}
	return nil
}

func parseElementType(s string) ElementType {
	switch s {
	case "PointSet":
		return ElementPointSet
	case "PolyLine":
		return ElementPolyLine
	case "TriSurf":
		return ElementTriSurf
	case "TetraMesh":
		return ElementTetraMesh
	case "Volume":
		return ElementVolume
	}
	return ElementPointSet
}

func elementTypeName(t ElementType) string {
	switch t {
	case ElementPointSet:
		return "PointSet"
	case ElementPolyLine:
		return "PolyLine"
	case ElementTriSurf:
		return "TriSurf"
	case ElementTetraMesh:
		return "TetraMesh"
	case ElementVolume:
		return "Volume"
	}
	return "Unknown"
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
