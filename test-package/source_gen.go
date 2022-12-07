package testpackage

import (
	"encoding/binary"
	"fmt"
)

// Code generated by binarygen from ../../test-package/source_src.go. DO NOT EDIT

func (item *ImplicitITF1) mustParse(src []byte) {
	_ = src[6] // early bound checking
	item.kind = binary.BigEndian.Uint16(src[0:])
	item.data[0] = src[2]
	item.data[1] = src[3]
	item.data[2] = src[4]
	item.data[3] = src[5]
	item.data[4] = src[6]
}

func (item *ImplicitITF2) mustParse(src []byte) {
	_ = src[6] // early bound checking
	item.kind = binary.BigEndian.Uint16(src[0:])
	item.data[0] = src[2]
	item.data[1] = src[3]
	item.data[2] = src[4]
	item.data[3] = src[5]
	item.data[4] = src[6]
}

func (item *ImplicitITF3) mustParse(src []byte) {
	_ = src[41] // early bound checking
	item.kind = binary.BigEndian.Uint16(src[0:])
	item.data[0] = binary.BigEndian.Uint64(src[2:])
	item.data[1] = binary.BigEndian.Uint64(src[10:])
	item.data[2] = binary.BigEndian.Uint64(src[18:])
	item.data[3] = binary.BigEndian.Uint64(src[26:])
	item.data[4] = binary.BigEndian.Uint64(src[34:])
}

func ParseElement(src []byte, parentSrc []byte) (Element, int, error) {
	var item Element
	n := 0
	if L := len(src); L < 10 {
		return item, 0, fmt.Errorf("reading Element: "+"EOF: expected length: 10, got %d", L)
	}
	_ = src[9] // early bound checking
	item.A = int32(binary.BigEndian.Uint32(src[0:]))
	offsetV := int(binary.BigEndian.Uint32(src[4:]))
	arrayLengthVarSizes := int(binary.BigEndian.Uint16(src[8:]))
	n += 10

	{

		if offsetV != 0 { // ignore null offset
			if L := len(parentSrc); L < offsetV {
				return item, 0, fmt.Errorf("reading Element: "+"EOF: expected length: %d, got %d", offsetV, L)
			}

			var (
				err  error
				read int
			)
			item.v, read, err = parseVarSize(parentSrc[offsetV:])
			if err != nil {
				return item, 0, fmt.Errorf("reading Element: %s", err)
			}
			offsetV += read

		}
	}
	{

		if L := len(src); L < 10+arrayLengthVarSizes*4 {
			return item, 0, fmt.Errorf("reading Element: "+"EOF: expected length: %d, got %d", 10+arrayLengthVarSizes*4, L)
		}

		item.VarSizes = make([]varSize, arrayLengthVarSizes) // allocation guarded by the previous check
		for i := range item.VarSizes {
			offset := int(binary.BigEndian.Uint32(src[10+i*4:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(parentSrc); L < offset {
				return item, 0, fmt.Errorf("reading Element: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.VarSizes[i], _, err = parseVarSize(parentSrc[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading Element: %s", err)
			}
		}
		n += arrayLengthVarSizes * 4
	}
	if L := len(src); L < n+4 {
		return item, 0, fmt.Errorf("reading Element: "+"EOF: expected length: n + 4, got %d", L)
	}
	arrayLengthSl := int(binary.BigEndian.Uint32(src[n:]))
	n += 4

	{

		offset := n
		for i := 0; i < arrayLengthSl; i++ {
			elem, read, err := ParseSubElement(src[offset:], parentSrc)
			if err != nil {
				return item, 0, fmt.Errorf("reading Element: %s", err)
			}
			item.sl = append(item.sl, elem)
			offset += read
		}
		n = offset
	}
	return item, n, nil
}

func ParseImplicitITF(src []byte) (ImplicitITF, int, error) {
	var item ImplicitITF

	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading ImplicitITF: "+"EOF: expected length: 2, got %d", L)
	}
	format := uint16(binary.BigEndian.Uint16(src[0:]))
	var (
		read int
		err  error
	)
	switch format {
	case 1:
		item, read, err = ParseImplicitITF1(src[0:])
	case 2:
		item, read, err = ParseImplicitITF2(src[0:])
	case 3:
		item, read, err = ParseImplicitITF3(src[0:])
	default:
		err = fmt.Errorf("unsupported ImplicitITF format %d", format)
	}
	if err != nil {
		return item, 0, fmt.Errorf("reading ImplicitITF: %s", err)
	}

	return item, read, nil
}

func ParseImplicitITF1(src []byte) (ImplicitITF1, int, error) {
	var item ImplicitITF1
	n := 0
	if L := len(src); L < 7 {
		return item, 0, fmt.Errorf("reading ImplicitITF1: "+"EOF: expected length: 7, got %d", L)
	}
	item.mustParse(src)
	n += 7
	return item, n, nil
}

func ParseImplicitITF2(src []byte) (ImplicitITF2, int, error) {
	var item ImplicitITF2
	n := 0
	if L := len(src); L < 7 {
		return item, 0, fmt.Errorf("reading ImplicitITF2: "+"EOF: expected length: 7, got %d", L)
	}
	item.mustParse(src)
	n += 7
	return item, n, nil
}

func ParseImplicitITF3(src []byte) (ImplicitITF3, int, error) {
	var item ImplicitITF3
	n := 0
	if L := len(src); L < 42 {
		return item, 0, fmt.Errorf("reading ImplicitITF3: "+"EOF: expected length: 42, got %d", L)
	}
	item.mustParse(src)
	n += 42
	return item, n, nil
}

func ParsePassArg(src []byte) (PassArg, int, error) {
	var item PassArg
	n := 0
	if L := len(src); L < 8 {
		return item, 0, fmt.Errorf("reading PassArg: "+"EOF: expected length: 8, got %d", L)
	}
	_ = src[7] // early bound checking
	item.kind = binary.BigEndian.Uint16(src[0:])
	item.version = binary.BigEndian.Uint16(src[2:])
	item.count = int32(binary.BigEndian.Uint32(src[4:]))
	n += 8

	{
		var (
			err  error
			read int
		)
		item.customWithArg, read, err = parseWithArgument(src[8:], int(item.count), uint16(item.kind), uint16(item.version))
		if err != nil {
			return item, 0, fmt.Errorf("reading PassArg: %s", err)
		}
		n += read
	}
	return item, n, nil
}

func ParseRootTable(src []byte) (RootTable, int, error) {
	var item RootTable
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading RootTable: "+"EOF: expected length: 4, got %d", L)
	}
	_ = src[3] // early bound checking
	offsetE := int(binary.BigEndian.Uint16(src[0:]))
	arrayLengthEs := int(binary.BigEndian.Uint16(src[2:]))
	n += 4

	{

		if offsetE != 0 { // ignore null offset
			if L := len(src); L < offsetE {
				return item, 0, fmt.Errorf("reading RootTable: "+"EOF: expected length: %d, got %d", offsetE, L)
			}

			var (
				err  error
				read int
			)
			item.E, read, err = ParseElement(src[offsetE:], src)
			if err != nil {
				return item, 0, fmt.Errorf("reading RootTable: %s", err)
			}
			offsetE += read

		}
	}
	{

		offset := 4
		for i := 0; i < arrayLengthEs; i++ {
			elem, read, err := ParseElement(src[offset:], src)
			if err != nil {
				return item, 0, fmt.Errorf("reading RootTable: %s", err)
			}
			item.Es = append(item.Es, elem)
			offset += read
		}
		n = offset
	}
	return item, n, nil
}

func ParseSubElement(src []byte, grandParentSrc []byte) (SubElement, int, error) {
	var item SubElement
	n := 0
	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading SubElement: "+"EOF: expected length: 2, got %d", L)
	}
	offsetV := int(binary.BigEndian.Uint16(src[0:]))
	n += 2

	{

		if offsetV != 0 { // ignore null offset
			if L := len(grandParentSrc); L < offsetV {
				return item, 0, fmt.Errorf("reading SubElement: "+"EOF: expected length: %d, got %d", offsetV, L)
			}

			var (
				err  error
				read int
			)
			item.v, read, err = parseVarSize(grandParentSrc[offsetV:])
			if err != nil {
				return item, 0, fmt.Errorf("reading SubElement: %s", err)
			}
			offsetV += read

		}
	}
	return item, n, nil
}

func ParseVariableThenFixed(src []byte) (VariableThenFixed, int, error) {
	var item VariableThenFixed
	n := 0
	{
		var (
			err  error
			read int
		)
		item.v, read, err = parseVarSize(src[0:])
		if err != nil {
			return item, 0, fmt.Errorf("reading VariableThenFixed: %s", err)
		}
		n += read
	}
	if L := len(src); L < n+11 {
		return item, 0, fmt.Errorf("reading VariableThenFixed: "+"EOF: expected length: n + 11, got %d", L)
	}
	_ = src[n+10] // early bound checking
	item.a = binary.BigEndian.Uint16(src[n:])
	item.b = binary.BigEndian.Uint32(src[n+2:])
	item.c[0] = src[n+6]
	item.c[1] = src[n+7]
	item.c[2] = src[n+8]
	item.c[3] = src[n+9]
	item.c[4] = src[n+10]
	n += 11

	return item, n, nil
}

func ParseWithArray(src []byte) (WithArray, int, error) {
	var item WithArray
	n := 0
	if L := len(src); L < 21 {
		return item, 0, fmt.Errorf("reading WithArray: "+"EOF: expected length: 21, got %d", L)
	}
	item.mustParse(src)
	n += 21
	return item, n, nil
}

func ParseWithChildArgument(src []byte, arrayCount int, kind uint16, version uint16) (WithChildArgument, int, error) {
	var item WithChildArgument
	n := 0
	{
		var (
			err  error
			read int
		)
		item.child, read, err = parseWithArgument(src[0:], arrayCount, kind, version)
		if err != nil {
			return item, 0, fmt.Errorf("reading WithChildArgument: %s", err)
		}
		n += read
	}
	{
		var (
			err  error
			read int
		)
		item.child2, read, err = parseWithArgument(src[n:], arrayCount, kind, version)
		if err != nil {
			return item, 0, fmt.Errorf("reading WithChildArgument: %s", err)
		}
		n += read
	}
	return item, n, nil
}

func ParseWithImplicitITF(src []byte) (WithImplicitITF, int, error) {
	var item WithImplicitITF
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading WithImplicitITF: "+"EOF: expected length: 4, got %d", L)
	}
	item.field1 = binary.BigEndian.Uint32(src[0:])
	n += 4

	{
		var (
			err  error
			read int
		)
		item.itf, read, err = ParseImplicitITF(src[4:])
		if err != nil {
			return item, 0, fmt.Errorf("reading WithImplicitITF: %s", err)
		}
		n += read
	}
	return item, n, nil
}

func ParseWithOffset(src []byte, offsetToSliceCount int) (WithOffset, int, error) {
	var item WithOffset
	n := 0
	if L := len(src); L < 19 {
		return item, 0, fmt.Errorf("reading WithOffset: "+"EOF: expected length: 19, got %d", L)
	}
	_ = src[18] // early bound checking
	item.version = binary.BigEndian.Uint16(src[0:])
	offsetOffsetToSlice := int(binary.BigEndian.Uint32(src[2:]))
	offsetOffsetToStruct := int(binary.BigEndian.Uint32(src[6:]))
	item.a = src[10]
	item.b = src[11]
	item.c = src[12]
	offsetOffsetToUnbounded := int(binary.BigEndian.Uint16(src[13:]))
	offsetOptional := int(binary.BigEndian.Uint32(src[15:]))
	n += 19

	{

		if offsetOffsetToSlice != 0 { // ignore null offset
			if L := len(src); L < offsetOffsetToSlice {
				return item, 0, fmt.Errorf("reading WithOffset: "+"EOF: expected length: %d, got %d", offsetOffsetToSlice, L)
			}

			if L := len(src); L < offsetOffsetToSlice+offsetToSliceCount*8 {
				return item, 0, fmt.Errorf("reading WithOffset: "+"EOF: expected length: %d, got %d", offsetOffsetToSlice+offsetToSliceCount*8, L)
			}

			item.offsetToSlice = make([]uint64, offsetToSliceCount) // allocation guarded by the previous check
			for i := range item.offsetToSlice {
				item.offsetToSlice[i] = binary.BigEndian.Uint64(src[offsetOffsetToSlice+i*8:])
			}
			offsetOffsetToSlice += offsetToSliceCount * 8
		}
	}
	{

		if offsetOffsetToStruct != 0 { // ignore null offset
			if L := len(src); L < offsetOffsetToStruct {
				return item, 0, fmt.Errorf("reading WithOffset: "+"EOF: expected length: %d, got %d", offsetOffsetToStruct, L)
			}

			var (
				err  error
				read int
			)
			item.offsetToStruct, read, err = parseVarSize(src[offsetOffsetToStruct:])
			if err != nil {
				return item, 0, fmt.Errorf("reading WithOffset: %s", err)
			}
			offsetOffsetToStruct += read

		}
	}
	{

		if offsetOffsetToUnbounded != 0 { // ignore null offset
			if L := len(src); L < offsetOffsetToUnbounded {
				return item, 0, fmt.Errorf("reading WithOffset: "+"EOF: expected length: %d, got %d", offsetOffsetToUnbounded, L)
			}

			item.offsetToUnbounded = src[offsetOffsetToUnbounded:]
			offsetOffsetToUnbounded = len(src)
		}
	}
	{

		if offsetOptional != 0 { // ignore null offset
			if L := len(src); L < offsetOptional {
				return item, 0, fmt.Errorf("reading WithOffset: "+"EOF: expected length: %d, got %d", offsetOptional, L)
			}

			var tmpOptional varSize
			var (
				err  error
				read int
			)
			tmpOptional, read, err = parseVarSize(src[offsetOptional:])
			if err != nil {
				return item, 0, fmt.Errorf("reading WithOffset: %s", err)
			}
			offsetOptional += read

			item.optional = &tmpOptional
		}
	}
	return item, n, nil
}

func ParseWithOffsetArray(src []byte) (WithOffsetArray, int, error) {
	var item WithOffsetArray
	n := 0
	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading WithOffsetArray: "+"EOF: expected length: 2, got %d", L)
	}
	arrayLengthArray := int(binary.BigEndian.Uint16(src[0:]))
	n += 2

	{

		if L := len(src); L < 2+arrayLengthArray*4 {
			return item, 0, fmt.Errorf("reading WithOffsetArray: "+"EOF: expected length: %d, got %d", 2+arrayLengthArray*4, L)
		}

		item.array = make([]WithSlices, arrayLengthArray) // allocation guarded by the previous check
		for i := range item.array {
			offset := int(binary.BigEndian.Uint32(src[2+i*4:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading WithOffsetArray: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.array[i], _, err = ParseWithSlices(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading WithOffsetArray: %s", err)
			}
		}
		n += arrayLengthArray * 4
	}
	return item, n, nil
}

func ParseWithOpaque(src []byte) (WithOpaque, int, error) {
	var item WithOpaque
	n := 0
	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading WithOpaque: "+"EOF: expected length: 2, got %d", L)
	}
	item.f = binary.BigEndian.Uint16(src[0:])
	n += 2

	{

		err := item.parseOpaque(src[:])
		if err != nil {
			return item, 0, fmt.Errorf("reading WithOpaque: %s", err)
		}
	}
	{

		read, err := item.parseOpaqueWithLength(src[:])
		if err != nil {
			return item, 0, fmt.Errorf("reading WithOpaque: %s", err)
		}
		n = read
	}
	return item, n, nil
}

func ParseWithRawdata(src []byte, defautCount int, startToCount int) (WithRawdata, int, error) {
	var item WithRawdata
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading WithRawdata: "+"EOF: expected length: 4, got %d", L)
	}
	item.length = binary.BigEndian.Uint32(src[0:])
	n += 4

	{

		L := int(4 + defautCount)
		if len(src) < L {
			return item, 0, fmt.Errorf("reading WithRawdata: "+"EOF: expected length: %d, got %d", L, len(src))
		}
		item.defaut = src[4:L]
		n = L
	}
	{

		L := int(0 + startToCount)
		if len(src) < L {
			return item, 0, fmt.Errorf("reading WithRawdata: "+"EOF: expected length: %d, got %d", L, len(src))
		}
		item.startTo = src[0:L]
		n = L
	}
	{

		item.currentToEnd = src[n:]
		n = len(src)
	}
	{

		item.startToEnd = src[0:]
		n = len(src)
	}
	{

		L := int(item.length)
		if len(src) < L {
			return item, 0, fmt.Errorf("reading WithRawdata: "+"EOF: expected length: %d, got %d", L, len(src))
		}
		item.currentToOffset = src[n:L]
		n = L
	}
	return item, n, nil
}

func ParseWithSlices(src []byte) (WithSlices, int, error) {
	var item WithSlices
	n := 0
	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading WithSlices: "+"EOF: expected length: 2, got %d", L)
	}
	item.length = binary.BigEndian.Uint16(src[0:])
	n += 2

	{
		arrayLength := int(item.length)

		offset := 2
		for i := 0; i < arrayLength; i++ {
			elem, read, err := parseVarSize(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading WithSlices: %s", err)
			}
			item.s1 = append(item.s1, elem)
			offset += read
		}
		n = offset
	}
	return item, n, nil
}

func ParseWithUnion(src []byte) (WithUnion, int, error) {
	var item WithUnion
	n := 0
	if L := len(src); L < 3 {
		return item, 0, fmt.Errorf("reading WithUnion: "+"EOF: expected length: 3, got %d", L)
	}
	_ = src[2] // early bound checking
	item.version = subtableFlagVersion(binary.BigEndian.Uint16(src[0:]))
	item.otherField = src[2]
	n += 3

	{
		var (
			read int
			err  error
		)
		switch item.version {
		case subtableFlagVersion1:
			item.data, read, err = parseSubtableITF1(src[3:])
		case subtableFlagVersion2:
			item.data, read, err = parseSubtableITF2(src[3:])
		default:
			err = fmt.Errorf("unsupported subtableITFVersion %d", item.version)
		}
		if err != nil {
			return item, 0, fmt.Errorf("reading WithUnion: %s", err)
		}
		n += read
	}
	return item, n, nil
}

func (item *WithAlias) mustParse(src []byte) {
	item.f = fl32FromUint(binary.BigEndian.Uint32(src[0:]))
}

func (item *WithArray) mustParse(src []byte) {
	_ = src[20] // early bound checking
	item.a = binary.BigEndian.Uint16(src[0:])
	item.b[0] = binary.BigEndian.Uint32(src[2:])
	item.b[1] = binary.BigEndian.Uint32(src[6:])
	item.b[2] = binary.BigEndian.Uint32(src[10:])
	item.b[3] = binary.BigEndian.Uint32(src[14:])
	item.c[0] = src[18]
	item.c[1] = src[19]
	item.c[2] = src[20]
}

func parseSubtableITF1(src []byte) (subtableITF1, int, error) {
	var item subtableITF1
	n := 0
	if L := len(src); L < 8 {
		return item, 0, fmt.Errorf("reading subtableITF1: "+"EOF: expected length: 8, got %d", L)
	}
	item.mustParse(src)
	n += 8
	return item, n, nil
}

func parseSubtableITF2(src []byte) (subtableITF2, int, error) {
	var item subtableITF2
	n := 0
	if L := len(src); L < 1 {
		return item, 0, fmt.Errorf("reading subtableITF2: "+"EOF: expected length: 1, got %d", L)
	}
	item.mustParse(src)
	n += 1
	return item, n, nil
}

func parseToBeEmbeded(src []byte) (toBeEmbeded, int, error) {
	var item toBeEmbeded
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading toBeEmbeded: "+"EOF: expected length: 4, got %d", L)
	}
	_ = src[3] // early bound checking
	item.a = src[0]
	item.b = src[1]
	arrayLengthC := int(binary.BigEndian.Uint16(src[2:]))
	n += 4

	{

		if L := len(src); L < 4+arrayLengthC*2 {
			return item, 0, fmt.Errorf("reading toBeEmbeded: "+"EOF: expected length: %d, got %d", 4+arrayLengthC*2, L)
		}

		item.c = make([]uint16, arrayLengthC) // allocation guarded by the previous check
		for i := range item.c {
			item.c[i] = binary.BigEndian.Uint16(src[4+i*2:])
		}
		n += arrayLengthC * 2
	}
	return item, n, nil
}

func parseVarSize(src []byte) (varSize, int, error) {
	var item varSize
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading varSize: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.f1 = binary.BigEndian.Uint32(src[0:])
	arrayLengthArray := int(binary.BigEndian.Uint16(src[4:]))
	n += 6

	{

		if L := len(src); L < 6+arrayLengthArray*4 {
			return item, 0, fmt.Errorf("reading varSize: "+"EOF: expected length: %d, got %d", 6+arrayLengthArray*4, L)
		}

		item.array = make([]uint32, arrayLengthArray) // allocation guarded by the previous check
		for i := range item.array {
			item.array[i] = binary.BigEndian.Uint32(src[6+i*4:])
		}
		n += arrayLengthArray * 4
	}
	if L := len(src); L < n+4 {
		return item, 0, fmt.Errorf("reading varSize: "+"EOF: expected length: n + 4, got %d", L)
	}
	arrayLengthStucts := int(binary.BigEndian.Uint32(src[n:]))
	n += 4

	{

		if L := len(src); L < n+arrayLengthStucts*4 {
			return item, 0, fmt.Errorf("reading varSize: "+"EOF: expected length: %d, got %d", n+arrayLengthStucts*4, L)
		}

		item.stucts = make([]WithAlias, arrayLengthStucts) // allocation guarded by the previous check
		for i := range item.stucts {
			item.stucts[i].mustParse(src[n+i*4:])
		}
		n += arrayLengthStucts * 4
	}
	var err error
	n, err = item.parseEnd(src)
	if err != nil {
		return item, 0, fmt.Errorf("reading varSize: %s", err)
	}

	return item, n, nil
}

func parseWithArgument(src []byte, arrayCount int, kind uint16, version uint16) (withArgument, int, error) {
	var item withArgument
	n := 0
	{

		if L := len(src); L < arrayCount*2 {
			return item, 0, fmt.Errorf("reading withArgument: "+"EOF: expected length: %d, got %d", arrayCount*2, L)
		}

		item.array = make([]uint16, arrayCount) // allocation guarded by the previous check
		for i := range item.array {
			item.array[i] = binary.BigEndian.Uint16(src[i*2:])
		}
		n += arrayCount * 2
	}
	return item, n, nil
}

func (item *singleScope) mustParse(src []byte) {
	_ = src[52] // early bound checking
	item.a = int32(binary.BigEndian.Uint32(src[0:]))
	item.b = int32(binary.BigEndian.Uint32(src[4:]))
	item.c = int32(binary.BigEndian.Uint32(src[8:]))
	item.d = binary.BigEndian.Uint32(src[12:])
	item.e = int64(binary.BigEndian.Uint64(src[16:]))
	item.g = src[24]
	item.h = src[25]
	item.t = tag(binary.BigEndian.Uint32(src[26:]))
	item.v = float214(binary.BigEndian.Uint32(src[30:]))
	item.w = fl32FromUint(binary.BigEndian.Uint32(src[34:]))
	item.array1[0] = src[38]
	item.array1[1] = src[39]
	item.array1[2] = src[40]
	item.array1[3] = src[41]
	item.array1[4] = src[42]
	item.array2[0] = binary.BigEndian.Uint16(src[43:])
	item.array2[1] = binary.BigEndian.Uint16(src[45:])
	item.array2[2] = binary.BigEndian.Uint16(src[47:])
	item.array2[3] = binary.BigEndian.Uint16(src[49:])
	item.array2[4] = binary.BigEndian.Uint16(src[51:])
}

func (item *subtableITF1) mustParse(src []byte) {
	item.F = binary.BigEndian.Uint64(src[0:])
}

func (item *subtableITF2) mustParse(src []byte) {
	item.F = src[0]
}

func (item *withFixedSize) mustParse(src []byte) {
	_ = src[52] // early bound checking
	item.a = int32(binary.BigEndian.Uint32(src[0:]))
	item.b = int32(binary.BigEndian.Uint32(src[4:]))
	item.c = int32(binary.BigEndian.Uint32(src[8:]))
	item.d = binary.BigEndian.Uint32(src[12:])
	item.e = int64(binary.BigEndian.Uint64(src[16:]))
	item.g = src[24]
	item.h = src[25]
	item.t = tag(binary.BigEndian.Uint32(src[26:]))
	item.v = float214(binary.BigEndian.Uint32(src[30:]))
	item.w = fl32FromUint(binary.BigEndian.Uint32(src[34:]))
	item.array1[0] = src[38]
	item.array1[1] = src[39]
	item.array1[2] = src[40]
	item.array1[3] = src[41]
	item.array1[4] = src[42]
	item.array2[0] = binary.BigEndian.Uint16(src[43:])
	item.array2[1] = binary.BigEndian.Uint16(src[45:])
	item.array2[2] = binary.BigEndian.Uint16(src[47:])
	item.array2[3] = binary.BigEndian.Uint16(src[49:])
	item.array2[4] = binary.BigEndian.Uint16(src[51:])
}

func (item *withFromExternalFile) mustParse(src []byte) {
	_ = src[105] // early bound checking
	item.a.mustParse(src[0:])
	item.b.mustParse(src[53:])
}
