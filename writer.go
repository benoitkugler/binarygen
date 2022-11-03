package binarygen

// generated code - writer

// func sliceExpr(sliceName string, offset int) string {
// 	sliceExpr := sliceName
// 	if offset != 0 {
// 		sliceExpr = fmt.Sprintf("%s[%d:]", sliceName, offset)
// 	}
// 	return sliceExpr
// }

// // do not perform bounds check
// func writeBasicType(sliceName, varName string, size int, offset int) string {
// 	slice := sliceExpr(sliceName, offset)
// 	switch size {
// 	case bytes1:
// 		return fmt.Sprintf("%s[%d] = byte(%s)", sliceName, offset, varName)
// 	case bytes2:
// 		return fmt.Sprintf("binary.BigEndian.PutUint16(%s, uint16(%s))", slice, varName)
// 	case bytes4:
// 		return fmt.Sprintf("binary.BigEndian.PutUint32(%s, uint32(%s))", slice, varName)
// 	case bytes8:
// 		return fmt.Sprintf("binary.BigEndian.PutUint64(%s, uint64(%s))", slice, varName)
// 	default:
// 		panic("not supported")
// 	}
// }

// TODO:
// func (wc withConstructor) generateWriter(srcVar, dstSlice string, offset int) string {
// 	var accesVar string
// 	if wc.isMethod {
// 		accesVar = fmt.Sprintf("%s.toUint()", srcVar)
// 	} else {
// 		accesVar = fmt.Sprintf("%sToUint(%s)", wc.name_, srcVar)
// 	}
// 	return writeBasicType(dstSlice, accesVar, wc.size_, offset)
// }

// func (bt basicType) generateWriter(srcVar, dstSlice string, offset int) string {
// 	return writeBasicType(dstSlice, srcVar, bt.size(), offset)
// }

// func (fs fixedSizeStruct) generateWriter(srcVar, dstSlice string, offset int) string {
// 	return fmt.Sprintf("%s.writeTo(%s)", srcVar, sliceExpr(dstSlice, offset))
// }

// // write in place
// func (fs fixedSizeFields) generateWriter(dstSlice, objectName string) string {
// 	code := fmt.Sprintf("_ = %s[%d] // early bound checking\n", dstSlice, fs.size()-1)
// 	pos := 0
// 	for _, field := range fs {
// 		writeCode := field.type_.generateWriter(fmt.Sprintf("%s.%s", objectName, field.field.Name()), dstSlice, pos)
// 		code += writeCode + "\n"
// 		pos += field.type_.size()
// 	}

// 	return code
// }

// // append and return
// func (fs fixedSizeFields) generateAppenderUnique(typeName string) string {
// 	totalSize := fs.size()

// 	finalCode := fmt.Sprintf(`func (item %s) appendTo(data []byte) []byte {
// 		L := len(data)
// 		data = append(data, make([]byte, %d)...)
// 		dst := data[L:]
// 		item.writeTo(dst)
// 		return data
// 	}

// 	`, typeName, totalSize)

// 	return finalCode
// }

// // append and return
// func (fs fixedSizeFields) generateAppender(index int, srcVar, dstSlice string) string {
// 	if len(fs) == 0 {
// 		return ""
// 	}
// 	totalSize := fs.size()

// 	code := fmt.Sprintf(`L%d := len(%s)
// 	%s = append(%s, make([]byte, %d)...)
// 	dst%d := %s[L%d:]
// 	`, index, dstSlice, dstSlice, dstSlice, totalSize, index, dstSlice, index)

// 	code += fs.generateWriter(fmt.Sprintf("dst%d", index), srcVar)

// 	return code
// }

// func (af arrayField) generateAppender(index int, srcVar, dstSlice string) string {
// 	srcSliceName := fmt.Sprintf("%s.%s", srcVar, af.field.Name())
// 	code := fmt.Sprintf(`L%d := len(%s)
// 	%s = append(%s, make([]byte, %d + len(%s) * %d)...)
// 	dst%d := %s[L%d:]
// 	`, index, dstSlice, dstSlice, dstSlice, af.sizeLen, srcSliceName, af.element.size(), index, dstSlice, index)

// 	// write the array length, if required
// 	if af.sizeLen != 0 {
// 		code += writeBasicType(fmt.Sprintf("dst%d", index), fmt.Sprintf("len(%s)", srcSliceName), af.sizeLen, 0) + "\n"
// 	}

// 	// write the elements
// 	code += fmt.Sprintf(`for i, v := range %s {
// 		chunk := %s[%d + i * %d:]
// 		%s
// 	}`, srcSliceName, fmt.Sprintf("dst%d", index), af.sizeLen, af.element.size(), af.element.generateWriter("v", "chunk", 0))

// 	return code
// }

// func (vs namedTypeField) generateAppender(index int, srcVar, dstSlice string) string {
// 	return fmt.Sprintf("%s = %s.%s.appendTo(%s)", dstSlice, srcVar, vs.field.Name(), dstSlice)
// }

// func generateAppenderForStruct(chunks []structChunk, typeName string) string {
// 	var finalCode string

// 	body := ""
// 	for j, chunk := range chunks {
// 		body += chunk.generateAppender(j, "item", "data") + "\n"
// 	}

// 	finalCode += fmt.Sprintf(`func (item %s) appendTo(data []byte) []byte {
// 		%s
// 		return data
// 	}

// 	`, typeName, body)

// 	return finalCode
// }
