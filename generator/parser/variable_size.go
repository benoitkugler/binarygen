package parser

import (
	"fmt"
	"strings"

	an "github.com/benoitkugler/binarygen/analysis"
	gen "github.com/benoitkugler/binarygen/generator"
)

func parserForVariableSize(field an.Field, parent an.Struct, cc *gen.Context) string {
	switch ty := field.Type.(type) {
	case an.Slice:
		return parserForSlice(field, cc)
	case an.Opaque:
		return parserForOpaque(field, parent, cc)
	case an.Offset:
		return parserForOffset(field, parent, cc)
	case an.Union:
		// TODO: for implicit interfaces, generate a separate parsing function
		return parserForUnion(field, cc)
	case an.Struct:
		args := resolveArguments(cc.ObjectVar, field, requiredArgs(ty))
		return fmt.Sprintf(`var (
			err error
			read int
		)
		%s, read, err = %s(%s[%s:], %s)
		if err != nil {
			%s 
		}
		%s
		`, cc.Selector(field.Name), gen.ParseFunctionName(gen.Name(field.Type)), cc.Slice, cc.Offset.Value(), args,
			cc.ErrReturn(gen.ErrVariable("err")),
			cc.Offset.UpdateStatementDynamic("read"))
	}
	return ""
}

// delegate the parsing to a user written method of the form
// <structName>.customParse<fieldName>
func parserForOpaque(field an.Field, parent an.Struct, cc *gen.Context) string {
	start := cc.Offset.Value()
	updateOffset := cc.Offset.UpdateStatementDynamic("read")
	if field.Layout.SubsliceStart == an.AtStart { // do not use the current offset as start
		start = ""
		updateOffset = cc.Offset.SetStatement("read")
	}
	args := resolveArguments(cc.ObjectVar, field, requiredArgs(parent))
	return fmt.Sprintf(`
	read, err := %s.customParse%s(%s[%s:], %s)
	if err != nil {
		%s
	}
	%s
	`, cc.ObjectVar, strings.Title(field.Name), cc.Slice, start, args,
		cc.ErrReturn(gen.ErrVariable("err")),
		updateOffset,
	)
}

// ------------------------- slices -------------------------

// we distinguish the following cases for a Slice :
//   - elements have a static sized : we can check the length early
//     and use mustParse on each element
//   - elements have a variable length : we have to check the length at each iteration
//   - as an optimization, we special case raw bytes (see [Slice.IsRawData])
//   - slice of offsets are handled is in dedicated function
//   - opaque types, whose interpretation is defered are represented by an [an.Opaque] type,
//     and handled in a separate function
func parserForSlice(field an.Field, cc *gen.Context) string {
	sl := field.Type.(an.Slice)
	// no matter the kind of element, resolve the count...
	countExpr, countCode := codeForSliceCount(sl, field.Name, cc)

	// ... and adjust the start offset
	if field.Layout.SubsliceStart == an.AtStart { // do not use the current offset as start
		cc.Offset = gen.NewOffset(cc.Offset.Name, 0)
	}

	codes := []string{countCode}

	if sl.IsRawData() { // special case for bytes data
		codes = append(codes, parserForSliceBytes(sl, cc, countExpr, field.Name))
	} else if offset, isOffset := sl.Elem.(an.Offset); isOffset { // special case for slice of offsets
		codes = append(codes, parserForSliceOfOffsets(offset, cc, countExpr, field))
	} else if _, isFixedSize := sl.Elem.IsFixedSize(); isFixedSize { // else, check for fixed size elements
		codes = append(codes, parserForSliceFixedSizeElement(sl, cc, countExpr, field.Name))
	} else {
		codes = append(codes, parserForSliceVariableSizeElement(sl, cc, countExpr, field))
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
	case an.ToEnd, an.ToComputedField:
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

	lengthDefinition := fmt.Sprintf("L := int(%s + %s)", start, count)
	if sl.Count == an.ToComputedField { // the length is not relative to the start
		lengthDefinition = fmt.Sprintf("L := int(%s)", cc.Selector(sl.CountExpr))
	}

	errorStatement := fmt.Sprintf(`"EOF: expected length: %%d, got %%d", L, len(%s)`, cc.Slice)
	return fmt.Sprintf(` 
			%s
			if len(%s) < L {
				%s
			}
			%s = %s[%s:L]
			%s
			`,
		lengthDefinition,
		cc.Slice,
		cc.ErrReturn(gen.ErrFormated(errorStatement)),
		target, cc.Slice, start,
		cc.Offset.SetStatement("L"),
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
	// temporarily changing the offset
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
func parserForSliceVariableSizeElement(sl an.Slice, cc *gen.Context, count gen.Expression, field an.Field) string {
	// if start is a constant, we have to use an additional variable

	args := ""
	if st, isStruct := sl.Elem.(an.Struct); isStruct {
		args = resolveArguments(cc.ObjectVar, field, requiredArgs(st))
	}
	// loop and update the offset
	return fmt.Sprintf(`
		offset := %s
		for i := 0; i < %s; i++ {
		elem, read, err := %s(%s[offset:], %s)
		if err != nil {
			%s
		}
		%s = append(%s, elem)
		offset += read
		}
		%s`,
		cc.Offset.Value(),
		count,
		gen.ParseFunctionName(gen.Name(sl.Elem)), cc.Slice, args,
		cc.ErrReturn(gen.ErrVariable("err")),
		cc.Selector(field.Name), cc.Selector(field.Name),
		cc.Offset.SetStatement("offset"),
	)
}

// ------------------------ Offsets ------------------------

func parserForOffset(fi an.Field, parent an.Struct, cc *gen.Context) string {
	of := fi.Type.(an.Offset)
	var statements []string
	// Step 1 - check the length for the offset integer value
	statements = append(statements, staticLengthCheckAt(*cc, of.Size))

	// Step 2 - read the offset value
	statements = append(statements, fmt.Sprintf("offset := int(%s)", readBasicTypeAt(*cc, of.Size)))
	cc.Offset.Increment(of.Size)
	// generally speaking with have to update the main offset as well
	statements = append(statements, cc.Offset.UpdateStatement(of.Size))
	// Step 3 - check the length for the pointed value
	statements = append(statements, lengthCheck(*cc, "offset"))

	// Step 4 - finally delegate to the target parser
	savedOffset := cc.Offset
	cc.Offset = gen.NewOffsetDynamic("offset")
	statements = append(statements, parserForVariableSize(an.Field{
		Type:                      of.Target,
		Layout:                    fi.Layout,
		Name:                      fi.Name,
		ArgumentsProvidedByFields: fi.ArgumentsProvidedByFields,
		UnionTag:                  fi.UnionTag,
	}, parent, cc))
	cc.Offset = savedOffset
	return strings.Join(statements, "\n")
}

// slice of offsets: this is somewhat a mix of [parserForSliceVariableSizeElement] and [parserForOffset].
// The generated code looks like :
//
//	if len(src) < arrayCount * offsetSize {
//		return err
//	}
//	elems := make([]ElemType, arrayCount)
//	for i := range elems {
//		offset := readUint()
//		if len(src) < offset {
//			return err
//		}
//		elems[i] = parseElemType(src[offset:])
//	}
func parserForSliceOfOffsets(of an.Offset, cc *gen.Context, count gen.Expression, field an.Field) string {
	target := cc.Selector(field.Name)
	out := []string{""}

	// step 1 : check the expected length
	elementSize := of.Size
	out = append(out, affineLengthCheckAt(*cc, count, elementSize))

	// step 2 : allocate the slice of offsets target - it is garded by the check above
	out = append(out, fmt.Sprintf("%s = make([]%s, %s) // allocation guarded by the previous check",
		target, gen.Name(of.Target), count))

	// step 3 : loop to parse every elements,
	// temporarily changing the offset
	startOffset := cc.Offset
	cc.Offset = gen.NewOffsetDynamic(cc.Offset.WithAffine("i", elementSize))

	args := ""
	if st, isStruct := of.Target.(an.Struct); isStruct {
		args = resolveArguments(cc.ObjectVar, field, requiredArgs(st))
	}

	// Loop body :
	// Step 1 - read the offset value
	// Step 2 - check the length for the pointed value
	// Step 3 - finally delegate to the target parser
	loopBody := fmt.Sprintf(`offset := int(%s)
	// ignore null offsets 
	if offset == 0 {
		continue
	}
	
	%s
	var err error
	%s[i], _, err = %s(%s[offset:], %s)
	if err != nil {
		%s
	}
	`,
		readBasicTypeAt(*cc, elementSize),
		lengthCheck(*cc, "offset"),
		target, gen.ParseFunctionName(gen.Name(of.Target)), cc.Slice, args,
		cc.ErrReturn(gen.ErrVariable("err")),
	)

	out = append(out, fmt.Sprintf(`for i := range %s {
		%s
	}`, target, loopBody))

	// step 4 : update the offset
	cc.Offset = startOffset
	out = append(out,
		cc.Offset.UpdateStatementDynamic(fmt.Sprintf("%s * %d", count, elementSize)))

	return strings.Join(out, "\n")
}

// -- unions --

func parserForUnion(fl an.Field, cc *gen.Context) string {
	u := fl.Type.(an.Union)

	start := cc.Offset.Value()
	if fl.Layout.SubsliceStart == an.AtStart { // do not use the current offset as start
		start = ""
	}

	flags := u.UnionTag.TagsCode()
	var cases []string
	for i, flag := range flags {
		member := u.Members[i]
		args := resolveArguments(cc.ObjectVar, fl, requiredArgs(member))
		cases = append(cases, fmt.Sprintf(`case %s :
		%s, read, err = %s(%s[%s:], %s)`,
			flag, cc.Selector(fl.Name), gen.ParseFunctionName(gen.Name(member)), cc.Slice,
			start, args,
		))
	}

	var code string
	switch scheme := u.UnionTag.(type) {
	case an.UnionTagExplicit:
		kindVariable := cc.Selector(scheme.FlagField)
		code = fmt.Sprintf(`var (
				read int
				err error
			)
			switch %s {
			%s
			default:
				err = fmt.Errorf("unsupported %sVersion %%d", %s)
			}
			if err != nil {
				%s
			}
			`, kindVariable,
			strings.Join(cases, "\n"),
			gen.Name(u),
			kindVariable,
			cc.ErrReturn(gen.ErrVariable("err")),
		)
	case an.UnionTagImplicit:
		// steps :
		// 	1 : check the length for the format tag
		//	2 : read the format tag
		//	3 : defer to the corresponding member parsing function
		tagSize, _ := scheme.Tag.IsFixedSize()
		code = fmt.Sprintf(`
			%s
			format := %s(%s)
			var (
				read int
				err error
			)
			switch format {
			%s
			default:
				err = fmt.Errorf("unsupported %s format %%d", format)
			}
			if err != nil {
				%s
			}
			`,
			staticLengthCheckAt(*cc, tagSize),
			gen.Name(scheme.Tag), readBasicTypeAt(*cc, tagSize),
			strings.Join(cases, "\n"),
			gen.Name(u),
			cc.ErrReturn(gen.ErrVariable("err")),
		)
	default:
		panic("exhaustive type switch")
	}

	updateOffset := cc.Offset.UpdateStatementDynamic("read")
	if fl.Layout.SubsliceStart == an.AtStart { // do not use the current offset as start
		updateOffset = cc.Offset.SetStatement("read")
	}
	return code + updateOffset
}
