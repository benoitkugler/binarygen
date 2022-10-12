package binarygen

import (
	"fmt"
	"strconv"
	"strings"
)

// generated code - parser

func (of offset) parser(cc codeContext, dstSelector string) string {
	return parserForFixedSize(dstSelector, of, cc)
}

func (wc withConstructor) parser(cc codeContext, dstSelector string) string {
	return parserForFixedSize(dstSelector, wc, cc)
}

func (bt basicType) parser(cc codeContext, dstSelector string) string {
	return parserForFixedSize(dstSelector, bt, cc)
}

func (ar array) parser(cc codeContext, dstSelector string) string {
	return parserForFixedSize(dstSelector, ar, cc)
}

// do not perform bounds check
func readBasicType(sliceName string, size int, offset string) string {
	switch size {
	case bytes1:
		if offset == "" {
			offset = "0"
		}
		return fmt.Sprintf("%s[%s]", sliceName, offset)
	case bytes2:
		return fmt.Sprintf("binary.BigEndian.Uint16(%s[%s:])", sliceName, offset)
	case bytes4:
		return fmt.Sprintf("binary.BigEndian.Uint32(%s[%s:])", sliceName, offset)
	case bytes8:
		return fmt.Sprintf("binary.BigEndian.Uint64(%s[%s:])", sliceName, offset)
	default:
		panic(fmt.Sprintf("size not supported %d", size))
	}
}

// func (vs namedTypeField) generateParser(fieldIndex int, srcVar, returnVars, offsetExpression string) (code string, args string) {
// 	code = fmt.Sprintf(` var (
// 		read%d int
// 		err%d error
// 	)
// 	out.%s, read%d, err%d = parse%s(%s[%s:])
// 	if err%d != nil {
// 		return %s,  err%d
// 	}
// 	`, fieldIndex, fieldIndex, vs.field.Name(), fieldIndex, fieldIndex,
// 		strings.Title(vs.name), srcVar, offsetExpression, fieldIndex, returnVars, fieldIndex)
// 	code += fmt.Sprintf("%s += read%d\n", offsetExpression, fieldIndex)
// 	return code, ""
// }

// instruction to check the length of <sliceName>
// the `codeContext` is used to generate the proper error return statement,
// and to identify the input slice
// there are 3 cases :
//	- static length
//	- length dependent on the runtime length of an array
//	- length depends on external condition (optional fields)

// for fixed size types
func staticLengthCheck(size int, cc codeContext) string {
	errReturn := cc.returnError(fmt.Sprintf(`fmt.Errorf("EOF: expected length: %d, got %%d", L)`, size))
	return fmt.Sprintf(`if L := len(%s); L < %d {
		%s
	}
	`, cc.byteSliceName, size, errReturn)
}

type affine struct {
	offsetExpr, lengthName string
	elementSize            int
}

// check for <offset> + <elementSize> * <lengthName>
func affineLengthCheck(args affine, cc codeContext) string {
	var lengthExpr string
	if args.offsetExpr != "" && args.offsetExpr != "0" {
		lengthExpr += args.offsetExpr
	}
	if args.lengthName != "" {
		if args.elementSize == 1 {
			lengthExpr += fmt.Sprintf("+ %s", args.lengthName)
		} else if args.elementSize != 0 {
			lengthExpr += fmt.Sprintf("+ %s * %d", args.lengthName, args.elementSize)
		}
	}
	errReturn := cc.returnError(fmt.Sprintf(`fmt.Errorf("EOF: expected length: %%d, got %%d", %s, L)`, lengthExpr))
	return fmt.Sprintf(`if L := len(%s); L < %s {
		%s
	}
	`, cc.byteSliceName, lengthExpr, errReturn)
}

type conditionalField struct {
	name string
	size int
}

func (cf conditionalField) variableName() string { return "has" + strings.Title(cf.name) }

type conditionalLength struct {
	baseLength string // size without the optional fields
	conditions []conditionalField
}

func conditionalLengthCheck(args conditionalLength, cc codeContext) string {
	out := fmt.Sprintf(`{
		expectedLength := %s
	`, args.baseLength)
	for _, cd := range args.conditions {
		out += fmt.Sprintf(`if %s {
			expectedLength += %d
		}
		`, cd.variableName(), cd.size)
	}
	errReturn := cc.returnError(fmt.Sprintf(`fmt.Errorf("EOF: expected length: %%d, got %%d", expectedLength, L)`))
	out += fmt.Sprintf(`if L := len(%s); L < expectedLength {
		%s
		}
	}
	`, cc.byteSliceName, errReturn)
	return out
}

// slices the current input slice at the current offset
// and assigns it to `byteSubSliceName`
// Also updates the codeContext.byteSliceName
func (cc *codeContext) subSlice(byteSubSliceName string) string {
	out := fmt.Sprintf("%s := %s[%s:]", byteSubSliceName, cc.byteSliceName, cc.offsetExpr)
	cc.byteSliceName = byteSubSliceName
	return out
}

// offset += <expr>
func updateOffset(size int, cc codeContext) string { return updateOffsetExpr(strconv.Itoa(size), cc) }

// offset += <expr>
func updateOffsetExpr(size string, cc codeContext) string {
	return fmt.Sprintf("%s += %s", cc.offsetExpr, size)
}

// --------------------------- fixed size types ---------------------------

func (wc withConstructor) mustParser(cc codeContext, selector string) string {
	readCode := readBasicType(cc.byteSliceName, wc.size_, cc.offsetExpr)

	if wc.isMethod {
		return fmt.Sprintf("%s.fromUint(%s)", cc.variableExpr(selector), readCode)
	}
	return fmt.Sprintf("%s = %sFromUint(%s)", cc.variableExpr(selector), wc.name_, readCode)
}

func (bt basicType) mustParser(cc codeContext, selector string) string {
	readCode := readBasicType(cc.byteSliceName, bt.size(), cc.offsetExpr)

	constructor := bt.name()
	constructorExpr := fmt.Sprintf("%s(%s)", constructor, readCode)
	switch bt.name_ {
	case "uint8", "byte", "uint16", "uint32", "uint64": // simplify by removing the unnecessary conversion
		constructorExpr = readCode
	}
	return fmt.Sprintf("%s = %s", cc.variableExpr(selector), constructorExpr)
}

func (structLayout) mustParser(cc codeContext, selector string) string {
	return fmt.Sprintf("%s.mustParse(%s[%s:])", cc.variableExpr(selector), cc.byteSliceName, cc.offsetExpr)
}

func (of offset) mustParser(cc codeContext, selector string) string {
	readCode := readBasicType(cc.byteSliceName, of.size_, cc.offsetExpr)
	return fmt.Sprintf("%s := int(%s)", of.offsetVariableName(selector), readCode)
}

func (ar array) mustParser(cc codeContext, selector string) string {
	cc.setArrayLikeOffsetExpr(ar.element.binarySize, cc.offsetExpr)
	return fmt.Sprintf(`for i := range %s {
		%s
	}`, cc.variableExpr(selector), ar.element.mustParser(cc, selector+"[i]"))
}

// returns the reading instructions, without bounds check
// it can be used for example when parsing a slice of such fields
// note that offset are not resolved (only an offset variable is generated)
func (fs fixedSizeList) mustParser(cc codeContext) string {
	code := []string{
		fmt.Sprintf("_ = %s[%d] // early bound checking", cc.byteSliceName, fs.size()-1),
	}

	pos := 0
	for _, field := range fs {
		ty := field.type_.(fixedSizeType)

		cc.offsetExpr = strconv.Itoa(pos) // adjust the offset
		code = append(code, ty.mustParser(cc, field.name))

		fieldSize, _ := ty.staticSize()
		pos += fieldSize
	}

	return strings.Join(code, "\n")
}

// return the mustParse function and the body of the parse function
func (fs fixedSizeList) mustParserFunction(cc codeContext) (mustParse string, parseBody []string) {
	mustParseBody := fs.mustParser(cc)

	mustParse = fmt.Sprintf(`func (%s *%s) mustParse(%s []byte) {
		%s
	}
	`, cc.objectName, cc.typeName, cc.byteSliceName, mustParseBody)

	// for parse: check length and call mustParse
	check := staticLengthCheck(fs.size(), cc)
	mustParseCall := fmt.Sprintf("%s.mustParse(%s)", cc.objectName, cc.byteSliceName)
	offset := updateOffset(fs.size(), cc)

	parseBody = []string{
		check,
		mustParseCall,
		offset,
	}

	return mustParse, parseBody
}

// handle the parsing of the data pointed to by the offset
func (off offset) targetParser(cc codeContext, fieldName string) string {
	offsetVariable := off.offsetVariableName(fieldName)

	check := affineLengthCheck(affine{offsetExpr: offsetVariable}, cc)

	cc.offsetExpr = offsetVariable
	parse := off.target.parser(cc, fieldName)

	return check + "\n" + parse
}

func (fs fixedSizeList) offsetTargetsParser(cc codeContext) (out string) {
	var chunks []string
	for _, field := range fs {
		off, isOffset := field.type_.(offset)
		if !isOffset {
			continue
		}

		chunks = append(chunks, off.targetParser(cc, field.name))
	}

	return strings.Join(chunks, "\n")
}

func (fs fixedSizeList) parser(cc codeContext) string {
	if len(fs) == 0 {
		return ""
	}

	size := fs.size()

	// offset are relative to the whole slice, not the subslice
	targets := fs.offsetTargetsParser(cc)

	return fmt.Sprintf(`{
		%s
		%s
		%s
		%s
		%s
	}`, cc.subSlice("subSlice"),
		staticLengthCheck(size, cc),
		fs.mustParser(cc),
		updateOffset(size, cc),
		targets)
}

func (sl slice) externalLengthVariable(fieldName string) string {
	return strings.ToLower(fieldName) + "Length"
}

func (sl slice) parser(cc codeContext, fieldName string) string {
	// special case for unbounded data
	if sl.lengthLocation == "_toEnd" {
		return fmt.Sprintf(`%s = %s[%s:]
		%s = len(%s)
		`, cc.variableExpr(fieldName), cc.byteSliceName, cc.offsetExpr,
			cc.offsetExpr, cc.byteSliceName,
		)
	}

	out := []string{
		"{",
		cc.subSlice("subSlice"),
	}

	lengthName := "arrayLength"
	if sl.lengthLocation == "" {
		lengthName = sl.externalLengthVariable(fieldName)
	}
	elementSize, _ := sl.element.staticSize()

	// step 1 : read the array length, if written in the start of the array
	sizeOffsetExpr := ""
	if strings.HasPrefix(sl.lengthLocation, "_first") {
		size := sizeFromTag(strings.TrimPrefix(sl.lengthLocation, "_first"))
		sizeOffsetExpr = strconv.Itoa(size)
		out = append(out,
			affineLengthCheck(affine{offsetExpr: strconv.Itoa(size)}, cc))
		out = append(out,
			fmt.Sprintf("%s := int(%s)", lengthName, readBasicType(cc.byteSliceName, size, "")))
	} else if sl.lengthLocation != "" {
		// length is provided by a field
		out = append(out, fmt.Sprintf("%s := int(%s)", lengthName, cc.variableExpr(sl.lengthLocation)))
	}

	// step 2 : check the expected length
	out = append(out,
		affineLengthCheck(affine{offsetExpr: sizeOffsetExpr, lengthName: lengthName, elementSize: elementSize}, cc))

	// step 3 : allocate the slice - it is garded by the check above
	out = append(out, fmt.Sprintf("%s = make([]%s, %s) // allocation guarded by the previous check", cc.variableExpr(fieldName), sl.element.name(), lengthName))

	// step 4 : loop to parse every elements
	offset := cc.offsetExpr
	cc.setArrayLikeOffsetExpr(elementSize, sizeOffsetExpr)
	loopBody := sl.element.mustParser(cc, fmt.Sprintf("%s[i]", fieldName))
	out = append(out, fmt.Sprintf(`for i := range %s {
		%s
	}
	`, cc.variableExpr(fieldName), loopBody))

	// step 5 : update the offset and close the scope
	cc.offsetExpr = offset
	increment := fmt.Sprintf("%s + %s * %d", sizeOffsetExpr, lengthName, elementSize)
	if sizeOffsetExpr == "" {
		increment = fmt.Sprintf("%s * %d", lengthName, elementSize)
	}
	out = append(out,
		updateOffsetExpr(increment, cc),
		"}",
	)

	return strings.Join(out, "\n")
}

// add the bound checks
func parserForFixedSize(fieldName string, ty fixedSizeType, cc codeContext) string {
	ls := fixedSizeList{{name: fieldName, type_: ty}}
	return ls.parser(cc)
}

func (st structLayout) parser(cc codeContext, dstSelector string) string {
	var args []string
	for _, arg := range st.requiredArgs() {
		args = append(args, arg.variableName)
	}
	return fmt.Sprintf(`
		{
			var read int
			var err error
			%s, read, err = parse%s(%s[%s:], %s)
 			if err != nil {
				%s
			}
			%s
		}`, cc.variableExpr(dstSelector), strings.Title(st.name_), cc.byteSliceName, cc.offsetExpr, strings.Join(args, ", "),
		cc.returnError("err"),
		updateOffsetExpr("read", cc),
	)
}

func (st structField) parser(cc codeContext) string {
	return st.type_.parser(cc, st.name)
}

func (u union) parser(cc codeContext, dstSelector string) string {
	var cases []string
	for i, flag := range u.flags {
		cases = append(cases, fmt.Sprintf(`case %s :
		%s, read, err = parse%s(%s[%s:], %s)`,
			flag.Name(), cc.variableExpr(dstSelector), strings.Title(u.members[i].name()), cc.byteSliceName,
			cc.offsetExpr, "", // TODO: if needed handle args
		))
	}
	kindVariable := cc.variableExpr(u.flagFieldName)
	return fmt.Sprintf(`{
		var read int
		var err error
		switch %s {
		%s
		default:
			err = fmt.Errorf("unsupported %sVersion %%d", %s)
		}
		if err != nil {
			%s
		}
		%s
	}`, kindVariable,
		strings.Join(cases, "\n"),
		u.name(),
		kindVariable,
		cc.returnError("err"),
		updateOffsetExpr("read", cc))
}
