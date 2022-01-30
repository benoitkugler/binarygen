package binarygen

import (
	"fmt"
	"strconv"
	"strings"
)

// generated code - parser

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
	if args.elementSize != 0 && args.lengthName != "" {
		lengthExpr += fmt.Sprintf("+ %s * %d", args.lengthName, args.elementSize)
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
	return fmt.Sprintf("%s = %s(%s)", cc.variableExpr(selector), constructor, readCode)
}

func (structLayout) mustParser(cc codeContext, selector string) string {
	return fmt.Sprintf("%s.mustParse(%s[%s:])", cc.variableExpr(selector), cc.byteSliceName, cc.offsetExpr)
}

func (of offset) mustParser(cc codeContext, selector string) string {
	readCode := readBasicType(cc.byteSliceName, of.size_, cc.offsetExpr)
	return fmt.Sprintf("%s := int(%s)", of.offsetVariableName(selector), readCode)
}

// returns the reading instructions, without bounds check
// it can be used for example when parsing a slice of such fields
// note that offset are not resolved (only an offset variable is generated)
func (fs fixedSizeList) mustParser(cc codeContext) []string {
	code := []string{
		fmt.Sprintf("_ = %s[%d] // early bound checking", cc.byteSliceName, fs.size()-1),
	}

	pos := 0
	for _, field := range fs {
		ty := field.type_

		cc.offsetExpr = strconv.Itoa(pos) // adjust the offset
		code = append(code, ty.mustParser(cc, field.name))

		fieldSize, _ := ty.staticSize()
		pos += fieldSize
	}

	return code
}

// return the mustParse function and the body of the parse function
func (fs fixedSizeList) mustParserFunction(cc codeContext) (mustParse string, parseBody []string) {
	mustParseBody := fs.mustParser(cc)

	mustParse = fmt.Sprintf(`func (%s *%s) mustParse(%s []byte) {
		%s
	}
	`, cc.objectName, cc.typeName, cc.byteSliceName, strings.Join(mustParseBody, "\n"))

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
func (off offset) targetParser(cc codeContext, fieldName string) []string {
	offsetVariable := off.offsetVariableName(fieldName)

	check := affineLengthCheck(affine{offsetExpr: offsetVariable}, cc)
	// TODO:

	return []string{check}
}

func (fs fixedSizeList) offsetTargetsParser(cc codeContext) (out []string) {
	for _, field := range fs {
		off, isOffset := field.type_.(offset)
		if !isOffset {
			continue
		}

		out = append(out, off.targetParser(cc, field.name)...)
	}

	return out
}

func (fs fixedSizeList) parser(cc codeContext) []string {
	if len(fs) == 0 {
		return nil
	}

	size := fs.size()

	// offset are relative to the whole slice, not the subslice
	targets := fs.offsetTargetsParser(cc)

	out := []string{
		"{",
		cc.subSlice("subSlice"),
		staticLengthCheck(size, cc),
	}
	out = append(out, fs.mustParser(cc)...)
	out = append(out, "\n")
	out = append(out, targets...)
	out = append(out,
		updateOffset(size, cc),
		"}",
	)

	// TODO: hande offset fields

	return out
}

func (sl slice) externalLengthVariable(fieldName string) string {
	return strings.ToLower(fieldName) + "Length"
}

func (sl slice) requiredArgs(fieldName string) []string {
	if sl.sizeLen == 0 { // provided as function argument
		return []string{sl.externalLengthVariable(fieldName) + " int"}
	}
	return nil
}

func (sl slice) mustParser(cc codeContext, dstSelector string) string {
	panic("slice are supported as child type")
}

func (sl slice) parser(cc codeContext, fieldName string) []string {
	out := []string{
		"{",
		cc.subSlice("subSlice"),
	}

	lengthName := "arrayLength"
	if sl.sizeLen == 0 {
		lengthName = sl.externalLengthVariable(fieldName)
	}
	elementSize, _ := sl.element.staticSize()

	// step 1 : read the array length, if written in the start of the array
	if sl.sizeLen != 0 {
		out = append(out,
			affineLengthCheck(affine{offsetExpr: strconv.Itoa(sl.sizeLen)}, cc))
		out = append(out,
			fmt.Sprintf("%s := int(%s)", lengthName, readBasicType(cc.byteSliceName, sl.sizeLen, "")))
	}

	// step 2 : check the expected length
	out = append(out,
		affineLengthCheck(affine{offsetExpr: strconv.Itoa(sl.sizeLen), lengthName: lengthName, elementSize: elementSize}, cc))

	// step 3 : allocate the slice - it is garded by the check above
	out = append(out, fmt.Sprintf("%s.%s = make([]%s, %s) // allocation guarded by the previous check", cc.objectName, fieldName, sl.element.name(), lengthName))

	// step 4 : loop to parse every elements
	offset := cc.offsetExpr
	cc.offsetExpr = fmt.Sprintf("%d + i * %d", sl.sizeLen, elementSize)
	loopBody := sl.element.mustParser(cc, fmt.Sprintf("%s[i]", fieldName))
	out = append(out, fmt.Sprintf(`for i := range %s.%s {
		%s
	}
	`, cc.objectName, fieldName, loopBody))

	// step 5 : update the offset and close the scope
	out = append(out,
		fmt.Sprintf("%s += %d +  %s * %d", offset, sl.sizeLen, lengthName, elementSize),
		"}",
	)

	return out
}

func (sf standaloneField) parser(cc codeContext) []string {
	switch ty := sf.type_.(type) {
	case slice:
		return ty.parser(cc, sf.name)
	case structLayout:
		// TODO: call the approriate function, with args
		return nil
	default:
		panic(fmt.Sprintf("not handled yet %T", sf.type_))
	}
}
