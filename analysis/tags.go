package analysis

import (
	"go/ast"
	"reflect"
	"strings"
)

// parsedTags is the result of parsing a field tag string
type parsedTags struct {
	arrayCountField string // used by [ComputedField]
	arrayCount      ArrayCount

	subsliceStart SubsliceStart

	// isCustom is true if the field has
	// a custom parser/writter
	isCustom bool
}

func newTags(tags reflect.StructTag) (out parsedTags) {
	switch tag := tags.Get("subsliceStart"); tag {
	case "AtStart":
		out.subsliceStart = AtStart
	case "", "AtCurrent":
		out.subsliceStart = AtCurrent
	default:
		panic("invalic tag for SubsliceStart : " + tag)
	}

	switch tag := tags.Get("arrayCount"); tag {
	case "FirstUint16":
		out.arrayCount = FirstUint16
	case "FirstUint32":
		out.arrayCount = FirstUint32
	default:
		if _, field, hasComputedField := strings.Cut(tag, "computedField-"); hasComputedField {
			out.arrayCount = ComputedField
			out.arrayCountField = field
		} else if tag == "" {
			// default to NoLength
			out.arrayCount = NoLength
		} else {
			panic("invalid tag for ArrayCount: " + tag)
		}
	}

	_, out.isCustom = tags.Lookup("isCustom")

	return out
}

type commments struct {
	// startingOffset may be provided if the type parsing/writting function
	// expect its input slice not to start
	// at the begining of the type data.
	// If empty, it default to 0 (the begining of the subslice).
	startingOffset string
}

// parse the type documentation looking for special comments
// of the following form :
func parseComments(doc *ast.CommentGroup) (out commments) {
	if doc == nil {
		return out
	}
	for _, comment := range doc.List {
		if _, value, ok := strings.Cut(comment.Text, " binarygen:"); ok {
			if _, so, ok := strings.Cut(value, "startOffset="); ok {
				out.startingOffset = so
			}
		}
	}
	return out
}

// Layout provides additionnal information about how
// a struct field is written in binary files.
// For simple scalar field, it is usually empty since the
// Go type provides enough information.
//
// It is defined in the Go source files using struct tags.
type Layout struct {
	// StartingOffset may be provided if the field does not start
	// at the begining of the subslice passed to the type parsing/writting function.
	// It empty, it default to 0 (the begining of the subslice).
	StartingOffset string

	ArrayCountField string // used by [ComputedField]
	ArrayCount      ArrayCount

	SubsliceStart SubsliceStart
}

// ArrayCount defines how the number of elements in an array is defined
type ArrayCount uint8

const (
	// The length must be provided by the context and is not found in the binary
	NoLength ArrayCount = iota

	// The length is written at the start of the array, as an uint16
	FirstUint16
	// The length is written at the start of the array, as an uint32
	FirstUint32

	// The length is deduced from an other field, parsed previously,
	// or computed by a method
	ComputedField
)

// SubsliceStart indicates where the start of the subslice
// given to the field parsing function shall be computed
type SubsliceStart uint8

const (
	// The current slice is sliced at the current offset for the field
	AtCurrent SubsliceStart = iota
	// The current slice is not resliced
	AtStart
)