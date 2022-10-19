package testpackage

import (
	"encoding/binary"
	"fmt"
)

// Code generated by binarygen from test-package/defs.go. DO NOT EDIT

func ParseComposed2(src []byte) (Composed2, int, error) {
	var item Composed2
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 3 {
			return Composed2{}, 0, fmt.Errorf("EOF: expected length: 3, got %d", L)
		}

		_ = subSlice[2] // early bound checking
		item.a = subSlice[0]
		item.b = subSlice[1]
		item.c = subSlice[2]
		n += 3

	}

	{
		var read int
		var err error
		item.embeded, read, err = parseEmbeded(src[n:])
		if err != nil {
			return Composed2{}, 0, err
		}
		n += read
	}
	return item, n, nil
}
func (item *composed) mustParse(src []byte) {
	_ = src[101] // early bound checking
	item.a.mustParse(src[0:])
	item.b.mustParse(src[51:])
}
func parseEmbeded(src []byte) (embeded, int, error) {
	var item embeded
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return embeded{}, 0, fmt.Errorf("EOF: expected length: 2, got %d", L)
		}

		_ = subSlice[1] // early bound checking
		item.a = subSlice[0]
		item.b = subSlice[1]
		n += 2

	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return embeded{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
		}
		arrayLength := int(binary.BigEndian.Uint16(subSlice[:]))
		if L := len(subSlice); L < 2+arrayLength*2 {
			return embeded{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*2, L)
		}

		item.c = make([]uint16, arrayLength) // allocation guarded by the previous check
		for i := range item.c {
			item.c[i] = binary.BigEndian.Uint16(subSlice[2+i*2:])
		}

		n += 2 + arrayLength*2
	}
	return item, n, nil
}
func (item *lookup) mustParse(src []byte) {
	_ = src[50] // early bound checking
	item.a = int32(binary.BigEndian.Uint32(src[0:]))
	item.b = int32(binary.BigEndian.Uint32(src[4:]))
	item.c = int32(binary.BigEndian.Uint32(src[8:]))
	item.d = binary.BigEndian.Uint32(src[12:])
	item.e = int64(binary.BigEndian.Uint64(src[16:]))
	item.g = src[24]
	item.h = src[25]
	item.t = tag(binary.BigEndian.Uint32(src[26:]))
	item.v.fromUint(binary.BigEndian.Uint16(src[30:]))
	item.w = fl32FromUint(binary.BigEndian.Uint32(src[32:]))
	for i := range item.array1 {
		item.array1[i] = src[36+i]
	}
	for i := range item.array2 {
		item.array2[i] = binary.BigEndian.Uint16(src[41+i*2:])
	}
}
func parseLookup(src []byte) (lookup, int, error) {
	var item lookup
	n := 0
	if L := len(src); L < 51 {
		return lookup{}, 0, fmt.Errorf("EOF: expected length: 51, got %d", L)
	}

	item.mustParse(src)
	n += 51
	return item, n, nil
}
func (item *subtable1) mustParse(src []byte) {
	_ = src[7] // early bound checking
	item.F = binary.BigEndian.Uint64(src[0:])
}
func parseSubtable1(src []byte) (subtable1, int, error) {
	var item subtable1
	n := 0
	if L := len(src); L < 8 {
		return subtable1{}, 0, fmt.Errorf("EOF: expected length: 8, got %d", L)
	}

	item.mustParse(src)
	n += 8
	return item, n, nil
}
func (item *subtable2) mustParse(src []byte) {
	_ = src[0] // early bound checking
	item.F = src[0]
}
func parseSubtable2(src []byte) (subtable2, int, error) {
	var item subtable2
	n := 0
	if L := len(src); L < 1 {
		return subtable2{}, 0, fmt.Errorf("EOF: expected length: 1, got %d", L)
	}

	item.mustParse(src)
	n += 1
	return item, n, nil
}
func parseVarInstance(src []byte, coordsNum int, coords2Num int) (varInstance, int, error) {
	var item varInstance
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < +coordsNum*4 {
			return varInstance{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", +coordsNum*4, L)
		}

		item.Coords = make([]fl1616, coordsNum) // allocation guarded by the previous check
		for i := range item.Coords {
			item.Coords[i] = fl1616FromUint(binary.BigEndian.Uint32(subSlice[+i*4:]))
		}

		n += coordsNum * 4
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < +coords2Num*4 {
			return varInstance{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", +coords2Num*4, L)
		}

		item.Coords2 = make([]fl1616, coords2Num) // allocation guarded by the previous check
		for i := range item.Coords2 {
			item.Coords2[i] = fl1616FromUint(binary.BigEndian.Uint32(subSlice[+i*4:]))
		}

		n += coords2Num * 4
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 4 {
			return varInstance{}, 0, fmt.Errorf("EOF: expected length: 4, got %d", L)
		}

		_ = subSlice[3] // early bound checking
		item.Subfamily = binary.BigEndian.Uint16(subSlice[0:])
		item.PSStringID = binary.BigEndian.Uint16(subSlice[2:])
		n += 4

	}
	return item, n, nil
}
