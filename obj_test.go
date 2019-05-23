package gwob

import (
	"bytes"
	"fmt"
	"testing"
)

func BenchmarkCube1(b *testing.B) {
	buf := []byte(cubeObj)
	options := &ObjParserOptions{}
	for i := 0; i < b.N; i++ {
		NewObjFromBuf("cubeObj", buf, options)
	}
}

func BenchmarkRelativeIndex1(b *testing.B) {
	buf := []byte(relativeObj)
	options := &ObjParserOptions{}
	for i := 0; i < b.N; i++ {
		NewObjFromBuf("relativeObj", buf, options)
	}
}

func BenchmarkForwardVertex1(b *testing.B) {
	buf := []byte(forwardObj)
	options := &ObjParserOptions{}
	for i := 0; i < b.N; i++ {
		NewObjFromBuf("forwardObj", buf, options)
	}
}

const LogStats = false

func expectInt(t *testing.T, label string, want, got int) {
	if want != got {
		t.Errorf("%s: want=%d got=%d", label, want, got)
	}
}

func sliceEqualInt(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func sliceEqualFloat(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func TestCube(t *testing.T) {

	options := ObjParserOptions{LogStats: LogStats, Logger: func(msg string) { fmt.Printf("TestCube NewObjFromBuf: log: %s\n", msg) }}

	o, err := NewObjFromBuf("cubeObj", []byte(cubeObj), &options)
	if err != nil {
		t.Errorf("TestCube: NewObjFromBuf: %v", err)
		return
	}

	if !sliceEqualInt(cubeIndices, o.Indices) {
		t.Errorf("TestCube: indices: want=%v got=%v", cubeIndices, o.Indices)
	}

	if !sliceEqualFloat(cubeCoord, o.Coord) {
		t.Errorf("TestCube: coord: want=%d%v got=%d%v", len(cubeCoord), cubeCoord, len(o.Coord), o.Coord)
	}

	if o.StrideSize != cubeStrideSize {
		t.Errorf("TestCube: stride size: want=%d got=%d", cubeStrideSize, o.StrideSize)
	}

	if o.StrideOffsetPosition != cubeStrideOffsetPosition {
		t.Errorf("TestCube: stride offset position: want=%d got=%d", cubeStrideOffsetPosition, o.StrideOffsetPosition)
	}

	if o.StrideOffsetTexture != cubeStrideOffsetTexture {
		t.Errorf("TestCube: stride offset texture: want=%d got=%d", cubeStrideOffsetTexture, o.StrideOffsetTexture)
	}

	if o.StrideOffsetNormal != cubeStrideOffsetNormal {
		t.Errorf("TestCube: stride offset normal: want=%d got=%d", cubeStrideOffsetNormal, o.StrideOffsetNormal)
	}
}

func TestWriteEmpty(t *testing.T) {

	// load
	options := ObjParserOptions{LogStats: LogStats, Logger: func(msg string) { fmt.Printf("TestWriteEmpty NewObjFromBuf: log: %s\n", msg) }}
	orig, err := NewObjFromBuf("empty", []byte{}, &options)
	if err != nil {
		t.Errorf("TestWriteEmpty: NewObjFromBuf: %v", err)
		return
	}

	// export
	buf := bytes.Buffer{}
	errWrite := orig.ToWriter(&buf)
	if errWrite != nil {
		t.Errorf("TestWriteEmpty: ToWriter: %v", errWrite)
		return
	}

	// reload
	_, errParse := NewObjFromReader("emtpy-reload", &buf, &options)
	if errParse != nil {
		t.Errorf("TestWriteEmpty: NewObjFromReader: %v", errParse)
		return
	}
}

func TestWriteBad(t *testing.T) {

	// load
	orig, err := NewObjFromVertex("bad", []float32{}, []int{0})
	if err != nil {
		t.Errorf("TestWriteBad: NewObjFromVertex: %v", err)
		return
	}

	// export
	buf := bytes.Buffer{}
	errWrite := orig.ToWriter(&buf)
	if errWrite == nil {
		t.Errorf("TestWriteBad: unexpected writer success for bad group index count (non multiple of 3)")
		return
	}

}

func TestCubeWrite(t *testing.T) {

	// load cube
	options := ObjParserOptions{LogStats: LogStats, Logger: func(msg string) { fmt.Printf("TestCube NewObjFromBuf: log: %s\n", msg) }}
	orig, err := NewObjFromBuf("cube-orig", []byte(cubeObj), &options)
	if err != nil {
		t.Errorf("TestCubeWrite: NewObjFromBuf: %v", err)
		return
	}

	// export cube
	buf := bytes.Buffer{}
	errWrite := orig.ToWriter(&buf)
	if errWrite != nil {
		t.Errorf("TestCubeWrite: ToWriter: %v", errWrite)
		return
	}

	// reload cube
	o, errParse := NewObjFromReader("cube-reload", &buf, &options)
	if errParse != nil {
		t.Errorf("TestCubeWrite: NewObjFromReader: %v", errParse)
		return
	}

	if !sliceEqualInt(cubeIndices, o.Indices) {
		t.Errorf("TestCubeWrite: indices: want=%v got=%v", cubeIndices, o.Indices)
	}

	if !sliceEqualFloat(cubeCoord, o.Coord) {
		t.Errorf("TestCubeWrite: coord: want=%d%v got=%d%v", len(cubeCoord), cubeCoord, len(o.Coord), o.Coord)
	}

	if o.StrideSize != cubeStrideSize {
		t.Errorf("TestCubeWrite: stride size: want=%d got=%d", cubeStrideSize, o.StrideSize)
	}

	if o.StrideOffsetPosition != cubeStrideOffsetPosition {
		t.Errorf("TestCubeWrite: stride offset position: want=%d got=%d", cubeStrideOffsetPosition, o.StrideOffsetPosition)
	}

	if o.StrideOffsetTexture != cubeStrideOffsetTexture {
		t.Errorf("TestCubeWrite: stride offset texture: want=%d got=%d", cubeStrideOffsetTexture, o.StrideOffsetTexture)
	}

	if o.StrideOffsetNormal != cubeStrideOffsetNormal {
		t.Errorf("TestCubeWrite: stride offset normal: want=%d got=%d", cubeStrideOffsetNormal, o.StrideOffsetNormal)
	}
}

func TestRelativeIndex(t *testing.T) {

	options := ObjParserOptions{LogStats: LogStats, Logger: func(msg string) { fmt.Printf("TestRelativeIndex NewObjFromBuf: log: %s\n", msg) }}

	o, err := NewObjFromBuf("relativeObj", []byte(relativeObj), &options)
	if err != nil {
		t.Errorf("TestRelativeIndex: NewObjFromBuf: %v", err)
		return
	}

	//indices := o.Indices[:len(o.Indices):len(o.Indices)]
	if !sliceEqualInt(relativeIndices, o.Indices) {
		t.Errorf("TestRelativeIndex: indices: want=%v got=%v", relativeIndices, o.Indices)
	}

	//coord := o.Coord[:len(o.Coord):len(o.Coord)]
	if !sliceEqualFloat(relativeCoord, o.Coord) {
		t.Errorf("TestRelativeIndex: coord: want=%v got=%v", relativeCoord, o.Coord)
	}
}

func TestForwardVertex(t *testing.T) {

	options := ObjParserOptions{LogStats: LogStats, Logger: func(msg string) { fmt.Printf("TestForwardVertex NewObjFromBuf: log: %s\n", msg) }}

	o, err := NewObjFromBuf("forwardObj", []byte(forwardObj), &options)
	if err != nil {
		t.Errorf("TestForwardVertex: NewObjFromBuf: %v", err)
		return
	}

	if !sliceEqualInt(forwardIndices, o.Indices) {
		t.Errorf("TestForwardVertex: indices: want=%v got=%v", forwardIndices, o.Indices)
	}

	if !sliceEqualFloat(forwardCoord, o.Coord) {
		t.Errorf("TestForwardVertex: coord: want=%v got=%v", forwardCoord, o.Coord)
	}
}

func TestMisc(t *testing.T) {
	str := `
mtllib lib1

usemtl mtl1
usemtl mtl2

s off
s 1
`

	options := ObjParserOptions{LogStats: LogStats, Logger: func(msg string) { fmt.Printf("TestMisc NewObjFromBuf: log: %s\n", msg) }}

	NewObjFromBuf("TestMisc local str obj", []byte(str), &options)
}

func TestSkippedUV1(t *testing.T) {

	options := ObjParserOptions{LogStats: LogStats, Logger: func(msg string) { fmt.Printf("TestSkippedUV NewObjFromBuf: log: %s\n", msg) }}

	o, err := NewObjFromBuf("skippedUV", []byte(skippedUVObj), &options)
	if err != nil {
		t.Errorf("TestSkippedUV: NewObjFromBuf: %v", err)
		return
	}

	if !sliceEqualInt(skippedUVIndices, o.Indices) {
		t.Errorf("TestSkippedUV: indices: want=%v got=%v", skippedUVIndices, o.Indices)
	}

	if !sliceEqualFloat(skippedUVCoord, o.Coord) {
		t.Errorf("TestSkippedUV: coord: want=%v got=%v", skippedUVCoord, o.Coord)
	}
}

func TestSkippedUV2(t *testing.T) {

	options := ObjParserOptions{LogStats: LogStats, Logger: func(msg string) { fmt.Printf("TestSkippedUV2 NewObjFromBuf: log: %s\n", msg) }}

	o, err := NewObjFromBuf("skippedUV2", []byte(skippedUV2Obj), &options)
	if err != nil {
		t.Errorf("TestSkippedUV2: NewObjFromBuf: %v", err)
		return
	}

	if !sliceEqualInt(skippedUV2Indices, o.Indices) {
		t.Errorf("TestSkippedUV2: indices: want=%v got=%v", skippedUV2Indices, o.Indices)
	}

	if !sliceEqualFloat(skippedUV2Coord, o.Coord) {
		t.Errorf("TestSkippedUV2: coord: want=%v got=%v", skippedUV2Coord, o.Coord)
	}
}

var cubeStrideSize = 32
var cubeStrideOffsetPosition = 0
var cubeStrideOffsetTexture = 12
var cubeStrideOffsetNormal = 20
var cubeIndices = []int{0, 1, 2, 2, 3, 0, 4, 5, 6, 6, 7, 4, 8, 9, 10, 10, 11, 8, 12, 13, 14, 14, 15, 12, 16, 17, 18, 18, 19, 16, 20, 21, 22, 22, 23, 20}
var cubeCoord = []float32{1, -1, 1, 0.5, 0, 0, -1, 0, -1, -1, 1, 0.5, 0, 0, -1, 0, -1, -1, -1, 0.5, 0, 0, -1, 0, 1, -1, -1, 0.5, 0, 0, -1, 0, 1, 1, -1, 0.5, 0, 0, 1, 0, -1, 1, -1, 0.5, 0, 0, 1, 0, -1, 1, 1, 0.5, 0, 0, 1, 0, 1, 1, 1, 0.5, 0, 0, 1, 0, 1, -1, -1, 0, 0, 1, 0, 0, 1, 1, -1, 0, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 0, 0, 1, -1, 1, 0, 0, 1, 0, 0, -1, -1, 1, 0, 0, -1, 0, 0, -1, 1, 1, 0, 0, -1, 0, 0, -1, 1, -1, 0, 0, -1, 0, 0, -1, -1, -1, 0, 0, -1, 0, 0, 1, -1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 1, -1, 1, 1, 1, 0, 0, 0, 1, -1, -1, 1, 1, 0, 0, 0, 1, -1, -1, -1, 1, 0, 0, 0, -1, -1, 1, -1, 1, 0, 0, 0, -1, 1, 1, -1, 1, 0, 0, 0, -1, 1, -1, -1, 1, 0, 0, 0, -1}

var relativeIndices = []int{0, 1, 2, 0, 1, 2, 3, 4, 5, 3, 4, 5, 0, 1, 2, 0, 1, 2}
var relativeCoord = []float32{1.0, 1.0, 1.0, 2.0, 2.0, 2.0, 3.0, 3.0, 3.0, 4.0, 4.0, 4.0, 5.0, 5.0, 5.0, 6.0, 6.0, 6.0}

var forwardIndices = []int{0, 1, 2}
var forwardCoord = []float32{1.0, 1.0, 1.0, 2.0, 2.0, 2.0, 3.0, 3.0, 3.0}

var cubeObj = `
# texture_cube.obj

mtllib texture_cube.mtl

o cube

# square bottom
v -1 -1 -1
v -1 -1 1
v 1 -1 1
v 1 -1 -1

# square top
v -1 1 -1
v -1 1 1
v 1 1 1
v 1 1 -1

# uv coord

# red -3
vt 0 0

# green -2
vt .5 0

# blue -1
vt 1 0

# normal coord

# down -6
vn 0 -1 0

# up -5
vn 0 1 0

# right -4
vn 1 0 0

# left -3
vn -1 0 0

# front -2
vn 0 0 1

# back -1
vn 0 0 -1

usemtl 3-pixel-rgb

# face down (green -2)
f -6/-2/-6 -7/-2/-6 -8/-2/-6
f -8/-2/-6 -5/-2/-6 -6/-2/-6

# face up (green -2)
f -1/-2/-5 -4/-2/-5 -3/-2/-5
f -3/-2/-5 -2/-2/-5 -1/-2/-5 

# face right (red -3)
f -5/-3/-4 -1/-3/-4 -2/-3/-4
f -2/-3/-4 -6/-3/-4 -5/-3/-4

# face left (red -3)
f -7/-3/-3 -3/-3/-3 -4/-3/-3
f -4/-3/-3 -8/-3/-3 -7/-3/-3

# face front (blue -1)
f -6/-1/-2 -2/-1/-2 -3/-1/-2
f -3/-1/-2 -7/-1/-2 -6/-1/-2

# face back (blue -1)
f -8/-1/-1 -4/-1/-1 -1/-1/-1
f -1/-1/-1 -5/-1/-1 -8/-1/-1
`

var relativeObj = `
o relative_test
v 1 1 1
v 2 2 2
v 3 3 3
f 1 2 3
# this line should affect indices, but not vertex array
f -3 -2 -1
v 4 4 4
v 5 5 5
v 6 6 6
f 4 5 6
# this line should affect indices, but not vertex array
f -3 -2 -1
# these lines should affect indices, but not vertex array
f 1 2 3
f -6 -5 -4
`

var forwardObj = `
o forward_vertices_test
# face pointing to forward vertex definitions
# support for this isn't usual in OBJ parsers
# since it requires multiple passes
# but currently we do support this layout
f 1 2 3
v 1 1 1
v 2 2 2
v 3 3 3
`

var skippedUVObj = `

o skipped_uv

v 1 1 1
v 2 2 2
v 3 3 3

vn 1 0 0
vn 0 1 0
vn 0 0 1

f 1//1 2//2 3//3 
`

var skippedUVIndices = []int{0, 1, 2}
var skippedUVCoord = []float32{1, 1, 1, 1, 0, 0, 2, 2, 2, 0, 1, 0, 3, 3, 3, 0, 0, 1}

var skippedUV2Obj = `

o skipped_uv

v 1 1 1
v 2 2 2
v 3 3 3

vt 0 0
vt .5 .5
vt 1 1

vn 1 0 0
vn 0 1 0
vn 0 0 1

f 1//1 2//2 3//3 
`

var skippedUV2Indices = []int{0, 1, 2}
var skippedUV2Coord = []float32{1, 1, 1, 1, 0, 0, 2, 2, 2, 0, 1, 0, 3, 3, 3, 0, 0, 1}
