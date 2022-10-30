package testpackage

import "math"

type withFixedSize struct {
	a, b, c int32
	d       uint32
	e       int64
	g, h    byte
	t       tag
	v       float214
	w       fl32
	array1  [5]byte
	array2  [5]uint16
}

type varInstance struct {
	Coords    []fl1616 `arrayCount:""`
	Coords2   []fl1616 `arrayCount:""`
	Subfamily uint16

	PSStringID uint16 `bin:"optional"`
}

type varInstanceContainer struct {
	inst varInstance
}

type embeded struct {
	a, b byte
	c    []uint16 `arrayCount:"FirstUint16"`
}

type Composed2 struct {
	a, b, c byte
	embeded
}

type composed struct {
	a withFixedSize
	b withFixedSize
}

type simpleSubtable struct {
	version uint16
	x, y    int16
	lookups []withFixedSize `arrayCount:"FirstUint16"`
	array2  []uint32        `arrayCount:"FirstUint16"`
}

type ComplexeSubtable struct {
	version uint16
	x, y    int16
	lookups []withFixedSize `arrayCount:"FirstUint16"`
	u, v    float214
	a, b, c int64
	array2  []uint32 `arrayCount:"FirstUint32"`
	array3  []fl32   `arrayCount:"FirstUint32"`
	opaque  []byte   `isOpaque:""`
	rawData []byte   `arrayCount:"ToEnd" subsliceStart:"AtStart"`
}

type arrayLike struct {
	size   uint16
	datas  []uint16        `arrayCount:"ComputedField-size"`
	array  []withFixedSize `arrayCount:"FirstUint16"`
	array2 []composed      `arrayCount:"FirstUint16"`
	data   []byte          `arrayCount:"ToEnd"`
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

// other constants not interpreted as flags

type flagNotVersion_ uint

const _dummy1 = ""

const _dummy2 = 2

const _dummy3 flagNotVersion_ = 8
