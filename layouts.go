package binarygen

import (
	"fmt"
	"go/types"
	"strings"
)

// This file defines the various type of layouts
// supported by the tool, from the simple fixed sized objects
// to the more complexe union types

type oType interface {
	// staticSize returns true for objects having a known,
	// fixed size, or false.
	staticSize() (int, bool)

	// name returns the Go type name,
	// used in the generated code.
	name() string

	// parser expression, with bounds check
	// <objectName>.<dstSelector>, read, err = parse(<byteSliceName[<offsetName>:])
	parser(cc codeContext, dstSelector string) []string
}

type withConstructor struct {
	name_ string
	size_ int

	isMethod bool // fromUint(), toUint() vs xxxFromUint() xxxtoUint()
}

func (wc withConstructor) size() int {
	return wc.size_
}

type basicType struct {
	name_ string // named type

	binarySize int // underlying
}

func (bt basicType) size() int { return bt.binarySize }

// integer offset to the actual data
type offset struct {
	target oType

	size_ int
}

func (o offset) size() int {
	return o.size_
}

func (offset) offsetVariableName(field string) string {
	return fmt.Sprintf("offsetTo%s", strings.Title(field))
}

// union is an union of types, identified by a tag
//
// It is defined in the source code by an interface <itfName>
// and tags with type following the convention <itfName>Kind
// Each concrete type must be named <itfName><concreteName>,
// and the flags value must be named <itfName>Kind<concreteName>
type union struct {
	type_ *types.Named

	// the possible flag values, in the same order as `members`
	flags []*types.Const
	// the possible types
	members []structLayout

	// the variable to read to get the flag
	flagFieldName string
}

func (of offset) name() string          { return of.target.name() }
func (wc withConstructor) name() string { return wc.name_ }
func (bt basicType) name() string       { return bt.name_ }
func (sl slice) name() string           { return "[]" + sl.element.name() }
func (sl structLayout) name() string    { return sl.name_ }
func (sl union) name() string           { return sl.type_.Obj().Name() }

func (of offset) staticSize() (int, bool)          { return of.size_, true }
func (wc withConstructor) staticSize() (int, bool) { return wc.size_, true }
func (bt basicType) staticSize() (int, bool)       { return bt.binarySize, true }
func (sl slice) staticSize() (int, bool)           { return 0, false }
func (u union) staticSize() (int, bool)            { return 0, false }

// staticSize returns the statically known size of the type
// or false if it is dynamic or requires additional length check
func (sl structLayout) staticSize() (int, bool) {
	totalSize := 0
	for _, field := range sl.fields {
		// special case for offsets : they have a static size
		// but still require additional length check
		if _, isOffset := field.type_.(offset); isOffset {
			return 0, false
		}

		size, ok := field.type_.staticSize()
		if !ok {
			return 0, false
		}
		totalSize += size
	}
	return totalSize, true
}
