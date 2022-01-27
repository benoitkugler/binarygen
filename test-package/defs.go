package testpackage

import "math"

type lookup struct {
	a, b, c int32
	d       uint32
	e       int64
	g, h    byte
	t       tag
	v       float32
	w       fl32
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

func fl32FromUint(v uint16) fl32 {
	return math.Float32frombits(uint32(v))
}

func fl32ToUint(f fl32) uint16 {
	return uint16(math.Float32bits(f))
}
