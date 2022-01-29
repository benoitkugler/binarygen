package binarygen

import (
	"fmt"
	"strconv"
	"strings"
)

// generated code - parser

// return the code needed to check the length of a byte slice
func checkLength(sliceName, returnVars string, minLength int) string {
	return fmt.Sprintf(`if L := len(%s); L < %d {
		return %s, fmt.Errorf("EOF: expected length: %d, got %%d", L)
	}
	`, sliceName, minLength, returnVars, minLength)
}

// return the code needed to check the length of a byte slice
func checkLengthVar(sliceName, returnVars string, lengthExpression string) string {
	return fmt.Sprintf(`if L := len(%s); L < %s {
		return %s, fmt.Errorf("EOF: expected length: %%d, got %%d",%s, L)
	}
	`, sliceName, lengthExpression, returnVars, lengthExpression)
}

// do not perform bounds check
func readBasicType(sliceName string, size int, offset int) string {
	switch size {
	case bytes1:
		return fmt.Sprintf("%s[%d]", sliceName, offset)
	case bytes2:
		return fmt.Sprintf("binary.BigEndian.Uint16(%s[%d:%d])", sliceName, offset, offset+2)
	case bytes4:
		return fmt.Sprintf("binary.BigEndian.Uint32(%s[%d:%d])", sliceName, offset, offset+4)
	case bytes8:
		return fmt.Sprintf("binary.BigEndian.Uint64(%s[%d:%d])", sliceName, offset, offset+8)
	default:
		panic(fmt.Sprintf("size not supported %d", size))
	}
}

func (wc withConstructor) generateParser(dstVar, dataSrcVar string, offset int) string {
	readCode := readBasicType(dataSrcVar, wc.size_, offset)

	if wc.isMethod {
		return fmt.Sprintf("%s.fromUint(%s)\n", dstVar, readCode)
	}
	return fmt.Sprintf("%s = %sFromUint(%s)\n", dstVar, wc.name_, readCode)
}

func (bt basicType) generateParser(dstVar, dataSrcVar string, offset int) string {
	readCode := readBasicType(dataSrcVar, bt.size(), offset)

	constructor := bt.name()
	return fmt.Sprintf("%s = %s(%s)\n", dstVar, constructor, readCode)
}

func (fs fixedSizeStruct) generateParser(dstVar, dataSrcVar string, offset int) string {
	return fmt.Sprintf("%s.mustParse(%s[%d:])\n", dstVar, dataSrcVar, offset)
}

func (fs fixedSizeFields) generateParser(j int, dataSrcVar string, returnVars string, offsetVar string) (code, args string) {
	if len(fs) == 0 {
		return "", ""
	}

	src := fmt.Sprintf("tmp%d", j)
	code = fmt.Sprintf("%s := %s[%s:]\n", src, dataSrcVar, offsetVar)
	code += checkLength(src, returnVars, fs.size())

	code += fs.generateMustParser(src)

	code += fmt.Sprintf("%s += %d\n", offsetVar, fs.size())
	return code, ""
}

func (fs fixedSizeFields) generateParserUnique(typeName string) string {
	if len(fs) == 0 {
		return ""
	}

	returnVars := fmt.Sprintf("%s{}, 0", typeName)
	body := checkLength("data", returnVars, fs.size())

	// call mustParse instead of copying instructions
	body += fmt.Sprintf("out.mustParse(%s)", "data")

	finalCode := fmt.Sprintf(`func parse%s(data []byte) (%s, int, error) {
		var out %s
		%s
		return out, %d, nil
	}
	`, strings.Title(typeName), typeName, typeName, body, fs.size())

	return finalCode
}

func (fs fixedSizeFields) generateMustParser(dataSrcVar string) string {
	code := fmt.Sprintf("_ = %s[%d] // early bound checking\n", dataSrcVar, fs.size()-1)

	pos := 0
	for _, field := range fs {
		code += field.type_.generateParser(fmt.Sprintf("out.%s", field.field.Name()), dataSrcVar, pos)
		pos += field.type_.size()
	}

	return code
}

func (af arrayField) generateParser(j int, dataSrcVar string, returnVars string, offsetVar string) (code, lengthName string) {
	sliceName, lengthName := fmt.Sprintf("tmp%d", j), fmt.Sprintf("arrayLength%d", j)

	returnArgs := ""
	if af.sizeLen == 0 {
		lengthName = strings.ToLower(af.field.Name()) + "Length"
		returnArgs = lengthName + " int"
	}

	code = fmt.Sprintf("%s := %s[%s:]\n", sliceName, dataSrcVar, offsetVar)

	// step 1 : read the array length, if written in the start of the array
	if af.sizeLen != 0 {
		code += checkLength(sliceName, returnVars, af.sizeLen)
		code += fmt.Sprintf("%s := int(%s)\n", lengthName, readBasicType(sliceName, af.sizeLen, 0))
	}

	// step 2 : check the expected length
	code += checkLengthVar(sliceName, returnVars, fmt.Sprintf("%d + %s * %d", af.sizeLen, lengthName, af.element.size()))

	// step 3 : allocate the slice - it is garded by the check above
	code += fmt.Sprintf("out.%s = make([]%s, %s)\n", af.field.Name(), af.element.name(), lengthName)

	// step 4 : loop to parse every elements
	loopBody := af.element.generateParser(fmt.Sprintf("out.%s[i]", af.field.Name()), "chunk", 0)
	code += fmt.Sprintf(`for i := range out.%s {
		chunk := %s[%d + i * %d:]
		%s
	}
	`, af.field.Name(), sliceName, af.sizeLen, af.element.size(), loopBody)

	// step 5 : update the offset
	code += fmt.Sprintf("%s += %d +  %s * %d\n", offsetVar, af.sizeLen, lengthName, af.element.size())

	return code, returnArgs
}

func (vs namedTypeField) generateParser(fieldIndex int, srcVar, returnVars, offsetExpression string) (code string, args string) {
	code = fmt.Sprintf(` var (
		read%d int 
		err%d error
	)
	out.%s, read%d, err%d = parse%s(%s[%s:])
	if err%d != nil {
		return %s,  err%d
	}
	`, fieldIndex, fieldIndex, vs.field.Name(), fieldIndex, fieldIndex,
		strings.Title(vs.name), srcVar, offsetExpression, fieldIndex, returnVars, fieldIndex)
	code += fmt.Sprintf("%s += read%d\n", offsetExpression, fieldIndex)
	return code, ""
}

func generateParserForStruct(chunks []structChunk, name string) string {
	var finalCode string

	body, additionalArgs := "n := 0\n", []string{}
	returnVars := fmt.Sprintf("%s{}, 0", name)
	for j, chunk := range chunks {
		code, args := chunk.generateParser(j, "data", returnVars, "n")
		body += code + "\n"
		if args != "" {
			additionalArgs = append(additionalArgs, args)
		}
	}

	finalCode += fmt.Sprintf(`func parse%s(data []byte, %s) (%s, int, error) {
		var out %s
		%s
		return out, n, nil
	}
	
	`, strings.Title(name), strings.Join(additionalArgs, ","), name, name, body)

	return finalCode
}

// ------------------------------------- V2 -------------------------------------

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
	out := fmt.Sprintf("%s := %s[%s:]", byteSubSliceName, cc.byteSliceName, cc.offsetName)
	cc.byteSliceName = byteSubSliceName
	return out
}

// offset += <expr>
func updateOffset(size int, cc codeContext) string { return updateOffsetExpr(strconv.Itoa(size), cc) }

// offset += <expr>
func updateOffsetExpr(size string, cc codeContext) string {
	return fmt.Sprintf("%s += %s", cc.offsetName, size)
}

// --------------------------- fixed size types ---------------------------

func (wc withConstructor) mustParser(cc codeContext, fieldName string, offset int) string {
	readCode := readBasicType(cc.byteSliceName, wc.size_, offset)

	if wc.isMethod {
		return fmt.Sprintf("%s.fromUint(%s)", cc.variableExpr(fieldName), readCode)
	}
	return fmt.Sprintf("%s = %sFromUint(%s)", cc.variableExpr(fieldName), wc.name_, readCode)
}

func (bt basicType) mustParser(cc codeContext, fieldName string, offset int) string {
	readCode := readBasicType(cc.byteSliceName, bt.size(), offset)

	constructor := bt.name()
	return fmt.Sprintf("%s = %s(%s)", cc.variableExpr(fieldName), constructor, readCode)
}

func (structLayout) mustParser(cc codeContext, fieldName string, offset int) string {
	return fmt.Sprintf("%s.mustParse(%s[%d:])", cc.variableExpr(fieldName), cc.byteSliceName, offset)
}

func (of offset) mustParser(cc codeContext, fieldName string, offset int) string {
	readCode := readBasicType(cc.byteSliceName, of.size_, offset)
	return fmt.Sprintf("offsetTo%s := int(%s)", strings.Title(fieldName), readCode)
}

// returns the reading instructions, without bounds check
// it can be used for example when parsing a slice of such fields
// note that offset are not resolved (only an offset variable is generated)
func (fs fixedSizeList) mustParser(cc codeContext) []string {
	type fixedSizeField interface {
		mustParser(cc codeContext, fieldName string, offset int) string
	}
	var (
		_ fixedSizeField = withConstructor{}
		_ fixedSizeField = basicType{}
		_ fixedSizeField = structLayout{}
		_ fixedSizeField = offset{}
	)

	code := []string{
		fmt.Sprintf("_ = %s[%d] // early bound checking", cc.byteSliceName, fs.size()-1),
	}

	pos := 0
	for _, field := range fs {
		ty := field.type_.(fixedSizeField)
		code = append(code, ty.mustParser(cc, field.name, pos))

		fieldSize, _ := field.type_.staticSize()
		pos += fieldSize
	}

	return code
}

func (fs fixedSizeList) parser(cc codeContext) []string {
	if len(fs) == 0 {
		return nil
	}

	size := fs.size()

	out := []string{
		"{",
		cc.subSlice("subSlice"),
		staticLengthCheck(size, cc),
	}
	out = append(out, fs.mustParser(cc)...)
	out = append(out,
		updateOffset(size, cc),
		"}",
	)

	// TODO: hande offset fields

	return out
}
