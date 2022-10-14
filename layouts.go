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
	parser(cc codeContext, dstSelector string) string
}

// to be constistent a type returning a static size
// must implement this interface
type fixedSizeType interface {
	oType

	mustParser(cc codeContext, dstSelector string) string
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

// fixed size array [...]<element>
type array struct {
	element basicType
	length  int
}

// []<element> slice type
type slice struct {
	element        oType
	lengthLocation string
}

// union is an union of types, identified by a tag
//
// It is defined in the source code by an interface <itfName>
// and tags with type following the convention Version<concreteName>
// Each concrete type must be named <itfName><concreteName>,
// and the flags value must be named <...>Version<concreteName>
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
func (ar array) name() string           { return fmt.Sprintf("[%d]%s", ar.length, ar.element.name()) }
func (sl slice) name() string           { return "[]" + sl.element.name() }
func (st structLayout) name() string    { return st.name_ }
func (u union) name() string            { return u.type_.Obj().Name() }

func (of offset) staticSize() (int, bool)          { return of.size_, true }
func (wc withConstructor) staticSize() (int, bool) { return wc.size_, true }
func (bt basicType) staticSize() (int, bool)       { return bt.binarySize, true }
func (ar array) staticSize() (int, bool)           { return ar.length * ar.element.binarySize, true }
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
