package testpackage

import "math"

type VarInstance struct {
	Coords    []fl1616 `len-size:"-"`
	Coords2   []fl1616 `len-size:"-"`
	Subfamily uint16

	PSStringID uint16 `bin:"optional"`
}

type varInstanceContainer struct {
	inst VarInstance
}

type lookup struct {
	a, b, c int32
	d       uint32
	e       int64
	g, h    byte
	t       tag
	v       float214 `bin:"optional"`
	w       fl32
	array1  [5]byte
	array2  [5]uint16
}

type embeded struct {
	a, b byte
	c    []uint16 `len-size:"16"`
}

type composed2 struct {
	a, b, c byte
	embeded
}

type composed struct {
	a lookup
	b lookup
}

type simpleSubtable struct {
	version uint16
	x, y    int16
	lookups []lookup `len-size:"16"`
	array2  []uint32 `len-size:"16"`
}

type complexeSubtable struct {
	version uint16
	x, y    int16
	lookups []lookup `len-size:"16"`
	u, v    float214
	a, b, c int64
	array2  []uint32 `len-size:"32"`
	array3  []fl32   `len-size:"64"`
}

type arrayLike struct {
	array  []lookup   `len-size:"16"`
	array2 []composed `len-size:"16"`
}

type tag uint32

type float214 float32 // representated as 2.14 fixed point

func (f *float214) fromUint(v uint16) {
	*f = float214(math.Float32frombits(uint32(v)))
}

func (f float214) toUint() uint16 {
	return uint16(math.Float32bits(float32(f)))
}

type fl32 = float32

func fl32FromUint(v uint32) fl32 {
	return math.Float32frombits(uint32(v))
}

func fl32ToUint(f fl32) uint32 {
	return math.Float32bits(f)
}

type fl1616 = float32

func fl1616FromUint(v uint32) fl1616 {
	// value are actually signed integers
	return fl1616(int32(v)) / (1 << 16)
}

func fl1616ToUint(f fl1616) uint32 {
	return uint32(int32(f * (1 << 16)))
}

type withOffset struct {
	version        uint16
	offsetToSlice  []uint64 `offset-size:"32" len-size:"16"`
	offsetToStruct lookup   `offset-size:"32"`
	a, b, c        byte
}

type withUnion struct {
	version    subtableKind
	otherField byte
	data       subtable `kind-field:"version"`
}

type subtableKind uint16

const (
	subtableKind1 subtableKind = iota
	subtableKind2
)

type subtable interface {
	isSubtable()
}

type subtable1 struct {
	F uint64
}
type subtable2 struct {
	F uint8
}

func (subtable1) isSubtable() {}
func (subtable2) isSubtable() {}
