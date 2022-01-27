package binarygen

import (
	"fmt"
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

func (fs fixedSizeFields) generateParser(j int, dataSrcVar string, returnVars string, offsetVar string) string {
	if len(fs) == 0 {
		return ""
	}

	src := fmt.Sprintf("tmp%d", j)
	code := fmt.Sprintf("%s := %s[%s:]\n", src, dataSrcVar, offsetVar)
	code += checkLength(src, returnVars, fs.size())

	code += fs.generateMustParser(src)

	code += fmt.Sprintf("%s += %d\n", offsetVar, fs.size())
	return code
}

func (fs fixedSizeFields) generateParserUnique(typeName string) string {
	if len(fs) == 0 {
		return ""
	}

	returnVars := fmt.Sprintf("%s{}", typeName)
	body := checkLength("data", returnVars, fs.size())

	// call mustParse instead of copying instructions
	body += fmt.Sprintf("out.mustParse(%s)", "data")

	finalCode := fmt.Sprintf(`func parse%s(data []byte) (%s, error) {
		var out %s
		%s
		return out, nil
	}
	`, strings.Title(typeName), typeName, typeName, body)

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

func (af arrayField) generateParser(j int, dataSrcVar string, returnVars string, offsetVar string) string {
	sliceName, lengthName := fmt.Sprintf("tmp%d", j), fmt.Sprintf("arrayLength%d", j)

	code := fmt.Sprintf("%s := %s[%s:]\n", sliceName, dataSrcVar, offsetVar)
	// step 1 : read the array length

	code += checkLength(sliceName, returnVars, af.sizeLen)
	code += fmt.Sprintf("%s := int(%s)\n", lengthName, readBasicType(sliceName, af.sizeLen, 0))

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

	return code
}

func generateParserForStruct(chunks []structChunk, name string) string {
	var finalCode string

	body := "n := 0\n"
	for j, chunk := range chunks {
		body += chunk.generateParser(j, "data", fmt.Sprintf("%s{}", name), "n") + "\n"
	}

	finalCode += fmt.Sprintf(`func parse%s(data []byte) (%s, error) {
		var out %s
		%s
		return out, nil
	}
	
	`, strings.Title(name), name, name, body)

	return finalCode
}
