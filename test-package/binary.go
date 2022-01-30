package testpackage

import (
	"encoding/binary"
	"fmt"
)

// Code generated by bin-parser-gen. DO NOT EDIT

func parseVarInstance(src []byte, coordsLength int, coords2Length int) (VarInstance, int, error) {
	var item VarInstance
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < +coordsLength*4 {
			return VarInstance{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", +coordsLength*4, L)
		}

		item.Coords = make([]fl1616, coordsLength)
		for i := range item.Coords {
			item.Coords[i] = fl1616FromUint(binary.BigEndian.Uint32(subSlice[0+i*4:]))
		}

		n += 0 + coordsLength*4
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < +coords2Length*4 {
			return VarInstance{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", +coords2Length*4, L)
		}

		item.Coords2 = make([]fl1616, coords2Length)
		for i := range item.Coords2 {
			item.Coords2[i] = fl1616FromUint(binary.BigEndian.Uint32(subSlice[0+i*4:]))
		}

		n += 0 + coords2Length*4
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 4 {
			return VarInstance{}, 0, fmt.Errorf("EOF: expected length: 4, got %d", L)
		}

		_ = subSlice[3] // early bound checking
		item.Subfamily = uint16(binary.BigEndian.Uint16(subSlice[0:]))
		item.PSStringID = uint16(binary.BigEndian.Uint16(subSlice[2:]))

		n += 4
	}
	return item, n, nil
}

func parseArrayLike(src []byte) (arrayLike, int, error) {
	var item arrayLike
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
		}

		arrayLength := int(binary.BigEndian.Uint16(subSlice[:]))

		if L := len(subSlice); L < 2+arrayLength*36 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*36, L)
		}

		item.array = make([]lookup, arrayLength)
		for i := range item.array {
			item.array[i].mustParse(subSlice[2+i*36:])
		}

		n += 2 + arrayLength*36
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
		}

		arrayLength := int(binary.BigEndian.Uint16(subSlice[:]))

		if L := len(subSlice); L < 2+arrayLength*72 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*72, L)
		}

		item.array2 = make([]composed, arrayLength)
		for i := range item.array2 {
			item.array2[i].mustParse(subSlice[2+i*72:])
		}

		n += 2 + arrayLength*72
	}
	return item, n, nil
}

func parseComplexeSubtable(src []byte) (complexeSubtable, int, error) {
	var item complexeSubtable
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 6 {
			return complexeSubtable{}, 0, fmt.Errorf("EOF: expected length: 6, got %d", L)
		}

		_ = subSlice[5] // early bound checking
		item.version = uint16(binary.BigEndian.Uint16(subSlice[0:]))
		item.x = int16(binary.BigEndian.Uint16(subSlice[2:]))
		item.y = int16(binary.BigEndian.Uint16(subSlice[4:]))

		n += 6
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return complexeSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
		}

		arrayLength := int(binary.BigEndian.Uint16(subSlice[:]))

		if L := len(subSlice); L < 2+arrayLength*36 {
			return complexeSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*36, L)
		}

		item.lookups = make([]lookup, arrayLength)
		for i := range item.lookups {
			item.lookups[i].mustParse(subSlice[2+i*36:])
		}

		n += 2 + arrayLength*36
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 28 {
			return complexeSubtable{}, 0, fmt.Errorf("EOF: expected length: 28, got %d", L)
		}

		_ = subSlice[27] // early bound checking
		item.u.fromUint(binary.BigEndian.Uint16(subSlice[0:]))
		item.v.fromUint(binary.BigEndian.Uint16(subSlice[2:]))
		item.a = int64(binary.BigEndian.Uint64(subSlice[4:]))
		item.b = int64(binary.BigEndian.Uint64(subSlice[12:]))
		item.c = int64(binary.BigEndian.Uint64(subSlice[20:]))

		n += 28
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 4 {
			return complexeSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 4, L)
		}

		arrayLength := int(binary.BigEndian.Uint32(subSlice[:]))

		if L := len(subSlice); L < 4+arrayLength*4 {
			return complexeSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 4+arrayLength*4, L)
		}

		item.array2 = make([]uint32, arrayLength)
		for i := range item.array2 {
			item.array2[i] = uint32(binary.BigEndian.Uint32(subSlice[4+i*4:]))
		}

		n += 4 + arrayLength*4
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 8 {
			return complexeSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 8, L)
		}

		arrayLength := int(binary.BigEndian.Uint64(subSlice[:]))

		if L := len(subSlice); L < 8+arrayLength*4 {
			return complexeSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 8+arrayLength*4, L)
		}

		item.array3 = make([]fl32, arrayLength)
		for i := range item.array3 {
			item.array3[i] = fl32FromUint(binary.BigEndian.Uint32(subSlice[8+i*4:]))
		}

		n += 8 + arrayLength*4
	}
	return item, n, nil
}

func (item *composed) mustParse(src []byte) {
	_ = src[71] // early bound checking
	item.a.mustParse(src[0:])
	item.b.mustParse(src[36:])
}

func parseComposed(src []byte) (composed, int, error) {
	var item composed
	n := 0
	if L := len(src); L < 72 {
		return composed{}, 0, fmt.Errorf("EOF: expected length: 72, got %d", L)
	}

	item.mustParse(src)
	n += 72
	return item, n, nil
}

func parseComposed2(src []byte) (composed2, int, error) {
	var item composed2
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 3 {
			return composed2{}, 0, fmt.Errorf("EOF: expected length: 3, got %d", L)
		}

		_ = subSlice[2] // early bound checking
		item.a = byte(subSlice[0])
		item.b = byte(subSlice[1])
		item.c = byte(subSlice[2])

		n += 3
	}
	return item, n, nil
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
		item.a = byte(subSlice[0])
		item.b = byte(subSlice[1])

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

		item.c = make([]uint16, arrayLength)
		for i := range item.c {
			item.c[i] = uint16(binary.BigEndian.Uint16(subSlice[2+i*2:]))
		}

		n += 2 + arrayLength*2
	}
	return item, n, nil
}

func (item *lookup) mustParse(src []byte) {
	_ = src[35] // early bound checking
	item.a = int32(binary.BigEndian.Uint32(src[0:]))
	item.b = int32(binary.BigEndian.Uint32(src[4:]))
	item.c = int32(binary.BigEndian.Uint32(src[8:]))
	item.d = uint32(binary.BigEndian.Uint32(src[12:]))
	item.e = int64(binary.BigEndian.Uint64(src[16:]))
	item.g = byte(src[24])
	item.h = byte(src[25])
	item.t = tag(binary.BigEndian.Uint32(src[26:]))
	item.v.fromUint(binary.BigEndian.Uint16(src[30:]))
	item.w = fl32FromUint(binary.BigEndian.Uint32(src[32:]))
}

func parseLookup(src []byte) (lookup, int, error) {
	var item lookup
	n := 0
	if L := len(src); L < 36 {
		return lookup{}, 0, fmt.Errorf("EOF: expected length: 36, got %d", L)
	}

	item.mustParse(src)
	n += 36
	return item, n, nil
}

func parseSimpleSubtable(src []byte) (simpleSubtable, int, error) {
	var item simpleSubtable
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 6 {
			return simpleSubtable{}, 0, fmt.Errorf("EOF: expected length: 6, got %d", L)
		}

		_ = subSlice[5] // early bound checking
		item.version = uint16(binary.BigEndian.Uint16(subSlice[0:]))
		item.x = int16(binary.BigEndian.Uint16(subSlice[2:]))
		item.y = int16(binary.BigEndian.Uint16(subSlice[4:]))

		n += 6
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return simpleSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
		}

		arrayLength := int(binary.BigEndian.Uint16(subSlice[:]))

		if L := len(subSlice); L < 2+arrayLength*36 {
			return simpleSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*36, L)
		}

		item.lookups = make([]lookup, arrayLength)
		for i := range item.lookups {
			item.lookups[i].mustParse(subSlice[2+i*36:])
		}

		n += 2 + arrayLength*36
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return simpleSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
		}

		arrayLength := int(binary.BigEndian.Uint16(subSlice[:]))

		if L := len(subSlice); L < 2+arrayLength*4 {
			return simpleSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*4, L)
		}

		item.array2 = make([]uint32, arrayLength)
		for i := range item.array2 {
			item.array2[i] = uint32(binary.BigEndian.Uint32(subSlice[2+i*4:]))
		}

		n += 2 + arrayLength*4
	}
	return item, n, nil
}

func parseWithOffset(src []byte) (withOffset, int, error) {
	var item withOffset
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 13 {
			return withOffset{}, 0, fmt.Errorf("EOF: expected length: 13, got %d", L)
		}

		_ = subSlice[12] // early bound checking
		item.version = uint16(binary.BigEndian.Uint16(subSlice[0:]))
		offsetToOffsetToSlice := int(binary.BigEndian.Uint32(subSlice[2:]))
		offsetToOffsetToStruct := int(binary.BigEndian.Uint32(subSlice[6:]))
		item.a = byte(subSlice[10])
		item.b = byte(subSlice[11])
		item.c = byte(subSlice[12])

		if L := len(src); L < offsetToOffsetToSlice {
			return withOffset{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", offsetToOffsetToSlice, L)
		}

		if L := len(src); L < offsetToOffsetToStruct {
			return withOffset{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", offsetToOffsetToStruct, L)
		}

		n += 13
	}
	return item, n, nil
}
