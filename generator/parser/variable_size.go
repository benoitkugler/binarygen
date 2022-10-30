package parser

import (
	"fmt"
	"strings"

	an "github.com/benoitkugler/binarygen/analysis"
	gen "github.com/benoitkugler/binarygen/generator"
)

func parserForVariableSize(field an.Field, cc *gen.Context) string {
	switch field.Type.(type) {
	case an.Slice:
		return parserForSlice(field, cc)
	case an.Opaque:
		return parserForOpaque(field, cc)
	}
	return ""
}

// delegate the parsing to a user written method
func parserForOpaque(field an.Field, cc *gen.Context) string {
	start := cc.Offset.Value()
	if field.Layout.SubsliceStart == an.AtStart { // do not use the current offset as start
		start = ""
	}
	return fmt.Sprintf(`
	read, err := %s.customParse(%s[%s:])
	if err != nil {
		%s
	}
	%s
	`, cc.Selector(field.Name), cc.Slice, start,
		cc.ErrReturn("err"),
		cc.Offset.UpdateStatementDynamic("read"),
	)
}

// we distinguish the following cases for a Slice :
//   - elements have a static sized : we can check the length early
//     and use mustParse on each element
//   - elements have a variable length : we have to check the length at each iteration
//   - as an optimization, we special case raw bytes (see [Slice.IsRawData])
//   - opaque types, whose interpretation is defered are represented by an [an.Opaque] type,
//     and handled in a separate function
func parserForSlice(field an.Field, cc *gen.Context) string {
	sl := field.Type.(an.Slice)
	// no matter the kind of element, resolve the count...
	countExpr, countCode := codeForSliceCount(sl, field.Name, cc)

	// ... and the start offset
	if field.Layout.SubsliceStart == an.AtStart { // do not use the current offset as start
		cc.Offset = gen.NewOffset(cc.Offset.Name, 0)
	}

	codes := []string{countCode}

	if sl.IsRawData() { // special case for bytes data
		codes = append(codes, parserForSliceBytes(sl, cc, countExpr, field.Name))
	} else if _, isFixedSize := sl.Elem.IsFixedSize(); isFixedSize { // else, check for fixed size elements
		codes = append(codes, parserForSliceFixedSizeElement(sl, cc, countExpr, field.Name))
	} else {
		codes = append(codes, parserForSliceVariableSizeElement(sl, cc, countExpr, field.Name))
	}

	return strings.Join(codes, "\n")
}

func codeForSliceCount(sl an.Slice, fieldName string, cc *gen.Context) (countVar gen.Expression, code string) {
	var statements []string
	switch sl.Count {
	case an.NoLength: // the length is provided as an external variable
		countVar = externalCountVariable(fieldName)
	case an.FirstUint16, an.FirstUint32: // the length is at the start of the array
		countVar = "arrayLength"
		// add the code to read it
		size := an.Uint16
		if sl.Count == an.FirstUint32 {
			size = an.Uint32
		}
		// 1 - check the length
		statements = append(statements, staticLengthCheckAt(*cc, size))
		// 2 - read the value
		statements = append(statements, fmt.Sprintf("%s := int(%s)", countVar, readBasicTypeAt(*cc, size)))
		// 3 - increment the offset value
		cc.Offset.Increment(size)
		statements = append(statements, cc.Offset.UpdateStatement(size))
	case an.ComputedField:
		countVar = "arrayLength"
		statements = append(statements, fmt.Sprintf("%s := int(%s)", countVar, cc.Selector(sl.CountExpr)))
	case an.ToEnd:
		// count is ignored in this case
	}

	return countVar, strings.Join(statements, "\n")
}

func parserForSliceBytes(sl an.Slice, cc *gen.Context, count gen.Expression, fieldName string) string {
	target := cc.Selector(fieldName)
	start := cc.Offset.Value()
	// special case for ToEnd : do not use an intermediate variable
	if sl.Count == an.ToEnd {
		readStatement := fmt.Sprintf("%s = %s[%s:]", target, cc.Slice, start)
		offsetStatemtent := cc.Offset.SetStatement(fmt.Sprintf("len(%s)", cc.Slice))
		return readStatement + "\n" + offsetStatemtent
	}

	errorStatement := fmt.Sprintf(`fmt.Errorf("EOF: expected length: %%d, got %%d", L, len(%s))`, cc.Slice)
	offsetStatement := cc.Offset.SetStatement("L")
	return fmt.Sprintf(` 
			L := int(%s + %s)
			if len(%s) < L {
				%s
			}
			%s = %s[%s:L]
			%s
			`,
		start, count,
		cc.Slice,
		cc.ErrReturn(errorStatement),
		target, cc.Slice, start,
		offsetStatement,
	)
}

// The field is a slice of structs (or basic type), whose size is known at compile time.
// We can thus check for the whole slice length, and use mustParseXXX functions.
// The generated code will look like
//
//	if len(data) < n + arrayLength * size {
//		return err
//	}
//	out = make([]MorxChain, arrayLength)
//	for i := range out {
//		out[i] = mustParseMorxChain(data[])
//	}
//	n += arrayLength * size
func parserForSliceFixedSizeElement(sl an.Slice, cc *gen.Context, count gen.Expression, fieldName string) string {
	target := cc.Selector(fieldName)
	out := []string{""}

	// step 1 : check the expected length
	elementSize, _ := sl.Elem.IsFixedSize()
	out = append(out, affineLengthCheckAt(*cc, count, elementSize))

	// step 2 : allocate the slice - it is garded by the check above
	out = append(out, fmt.Sprintf("%s = make([]%s, %s) // allocation guarded by the previous check",
		target, gen.Name(sl.Elem), count))

	// step 3 : loop to parse every elements,
	// temporarily chaning the offset
	startOffset := cc.Offset
	cc.Offset = gen.NewOffsetDynamic(cc.Offset.WithAffine("i", elementSize))
	loopBody := mustParser(sl.Elem, *cc, fmt.Sprintf("%s[i]", fieldName))
	out = append(out, fmt.Sprintf(`for i := range %s {
		%s
	}`, target, loopBody))

	// step 4 : update the offset
	cc.Offset = startOffset
	out = append(out,
		cc.Offset.UpdateStatementDynamic(fmt.Sprintf("%s * %d", count, elementSize)))

	return strings.Join(out, "\n")
}

// The field is a slice of structs, whose size is only known at run time
// The generated code will look like
//
//	offset := 2
//	for i := 0; i < arrayLength; i++ {
//		chain, read, err := parseMorxChain(data[offset:])
//		if err != nil {
//			return nil, err
//		}
//		out = append(out, chain)
//		offset += read
//	}
//	n = offset
func parserForSliceVariableSizeElement(sl an.Slice, cc *gen.Context, count gen.Expression, fieldName string) string {
	// if start is a constant, we have to use an additional variable

	// loop and update the offset
	return fmt.Sprintf(`
		offset := %s
		for i := 0; i < %s; i++ {
		elem, read, err := parse%s(%s[offset:])
		if err != nil {
			%s
		}
		%s = append(%s, elem)
		offset += read
		}
		%s`,
		cc.Offset.Value(),
		count,
		strings.Title(gen.Name(sl.Elem)), cc.Slice,
		cc.ErrReturn("err"),
		cc.Selector(fieldName), cc.Selector(fieldName),
		cc.Offset.SetStatement("offset"),
	)
}