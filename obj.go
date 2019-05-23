/*
Package gwob is a pure Go parser for Wavefront .OBJ 3D geometry file format.

Example:

    // Error handling omitted for simplicity.

    import "github.com/udhos/gwob"

    options := &gwob.ObjParserOptions{} // parser options

    o, errObj := gwob.NewObjFromFile("gopher.obj", options) // parse

    // Scan OBJ groups
    for _, g := range o.Groups {
        // snip
    }

See also: https://github.com/udhos/gwob
*/
package gwob

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// Internal parsing error
const (
	ErrFatal    = true  // ErrFatal means fatal stream error
	ErrNonFatal = false // ErrNonFatal means non-fatal parsing error
)

// Material holds information for a material.
type Material struct {
	Name  string
	MapKd string
	Kd    [3]float32
}

// MaterialLib stores materials.
type MaterialLib struct {
	Lib map[string]*Material
}

// StringReader is input for the parser.
type StringReader interface {
	ReadString(delim byte) (string, error) // Example: bufio.Reader
}

// ReadMaterialLibFromBuf parses material lib from a buffer.
func ReadMaterialLibFromBuf(buf []byte, options *ObjParserOptions) (MaterialLib, error) {
	return readLib(bytes.NewBuffer(buf), options)
}

// ReadMaterialLibFromReader parses material lib from a reader.
func ReadMaterialLibFromReader(rd io.Reader, options *ObjParserOptions) (MaterialLib, error) {
	return readLib(bufio.NewReader(rd), options)
}

// ReadMaterialLibFromStringReader parses material lib from StringReader.
func ReadMaterialLibFromStringReader(rd StringReader, options *ObjParserOptions) (MaterialLib, error) {
	return readLib(rd, options)
}

// ReadMaterialLibFromFile parses material lib from a file.
func ReadMaterialLibFromFile(filename string, options *ObjParserOptions) (MaterialLib, error) {

	input, errOpen := os.Open(filename)
	if errOpen != nil {
		return NewMaterialLib(), errOpen
	}

	defer input.Close()

	return ReadMaterialLibFromReader(input, options)
}

// NewMaterialLib creates a new material lib.
func NewMaterialLib() MaterialLib {
	return MaterialLib{Lib: map[string]*Material{}}
}

// libParser holds auxiliary internal state for the parsing.
type libParser struct {
	currMaterial *Material
}

func readLib(reader StringReader, options *ObjParserOptions) (MaterialLib, error) {

	lineCount := 0

	parser := &libParser{}
	lib := NewMaterialLib()

	for {
		lineCount++
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// parse last line
			if _, e := parseLibLine(parser, lib, line, lineCount); e != nil {
				options.log(fmt.Sprintf("readLib: %v", e))
				return lib, e
			}
			break // EOF
		}

		if err != nil {
			// unexpected IO error
			return lib, fmt.Errorf("readLib: error: %v", err)
		}

		if fatal, e := parseLibLine(parser, lib, line, lineCount); e != nil {
			options.log(fmt.Sprintf("readLib: %v", e))
			if fatal {
				return lib, e
			}
		}
	}

	return lib, nil
}

func parseLibLine(p *libParser, lib MaterialLib, rawLine string, lineCount int) (bool, error) {
	line := strings.TrimSpace(rawLine)

	switch {
	case line == "" || line[0] == '#':
	case strings.HasPrefix(line, "newmtl "):

		newmtl := line[7:]
		var mat *Material
		var ok bool
		if mat, ok = lib.Lib[newmtl]; !ok {
			// create new material
			mat = &Material{Name: newmtl}
			lib.Lib[newmtl] = mat
		}
		p.currMaterial = mat

	case strings.HasPrefix(line, "Kd "):
		Kd := line[3:]

		if p.currMaterial == nil {
			return ErrNonFatal, fmt.Errorf("parseLibLine: %d undefined material for Kd=%s [%s]", lineCount, Kd, line)
		}

		color, err := parseFloatVector3Space(Kd)
		if err != nil {
			return ErrNonFatal, fmt.Errorf("parseLibLine: %d parsing error for Kd=%s [%s]: %v", lineCount, Kd, line, err)
		}

		p.currMaterial.Kd[0] = float32(color[0])
		p.currMaterial.Kd[1] = float32(color[1])
		p.currMaterial.Kd[2] = float32(color[2])

	case strings.HasPrefix(line, "map_Kd "):
		mapKd := line[7:]

		if p.currMaterial == nil {
			return ErrNonFatal, fmt.Errorf("parseLibLine: %d undefined material for map_Kd=%s [%s]", lineCount, mapKd, line)
		}

		p.currMaterial.MapKd = mapKd

	case strings.HasPrefix(line, "map_Ka "):
	case strings.HasPrefix(line, "map_d "):
	case strings.HasPrefix(line, "map_Bump "):
	case strings.HasPrefix(line, "Ns "):
	case strings.HasPrefix(line, "Ka "):
	case strings.HasPrefix(line, "Ke "):
	case strings.HasPrefix(line, "Ks "):
	case strings.HasPrefix(line, "Ni "):
	case strings.HasPrefix(line, "d "):
	case strings.HasPrefix(line, "illum "):
	case strings.HasPrefix(line, "Tf "):
	case strings.HasPrefix(line, "Tr "):
	default:
		return ErrNonFatal, fmt.Errorf("parseLibLine %v: [%v]: unexpected", lineCount, line)
	}

	return ErrNonFatal, nil
}

// Group holds parser result for a group.
type Group struct {
	Name       string
	Smooth     int
	Usemtl     string
	IndexBegin int
	IndexCount int
}

// Obj holds parser result for .obj file.
type Obj struct {
	Indices []int
	Coord   []float32 // vertex data pos=(x,y,z) tex=(tx,ty) norm=(nx,ny,nz)
	Mtllib  string
	Groups  []*Group

	BigIndexFound  bool // index larger than 65535
	TextCoordFound bool // texture coord
	NormCoordFound bool // normal coord

	StrideSize           int // (px,py,pz),(tu,tv),(nx,ny,nz) = 8 x 4-byte floats = 32 bytes max
	StrideOffsetPosition int // 0
	StrideOffsetTexture  int // 3 x 4-byte floats
	StrideOffsetNormal   int // 5 x 4-byte floats
}

// objParser holds auxiliary internal parser state.
type objParser struct {
	lineBuf    []string
	lineCount  int
	vertCoord  []float32
	textCoord  []float32
	normCoord  []float32
	currGroup  *Group
	indexTable map[string]int
	indexCount int
	vertLines  int
	textLines  int
	normLines  int
	faceLines  int // stat-only
	triangles  int // stat-only
}

// ObjParserOptions sets options for the parser.
type ObjParserOptions struct {
	LogStats      bool
	Logger        func(string)
	IgnoreNormals bool
}

func (opt *ObjParserOptions) log(msg string) {
	if opt.Logger == nil {
		return
	}
	opt.Logger(msg)
}

func (o *Obj) newGroup(name, usemtl string, begin int, smooth int) *Group {
	gr := &Group{Name: name, Usemtl: usemtl, IndexBegin: begin, Smooth: smooth}
	o.Groups = append(o.Groups, gr)
	return gr
}

// Coord64 gets vertex data as float64.
func (o *Obj) Coord64(i int) float64 {
	return float64(o.Coord[i])
}

// NumberOfElements gets the number of strides.
func (o *Obj) NumberOfElements() int {
	return 4 * len(o.Coord) / o.StrideSize
}

// VertexCoordinates gets vertex coordinates for a stride index.
func (o *Obj) VertexCoordinates(stride int) (float32, float32, float32) {
	offset := o.StrideOffsetPosition / 4
	floatsPerStride := o.StrideSize / 4
	f := offset + stride*floatsPerStride
	return o.Coord[f], o.Coord[f+1], o.Coord[f+2]
}

// ToFile saves OBJ to file.
func (o *Obj) ToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return o.ToWriter(f)
}

// ToWriter writes OBJ to writer stream.
func (o *Obj) ToWriter(w io.Writer) error {

	fmt.Fprintf(w, "# OBJ exported by gwob - https://github.com/udhos/gwob\n")
	fmt.Fprintf(w, "\n")

	if o.Mtllib != "" {
		fmt.Fprintf(w, "mtllib %s\n", o.Mtllib)
	}

	// write vertex data
	strides := o.NumberOfElements()
	for s := 0; s < strides; s++ {
		stride := s * o.StrideSize / 4
		v := stride + o.StrideOffsetPosition/4
		fmt.Fprintf(w, "v %f %f %f\n", o.Coord[v], o.Coord[v+1], o.Coord[v+2])

		if o.TextCoordFound {
			t := stride + o.StrideOffsetTexture/4
			fmt.Fprintf(w, "vt %f %f\n", o.Coord[t], o.Coord[t+1])
		}

		if o.NormCoordFound {
			n := stride + o.StrideOffsetNormal/4
			fmt.Fprintf(w, "vn %f %f %f\n", o.Coord[n], o.Coord[n+1], o.Coord[n+2])
		}
	}

	// write group faces
	for _, g := range o.Groups {
		if g.Name != "" {
			fmt.Fprintf(w, "g %s\n", g.Name)
		}
		if g.Usemtl != "" {
			fmt.Fprintf(w, "usemtl %s\n", g.Usemtl)
		}
		fmt.Fprintf(w, "s %d\n", g.Smooth)
		if g.IndexCount%3 != 0 {
			return fmt.Errorf("group=%s count=%d must be a multiple of 3", g.Name, g.IndexCount)
		}
		pastEnd := g.IndexBegin + g.IndexCount
		for s := g.IndexBegin; s < pastEnd; s += 3 {
			fmt.Fprintf(w, "f")
			for f := s; f < s+3; f++ {
				ff := o.Indices[f] + 1
				str := strconv.Itoa(ff)
				if o.TextCoordFound {
					if o.NormCoordFound {
						fmt.Fprintf(w, " %s/%s/%s", str, str, str)
					} else {
						fmt.Fprintf(w, " %s/%s", str, str)
					}
				} else {
					if o.NormCoordFound {
						fmt.Fprintf(w, " %s//%s", str, str)
					} else {
						fmt.Fprintf(w, " %s", str)
					}
				}
			}
			fmt.Fprintf(w, "\n")
		}
	}

	return nil
}

// NewObjFromVertex creates Obj from vertex data.
func NewObjFromVertex(objName string, coord []float32, indices []int) (*Obj, error) {
	o := &Obj{}

	group := o.newGroup("", "", 0, 0)

	o.Coord = append(o.Coord, coord...)
	for _, ind := range indices {
		pushIndex(group, o, ind)
	}

	setupStride(o)

	return o, nil
}

// NewObjFromBuf parses Obj from a buffer.
func NewObjFromBuf(objName string, buf []byte, options *ObjParserOptions) (*Obj, error) {
	return readObj(objName, bytes.NewBuffer(buf), options)
}

// NewObjFromReader parses Obj from a reader.
func NewObjFromReader(objName string, rd io.Reader, options *ObjParserOptions) (*Obj, error) {
	return readObj(objName, bufio.NewReader(rd), options)
}

// NewObjFromStringReader parses Obj from a StringReader.
func NewObjFromStringReader(objName string, rd StringReader, options *ObjParserOptions) (*Obj, error) {
	return readObj(objName, rd, options)
}

// NewObjFromFile parses Obj from a file.
func NewObjFromFile(filename string, options *ObjParserOptions) (*Obj, error) {

	input, errOpen := os.Open(filename)
	if errOpen != nil {
		return nil, errOpen
	}

	defer input.Close()

	return NewObjFromReader(filename, input, options)
}

func setupStride(o *Obj) {
	o.StrideSize = 3 * 4 // (px,py,pz) = 3 x 4-byte floats
	o.StrideOffsetPosition = 0
	o.StrideOffsetTexture = 0
	o.StrideOffsetNormal = 0

	if o.TextCoordFound {
		o.StrideOffsetTexture = o.StrideSize
		o.StrideSize += 2 * 4 // add (tu,tv) = 2 x 4-byte floats
	}

	if o.NormCoordFound {
		o.StrideOffsetNormal = o.StrideSize
		o.StrideSize += 3 * 4 // add (nx,ny,nz) = 3 x 4-byte floats
	}
}

func readObj(objName string, reader StringReader, options *ObjParserOptions) (*Obj, error) {

	if options == nil {
		options = &ObjParserOptions{LogStats: true, Logger: func(msg string) { fmt.Print(msg) }}
	}

	p := &objParser{indexTable: make(map[string]int)}
	o := &Obj{}

	// 1. vertex-only parsing
	if fatal, err := readLines(p, o, reader, options); err != nil {
		if fatal {
			return o, err
		}
	}

	p.faceLines = 0
	p.vertLines = 0
	p.textLines = 0
	p.normLines = 0

	// 2. full parsing
	if fatal, err := scanLines(p, o, reader, options); err != nil {
		if fatal {
			return o, err
		}
	}

	// 3. output

	// drop empty groups
	tmp := []*Group{}
	for _, g := range o.Groups {
		switch {
		case g.IndexCount < 0:
			continue // discard empty bogus group created internally by parser
		case g.IndexCount < 3:
			options.log(fmt.Sprintf("readObj: obj=%s BAD GROUP SIZE group=%s size=%d < 3", objName, g.Name, g.IndexCount))
		}
		tmp = append(tmp, g)
	}
	o.Groups = tmp

	setupStride(o) // setup stride size

	if options.LogStats {
		options.log(fmt.Sprintf("readObj: INPUT lines=%v vertLines=%v textLines=%v normLines=%v faceLines=%v triangles=%v",
			p.lineCount, p.vertLines, p.textLines, p.normLines, p.faceLines, p.triangles))

		options.log(fmt.Sprintf("readObj: STATS numberOfElements=%v indicesArraySize=%v", p.indexCount, len(o.Indices)))
		options.log(fmt.Sprintf("readObj: STATS bigIndexFound=%v groups=%v", o.BigIndexFound, len(o.Groups)))
		options.log(fmt.Sprintf("readObj: STATS textureCoordFound=%v normalCoordFound=%v", o.TextCoordFound, o.NormCoordFound))
		options.log(fmt.Sprintf("readObj: STATS stride=%v textureOffset=%v normalOffset=%v", o.StrideSize, o.StrideOffsetTexture, o.StrideOffsetNormal))
		for _, g := range o.Groups {
			options.log(fmt.Sprintf("readObj: GROUP name=%s first=%d count=%d", g.Name, g.IndexBegin, g.IndexCount))
		}
	}

	return o, nil
}

func readLines(p *objParser, o *Obj, reader StringReader, options *ObjParserOptions) (bool, error) {
	p.lineCount = 0

	for {
		p.lineCount++
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// parse last line
			if fatal, e := parseLineVertex(p, o, line, options); e != nil {
				options.log(fmt.Sprintf("readLines: %v", e))
				return fatal, e
			}
			break // EOF
		}

		if err != nil {
			// unexpected IO error
			return ErrFatal, fmt.Errorf("readLines: error: %v", err)
		}

		if fatal, e := parseLineVertex(p, o, line, options); e != nil {
			options.log(fmt.Sprintf("readLines: %v", e))
			if fatal {
				return fatal, e
			}
		}
	}

	return ErrNonFatal, nil
}

// parseLineVertex: parse only vertex lines
func parseLineVertex(p *objParser, o *Obj, rawLine string, options *ObjParserOptions) (bool, error) {
	line := strings.TrimSpace(rawLine)

	p.lineBuf = append(p.lineBuf, line) // save line for 2nd pass

	switch {
	case line == "" || line[0] == '#':
	case strings.HasPrefix(line, "s "):
	case strings.HasPrefix(line, "o "):
	case strings.HasPrefix(line, "g "):
	case strings.HasPrefix(line, "usemtl "):
	case strings.HasPrefix(line, "mtllib "):
	case strings.HasPrefix(line, "f "):
	case strings.HasPrefix(line, "vt "):

		tex := line[3:]
		t, err := parseFloatSliceSpace(tex)
		if err != nil {
			return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad vertex texture=[%s]: %v", p.lineCount, tex, err)
		}
		size := len(t)
		if size < 2 || size > 3 {
			return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad vertex texture=[%s] size=%d", p.lineCount, tex, size)
		}
		if size > 2 {
			if w := t[2]; !closeToZero(w) {
				options.log(fmt.Sprintf("parseLine: line=%d non-zero third texture coordinate w=%f: [%v]", p.lineCount, w, line))
			}
		}
		p.textCoord = append(p.textCoord, float32(t[0]), float32(t[1]))

	case strings.HasPrefix(line, "vn "):

		norm := line[3:]
		n, err := parseFloatVector3Space(norm)
		if err != nil {
			return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad vertex normal=[%s]: %v", p.lineCount, norm, err)
		}
		p.normCoord = append(p.normCoord, float32(n[0]), float32(n[1]), float32(n[2]))

	case strings.HasPrefix(line, "v "):

		result, err := parseFloatSliceSpace(line[2:])
		if err != nil {
			return ErrNonFatal, fmt.Errorf("parseLine %v: [%v]: error: %v", p.lineCount, line, err)
		}
		coordLen := len(result)
		switch coordLen {
		case 3:
			p.vertCoord = append(p.vertCoord, float32(result[0]), float32(result[1]), float32(result[2]))
		case 4:
			w := result[3]
			p.vertCoord = append(p.vertCoord, float32(result[0]/w), float32(result[1]/w), float32(result[2]/w))
		default:
			return ErrNonFatal, fmt.Errorf("parseLine %v: [%v]: bad number of coords: %v", p.lineCount, line, coordLen)
		}

	default:
		return ErrNonFatal, fmt.Errorf("parseLine %v: [%v]: unexpected", p.lineCount, line)
	}

	return ErrNonFatal, nil
}

func scanLines(p *objParser, o *Obj, reader StringReader, options *ObjParserOptions) (bool, error) {

	p.currGroup = o.newGroup("", "", 0, 0)

	p.lineCount = 0

	for _, line := range p.lineBuf {
		p.lineCount++

		if fatal, e := parseLine(p, o, line, options); e != nil {
			options.log(fmt.Sprintf("scanLines: %v", e))
			if fatal {
				return fatal, e
			}
		}
	}

	return ErrNonFatal, nil
}

func solveRelativeIndex(index, size int) int {
	if index > 0 {
		return index - 1
	}
	return size + index
}

func splitSlash(s string) []string {
	isSlash := func(c rune) bool {
		return c == '/'
	}

	return strings.FieldsFunc(s, isSlash)

}

func pushIndex(currGroup *Group, o *Obj, i int) {
	if i > 65535 {
		o.BigIndexFound = true
	}
	o.Indices = append(o.Indices, i)
	currGroup.IndexCount++
}

func addVertex(p *objParser, o *Obj, index string, options *ObjParserOptions) error {
	ind := splitSlash(strings.Replace(index, "//", "/0/", 1))
	size := len(ind)
	if size < 1 || size > 3 {
		return fmt.Errorf("addVertex: line=%d bad index=[%s] size=%d", p.lineCount, index, size)
	}

	v, err := strconv.ParseInt(ind[0], 10, 32)
	if err != nil {
		return fmt.Errorf("addVertex: line=%d bad integer 1st index=[%s]: %v", p.lineCount, ind[0], err)
	}
	vi := solveRelativeIndex(int(v), p.vertLines)

	var ti int
	var tIndex string
	hasTextureCoord := strings.Index(index, "//") == -1 && size > 1
	if hasTextureCoord {
		t, e := strconv.ParseInt(ind[1], 10, 32)
		if e != nil {
			return fmt.Errorf("addVertex: line=%d bad integer 2nd index=[%s]: %v", p.lineCount, ind[1], e)
		}
		ti = solveRelativeIndex(int(t), p.textLines)
		tIndex = strconv.Itoa(ti)
	}

	var ni int
	var nIndex string
	if size > 2 {
		n, e := strconv.ParseInt(ind[2], 10, 32)
		if e != nil {
			return fmt.Errorf("addVertex: line=%d bad integer 3rd index=[%s]: %v", p.lineCount, ind[2], e)
		}
		ni = solveRelativeIndex(int(n), p.normLines)
		nIndex = strconv.Itoa(ni)
	}

	absIndex := fmt.Sprintf("%d/%s/%s", vi, tIndex, nIndex)

	// known unified index?
	if i, ok := p.indexTable[absIndex]; ok {
		pushIndex(p.currGroup, o, i)
		return nil
	}

	vOffset := vi * 3
	if vOffset+2 >= len(p.vertCoord) {
		return fmt.Errorf("err: line=%d invalid vertex index=[%s]", p.lineCount, ind[0])
	}

	o.Coord = append(o.Coord, p.vertCoord[vOffset+0]) // x
	o.Coord = append(o.Coord, p.vertCoord[vOffset+1]) // y
	o.Coord = append(o.Coord, p.vertCoord[vOffset+2]) // z

	if tIndex != "" && hasTextureCoord {
		tOffset := ti * 2

		if tOffset+1 >= len(p.textCoord) {
			return fmt.Errorf("err: line=%d invalid texture index=[%s]", p.lineCount, ind[1])
		}

		o.Coord = append(o.Coord, p.textCoord[tOffset+0]) // u
		o.Coord = append(o.Coord, p.textCoord[tOffset+1]) // v
		o.TextCoordFound = true
	}

	if !options.IgnoreNormals && nIndex != "" {
		nOffset := ni * 3

		o.Coord = append(o.Coord, p.normCoord[nOffset+0]) // x
		o.Coord = append(o.Coord, p.normCoord[nOffset+1]) // y
		o.Coord = append(o.Coord, p.normCoord[nOffset+2]) // z

		o.NormCoordFound = true
	}

	// add unified index
	pushIndex(p.currGroup, o, p.indexCount)
	p.indexTable[absIndex] = p.indexCount
	p.indexCount++

	return nil
}

func smoothGroup(s string) (int, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	if s == "off" {
		return 0, nil
	}

	i, err := strconv.ParseInt(s, 0, 32)

	return int(i), err
}

func parseLine(p *objParser, o *Obj, line string, options *ObjParserOptions) (bool, error) {

	switch {
	case line == "" || line[0] == '#':
	case strings.HasPrefix(line, "s "):
		smooth := line[2:]
		if s, err := smoothGroup(smooth); err == nil {
			if p.currGroup.Smooth != s {
				// create new group
				p.currGroup = o.newGroup(p.currGroup.Name, p.currGroup.Usemtl, len(o.Indices), s)
			}
		} else {
			return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad boolean smooth=[%s]: %v: line=[%v]", p.lineCount, smooth, err, line)
		}
	case strings.HasPrefix(line, "o ") || strings.HasPrefix(line, "g "):
		name := line[2:]
		if p.currGroup.Name == "" {
			// only set missing name for group
			p.currGroup.Name = name
		} else if p.currGroup.Name != name {
			// create new group
			p.currGroup = o.newGroup(name, p.currGroup.Usemtl, len(o.Indices), p.currGroup.Smooth)
		}
	case strings.HasPrefix(line, "usemtl "):
		usemtl := line[7:]
		if p.currGroup.Usemtl == "" {
			// only set the missing material name for group
			p.currGroup.Usemtl = usemtl
		} else if p.currGroup.Usemtl != usemtl {
			if p.currGroup.IndexCount == 0 {
				// mark previous empty group as bogus
				p.currGroup.IndexCount = -1
			}
			// create new group for material
			p.currGroup = o.newGroup(p.currGroup.Name, usemtl, len(o.Indices), p.currGroup.Smooth)
		}
	case strings.HasPrefix(line, "mtllib "):
		mtllib := line[7:]
		if o.Mtllib != "" {
			options.log(fmt.Sprintf("parseLine: line=%d mtllib redefinition old=%s new=%s", p.lineCount, o.Mtllib, mtllib))
		}
		o.Mtllib = mtllib
	case strings.HasPrefix(line, "f "):
		p.faceLines++

		face := line[2:]
		f := strings.Fields(face)
		size := len(f)
		if size < 3 || size > 4 {
			return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad face=[%s] size=%d", p.lineCount, face, size)
		}
		// triangle face: v0 v1 v2
		// quad face:
		// v0 v1 v2 v3 =>
		// v0 v1 v2
		// v2 v3 v0
		p.triangles++
		if err := addVertex(p, o, f[0], options); err != nil {
			return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad face=[%s] index_v0=[%s]: %v", p.lineCount, face, f[0], err)
		}
		if err := addVertex(p, o, f[1], options); err != nil {
			return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad face=[%s] index_v1=[%s]: %v", p.lineCount, face, f[1], err)
		}
		if err := addVertex(p, o, f[2], options); err != nil {
			return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad face=[%s] index_v2=[%s]: %v", p.lineCount, face, f[2], err)
		}
		if size > 3 {
			// quad face
			p.triangles++
			if err := addVertex(p, o, f[2], options); err != nil {
				return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad face=[%s] index_v2=[%s]: %v", p.lineCount, face, f[2], err)
			}
			if err := addVertex(p, o, f[3], options); err != nil {
				return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad face=[%s] index_v3=[%s]: %v", p.lineCount, face, f[3], err)
			}
			if err := addVertex(p, o, f[0], options); err != nil {
				return ErrNonFatal, fmt.Errorf("parseLine: line=%d bad face=[%s] index_v0=[%s]: %v", p.lineCount, face, f[0], err)
			}
		}
	case strings.HasPrefix(line, "v "):
		p.vertLines++
	case strings.HasPrefix(line, "vt "):
		p.textLines++
	case strings.HasPrefix(line, "vn "):
		p.normLines++
	default:
		return ErrNonFatal, fmt.Errorf("parseLine %v: [%v]: unexpected", p.lineCount, line)
	}

	return ErrNonFatal, nil
}

func closeToZero(f float64) bool {
	return math.Abs(f-0) < 0.000001
}
