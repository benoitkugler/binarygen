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
		cc.Offset.SetStatement("n + read"),
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
	start := cc.Offset.Value()
	if field.Layout.SubsliceStart == an.AtStart { // do not use the current offset as start
		start = ""
	}

	codes := []string{countCode}

	// special case for bytes data
	if sl.IsRawData() {
		codes = append(codes, parserForSliceBytes(sl, cc, start, countExpr, field.Name))
	}

	// // else, check for fixed size elements
	// if _, isFixedSize := sl.Elem.IsFixedSize(); isFixedSize {
	// 	return parserForSliceFixedSizeElement(cc, selector)
	// }

	// return parserForSliceVariableSizeElement(cc, selector)
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
		statements = append(statements, staticLengthCheckAt(size, *cc))
		// 2 - read the value
		statements = append(statements, fmt.Sprintf("%s := int(%s)", countVar, readBasicTypeAt(*cc, size)))
		// 3 - increment the offset value
		cc.Offset.Increment(size)
	case an.ComputedField:
		countVar = "arrayLength"
		statements = append(statements, fmt.Sprintf("%s := int(%s)", countVar, cc.Selector(sl.CountExpr)))
	case an.ToEnd:
		// count is ignored in this case
	}

	return countVar, strings.Join(statements, "\n")
}

func parserForSliceBytes(sl an.Slice, cc *gen.Context, start gen.Expression, count gen.Expression, fieldName string) string {
	target := cc.Selector(fieldName)

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

// func parserForSliceUnbounded(sl an.Slice, cc gen.Context, selector string, layout an.Layout) {
// 	if sl.lengthLocation == "__toEnd" {
// 		return fmt.Sprintf(`%s = %s[%s:]
// 			%s = len(%s)
// 			`, cc.variableExpr(fieldName), cc.byteSliceName, cc.offsetExpr,
// 			cc.offsetExpr, cc.byteSliceName,
// 		)
// 	} else if sl.lengthLocation == "__startToEnd" {
// 		return fmt.Sprintf(`%s = %s
// 			%s = len(%s)
// 			`, cc.variableExpr(fieldName), cc.byteSliceName,
// 			cc.offsetExpr, cc.byteSliceName,
// 		)
// 	} else {
// 		var fieldLength, sliceExpr string
// 		if strings.HasPrefix(sl.lengthLocation, "__to") {
// 			fieldLength = strings.TrimPrefix(sl.lengthLocation, "__to_")
// 			sliceExpr = fmt.Sprintf("%s:%s", cc.offsetExpr, cc.variableExpr(fieldLength))
// 		} else {
// 			fieldLength = strings.TrimPrefix(sl.lengthLocation, "__startTo_")
// 			sliceExpr = fmt.Sprintf(":%s", cc.variableExpr(fieldLength))
// 		}
// 		errorStatement := fmt.Sprintf(`fmt.Errorf("EOF: expected length: %%d, got %%d", L, len(%s))`, cc.byteSliceName)
// 		return fmt.Sprintf(`
// 			L := int(%s)
// 			if len(%s) < L {
// 				%s
// 			}
// 			%s = %s[%s]
// 			%s = L
// 			`, cc.variableExpr(fieldLength),
// 			cc.byteSliceName,
// 			cc.returnError(errorStatement),
// 			cc.variableExpr(fieldName), cc.byteSliceName, sliceExpr,
// 			cc.offsetExpr,
// 		)
// 	}
// }
