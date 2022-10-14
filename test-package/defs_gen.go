package testpackage

import (
	"encoding/binary"
	"fmt"
)

// Code generated by binarygen from test-package/defs.go. DO NOT EDIT

func parseArrayLike(src []byte) (arrayLike, int, error) {
	var item arrayLike
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: 2, got %d", L)
		}

		_ = subSlice[1] // early bound checking
		item.size = binary.BigEndian.Uint16(subSlice[0:])
		n += 2

	}
	{
		subSlice := src[n:]
		arrayLength := int(item.size)
		if L := len(subSlice); L < +arrayLength*2 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", +arrayLength*2, L)
		}

		item.datas = make([]uint16, arrayLength) // allocation guarded by the previous check
		for i := range item.datas {
			item.datas[i] = binary.BigEndian.Uint16(subSlice[+i*2:])
		}

		n += arrayLength * 2
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
		}

		arrayLength := int(binary.BigEndian.Uint16(subSlice[:]))
		if L := len(subSlice); L < 2+arrayLength*51 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*51, L)
		}

		item.array = make([]lookup, arrayLength) // allocation guarded by the previous check
		for i := range item.array {
			item.array[i].mustParse(subSlice[2+i*51:])
		}

		n += 2 + arrayLength*51
	}
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 2 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
		}

		arrayLength := int(binary.BigEndian.Uint16(subSlice[:]))
		if L := len(subSlice); L < 2+arrayLength*102 {
			return arrayLike{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*102, L)
		}

		item.array2 = make([]composed, arrayLength) // allocation guarded by the previous check
		for i := range item.array2 {
			item.array2[i].mustParse(subSlice[2+i*102:])
		}

		n += 2 + arrayLength*102
	}
	item.data = src[n:]
	n = len(src)

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
		item.version = binary.BigEndian.Uint16(subSlice[0:])
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
		if L := len(subSlice); L < 2+arrayLength*51 {
			return complexeSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*51, L)
		}

		item.lookups = make([]lookup, arrayLength) // allocation guarded by the previous check
		for i := range item.lookups {
			item.lookups[i].mustParse(subSlice[2+i*51:])
		}

		n += 2 + arrayLength*51
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

		item.array2 = make([]uint32, arrayLength) // allocation guarded by the previous check
		for i := range item.array2 {
			item.array2[i] = binary.BigEndian.Uint32(subSlice[4+i*4:])
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

		item.array3 = make([]fl32, arrayLength) // allocation guarded by the previous check
		for i := range item.array3 {
			item.array3[i] = fl32FromUint(binary.BigEndian.Uint32(subSlice[8+i*4:]))
		}

		n += 8 + arrayLength*4
	}
	item.rawData = src
	n = len(src)

	return item, n, nil
}
func (item *composed) mustParse(src []byte) {
	_ = src[101] // early bound checking
	item.a.mustParse(src[0:])
	item.b.mustParse(src[51:])
}
func parseComposed(src []byte) (composed, int, error) {
	var item composed
	n := 0
	if L := len(src); L < 102 {
		return composed{}, 0, fmt.Errorf("EOF: expected length: 102, got %d", L)
	}

	item.mustParse(src)
	n += 102
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
			return composed2{}, 0, err
		}
		n += read
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
func parseSimpleSubtable(src []byte) (simpleSubtable, int, error) {
	var item simpleSubtable
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 6 {
			return simpleSubtable{}, 0, fmt.Errorf("EOF: expected length: 6, got %d", L)
		}

		_ = subSlice[5] // early bound checking
		item.version = binary.BigEndian.Uint16(subSlice[0:])
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
		if L := len(subSlice); L < 2+arrayLength*51 {
			return simpleSubtable{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*51, L)
		}

		item.lookups = make([]lookup, arrayLength) // allocation guarded by the previous check
		for i := range item.lookups {
			item.lookups[i].mustParse(subSlice[2+i*51:])
		}

		n += 2 + arrayLength*51
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

		item.array2 = make([]uint32, arrayLength) // allocation guarded by the previous check
		for i := range item.array2 {
			item.array2[i] = binary.BigEndian.Uint32(subSlice[2+i*4:])
		}

		n += 2 + arrayLength*4
	}
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
func parseVarInstance(src []byte, coordsLength int, coords2Length int) (varInstance, int, error) {
	var item varInstance
	n := 0
	{
		subSlice := src[n:]

		if L := len(subSlice); L < +coordsLength*4 {
			return varInstance{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", +coordsLength*4, L)
		}

		item.Coords = make([]fl1616, coordsLength) // allocation guarded by the previous check
		for i := range item.Coords {
			item.Coords[i] = fl1616FromUint(binary.BigEndian.Uint32(subSlice[+i*4:]))
		}

		n += coordsLength * 4
	}
	{
		subSlice := src[n:]

		if L := len(subSlice); L < +coords2Length*4 {
			return varInstance{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", +coords2Length*4, L)
		}

		item.Coords2 = make([]fl1616, coords2Length) // allocation guarded by the previous check
		for i := range item.Coords2 {
			item.Coords2[i] = fl1616FromUint(binary.BigEndian.Uint32(subSlice[+i*4:]))
		}

		n += coords2Length * 4
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
func parseVarInstanceContainer(src []byte, coordsLength int, coords2Length int) (varInstanceContainer, int, error) {
	var item varInstanceContainer
	n := 0

	{
		var read int
		var err error
		item.inst, read, err = parseVarInstance(src[n:], coordsLength, coords2Length)
		if err != nil {
			return varInstanceContainer{}, 0, err
		}
		n += read
	}
	return item, n, nil
}
func parseWithOffset(src []byte) (withOffset, int, error) {
	var item withOffset
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 15 {
			return withOffset{}, 0, fmt.Errorf("EOF: expected length: 15, got %d", L)
		}

		_ = subSlice[14] // early bound checking
		item.version = binary.BigEndian.Uint16(subSlice[0:])
		offsetToOffsetToSlice := int(binary.BigEndian.Uint32(subSlice[2:]))
		offsetToOffsetToStruct := int(binary.BigEndian.Uint32(subSlice[6:]))
		item.a = subSlice[10]
		item.b = subSlice[11]
		item.c = subSlice[12]
		offsetToOffsetToUnbounded := int(binary.BigEndian.Uint16(subSlice[13:]))
		n += 15
		if L := len(src); L < offsetToOffsetToSlice {
			return withOffset{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", offsetToOffsetToSlice, L)
		}

		{
			subSlice := src[offsetToOffsetToSlice:]
			if L := len(subSlice); L < 2 {
				return withOffset{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2, L)
			}

			arrayLength := int(binary.BigEndian.Uint16(subSlice[:]))
			if L := len(subSlice); L < 2+arrayLength*8 {
				return withOffset{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", 2+arrayLength*8, L)
			}

			item.offsetToSlice = make([]uint64, arrayLength) // allocation guarded by the previous check
			for i := range item.offsetToSlice {
				item.offsetToSlice[i] = binary.BigEndian.Uint64(subSlice[2+i*8:])
			}

			offsetToOffsetToSlice += 2 + arrayLength*8
		}
		if L := len(src); L < offsetToOffsetToStruct {
			return withOffset{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", offsetToOffsetToStruct, L)
		}

		{
			var read int
			var err error
			item.offsetToStruct, read, err = parseLookup(src[offsetToOffsetToStruct:])
			if err != nil {
				return withOffset{}, 0, err
			}
			offsetToOffsetToStruct += read
		}
		if L := len(src); L < offsetToOffsetToUnbounded {
			return withOffset{}, 0, fmt.Errorf("EOF: expected length: %d, got %d", offsetToOffsetToUnbounded, L)
		}

		item.offsetToUnbounded = src[offsetToOffsetToUnbounded:]
		offsetToOffsetToUnbounded = len(src)

	}
	return item, n, nil
}
func parseWithUnion(src []byte) (withUnion, int, error) {
	var item withUnion
	n := 0
	{
		subSlice := src[n:]
		if L := len(subSlice); L < 3 {
			return withUnion{}, 0, fmt.Errorf("EOF: expected length: 3, got %d", L)
		}

		_ = subSlice[2] // early bound checking
		item.version = subtableVersion(binary.BigEndian.Uint16(subSlice[0:]))
		item.otherField = subSlice[2]
		n += 3

	}
	{
		var read int
		var err error
		switch item.version {
		case subtableVersion1:
			item.data, read, err = parseSubtable1(src[n:])
		case subtableVersion2:
			item.data, read, err = parseSubtable2(src[n:])
		default:
			err = fmt.Errorf("unsupported subtableVersion %d", item.version)
		}
		if err != nil {
			return withUnion{}, 0, err
		}
		n += read
	}
	return item, n, nil
}
