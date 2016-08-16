package gwob

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func parseFloatSlice(list []string) ([]float64, error) {
	result := make([]float64, len(list))

	for i, j := range list {
		j = strings.TrimSpace(j)
		var err error
		if result[i], err = strconv.ParseFloat(j, 64); err != nil {
			return nil, fmt.Errorf("parseFloatSlice: list=[%v] elem[%v]=[%s] failure: %v", list, i, j, err)
		}
	}

	return result, nil
}

func parseFloatSliceFunc(text string, f func(rune) bool) ([]float64, error) {
	return parseFloatSlice(strings.FieldsFunc(text, f))
}

func parseFloatSliceSpace(text string) ([]float64, error) {
	return parseFloatSliceFunc(text, unicode.IsSpace)
}

func parseFloatVectorFunc(text string, size int, f func(rune) bool) ([]float64, error) {
	list := strings.FieldsFunc(text, f)
	if s := len(list); s != size {
		return nil, fmt.Errorf("parseFloatVectorFunc: text=[%v] size=%v must be %v", text, s, size)
	}

	return parseFloatSlice(list)
}

func parseFloatVectorSpace(text string, size int) ([]float64, error) {
	return parseFloatVectorFunc(text, size, unicode.IsSpace)
}

func parseFloatVectorComma(text string, size int) ([]float64, error) {
	isComma := func(c rune) bool {
		return c == ','
	}

	return parseFloatVectorFunc(text, size, isComma)
}

func parseFloatVector3Space(text string) ([]float64, error) {
	return parseFloatVectorSpace(text, 3)
}

func parseFloatVector3Comma(text string) ([]float64, error) {
	return parseFloatVectorComma(text, 3)
}
