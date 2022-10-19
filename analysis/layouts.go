package analysis

import "go/types"

// BinarySize indicates how many bytes
// are needed to store a value
type BinarySize int

const (
	Byte BinarySize = 1 << iota
	Uint16
	Uint32
	Uint64
)

func newBinarySize(t *types.Basic) (BinarySize, bool) {
	switch t.Kind() {
	case types.Bool, types.Int8, types.Uint8:
		return Byte, true
	case types.Int16, types.Uint16:
		return Uint16, true
	case types.Int32, types.Uint32, types.Float32:
		return Uint32, true
	case types.Int64, types.Uint64, types.Float64:
		return Uint64, true
	default:
		return 0, false
	}
}

// Type is the common interface for struct field types
// supported by the package,
// describing the binary layout of a type.
type Type interface {
	// Origin returns the Go type yielding the type
	Origin() types.Type

	// IsFixedSize returns the number of byte needed to store an element,
	// or false if it is not known at compile time.
	IsFixedSize() (BinarySize, bool)
}

// ---------------------------- Concrete types ----------------------------

func (t Struct) Origin() types.Type           { return t.origin }
func (t Basic) Origin() types.Type            { return t.origin }
func (t DerivedFromBasic) Origin() types.Type { return t.origin }
func (t Offset) Origin() types.Type           { return t.target.Origin() }
func (t Array) Origin() types.Type            { return t.origin }
func (t Slice) Origin() types.Type            { return t.origin }
func (t Union) Origin() types.Type            { return t.origin }
func (t Opaque) Origin() types.Type           { return t.origin }

// Struct defines the the binary layout
// of a struct
type Struct struct {
	origin *types.Named
	Fields []Field
}

// Field is a struct field.
// Embeded fields are not resolved.
type Field struct {
	Type   Type
	Layout Layout
	Name   string
}

// IsFixedSize returns true if all the fields have fixed size.
func (st Struct) IsFixedSize() (BinarySize, bool) {
	var totalSize BinarySize
	for _, field := range st.Fields {
		size, ok := field.Type.IsFixedSize()
		if !ok {
			return 0, false
		}
		totalSize += size
	}
	return totalSize, true
}

// Basic is a fixed size type, directly
// convertible from and to uintXX
type Basic struct {
	origin types.Type // may be named, but with underlying Basic
}

func (ba Basic) IsFixedSize() (BinarySize, bool) {
	return newBinarySize(ba.origin.Underlying().(*types.Basic))
}

// DerivedFromBasic is stored as a an uintXX, but
// uses custom constructor to perform the convertion :
// <typeString>FromUintXX ; <typeString>ToUintXX
type DerivedFromBasic struct {
	origin types.Type // may be named, but with underlying Basic

	// For aliases, it is the name of the defined (not the "underlying" type)
	// For named types, the name of the defined type
	// Otherwise, it is the string representation
	name string

	// the size as read and written in binary files
	size BinarySize
}

func (de DerivedFromBasic) IsFixedSize() (BinarySize, bool) {
	return de.size, true
}

// Offset is a fixed size integer pointing to
// an other type.
type Offset struct {
	target Type
	size   BinarySize // of the offset field
}

// IsFixedSize returns `false`, since, even if the offset itself has a fixed size,
// the whole data has not and requires additional length check.
func (Offset) IsFixedSize() (BinarySize, bool) { return 0, false }

// Array is a fixed length array.
type Array struct {
	origin types.Type

	// Len is the length of elements in the array
	Len int

	// Elem is the type of the element
	Elem Type
}

func (ar Array) IsFixedSize() (BinarySize, bool) {
	elementSize, isElementFixed := ar.Elem.IsFixedSize()
	if !isElementFixed {
		return 0, false
	}
	return BinarySize(ar.Len) * elementSize, true
}

// Slice is a variable size array
type Slice struct {
	origin types.Type

	// Elem is the type of the element
	Elem Type

	// Count indicates how to read/write the length of the array
	Count ArrayCount
	// CountExpr is used when [Count] is [ComputedField]
	CountExpr string
}

func (Slice) IsFixedSize() (BinarySize, bool) { return 0, false }

// IsRawData returns true for []byte
func (sl Slice) IsRawData() bool {
	elem := sl.Elem.Origin().Underlying()
	if basic, isBasic := elem.(*types.Basic); isBasic {
		return basic.Kind() == types.Byte
	}
	return false
}

// Union represents an union of several types,
// which are identified by constant flags.
type Union struct {
	origin *types.Named // with underlying type Interface

	// Flags are the possible flag values, in the same order as `Members`
	Flags []*types.Const

	// Members stores the possible members
	Members []Struct

	// FlagField is the struct field indicating which
	// member is to be read
	FlagField string
}

func (Union) IsFixedSize() (BinarySize, bool) { return 0, false }

// Opaque represents a type with no binary structure.
// The parsing and writting step will be replaced by placeholder methods.
type Opaque struct {
	origin types.Type
}

func (Opaque) IsFixedSize() (BinarySize, bool) { return 0, false }
