// Package binarygen implements a binary parser and writer generator.
package binarygen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func Generate(path string) error {
	an, err := importSource(path)
	if err != nil {
		return err
	}

	an.performAnalysis()

	code := an.generateCode()

	outfile := filepath.Join(filepath.Dir(path), "binary.go")
	content := []byte(fmt.Sprintf(`
	package %s

	// Code generated by bin-parser-gen. DO NOT EDIT

	%s
	`, an.pkgName, code))

	err = os.WriteFile(outfile, content, os.ModePerm)
	if err != nil {
		return err
	}

	err = exec.Command("goimports", "-w", outfile).Run()

	return err
}

func (an *analyser) generateCode() string {
	var keys []string
	for k := range an.structLayouts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	code := ""
	for _, k := range keys {
		st := an.structLayouts[k]
		code += st.generateParser() + "\n"
	}

	return code
}

// ----------------------------- V2 -----------------------------
type codeContext struct {
	// <variableName> = parse<typeName>(<byteSliceName>)
	// <byteSliceName> = appendTo(<variableName>, <byteSliceName>)
	typeName      string // the name of the type being generated
	objectName    string // the go struct being parsed or dumped
	byteSliceName string // the name of the []byte being read or written
	offsetExpr    string // the name of the variable holding the current offset or a number
}

func (cc codeContext) returnError(errVariable string) string {
	return fmt.Sprintf("return %s{}, 0, %s", cc.typeName, errVariable)
}

// return <object>.<field>
func (cc codeContext) variableExpr(field string) string {
	return fmt.Sprintf("%s.%s", cc.objectName, field)
}

func (cc codeContext) parseFunction(args, body []string) string {
	return fmt.Sprintf(`func parse%s(%s) (%s, int, error) {
		var %s %s
		%s
		return %s, n, nil
	}
	`, strings.Title(cc.typeName), strings.Join(args, ","), cc.typeName, cc.objectName,
		cc.typeName, strings.Join(body, "\n"), cc.objectName)
}

// one or many field whose parsing (or writting)
// is grouped to reduce length checks and allocations
type group interface {
	// return an (optional) slice of arguments to add to the parse and appendTo functions
	requiredArgs() []string

	// returns the code blocks
	// no trailing new line is required
	parser(cc codeContext) []string

	// returns the code blocks
	// no trailing new line is required
	appender(cc codeContext) []string
}

// group definition

// func requiredArgs(fields []structField) []string {

// }

func (fixedSizeList) requiredArgs() []string {
	// TODO:
	return nil
}

func (sf standaloneField) requiredArgs() []string {
	switch ty := sf.type_.(type) {
	case slice:
		return ty.requiredArgs(sf.name)
	case structLayout:
		// TODO: call the approriate function, with args
		return nil
	default:
		panic(fmt.Sprintf("not handled yet %T", sf.type_))
	}
}

func (fixedSizeList) appender(cc codeContext) []string { return nil }

// non fixed fields, like slice or structs containing slices or offsets
type standaloneField structField

func (standaloneField) appender(cc codeContext) []string { return nil }

func (st structLayout) generateParser() string {
	groups := st.groups()
	if len(groups) == 0 {
		return ""
	}

	context := codeContext{
		typeName:      st.name_,
		objectName:    "item",
		byteSliceName: "src",
		offsetExpr:    "n",
	}

	body, args := []string{"n := 0"}, []string{fmt.Sprintf("%s []byte", context.byteSliceName)}

	// important special case : all fields have fixed size
	// with no offset
	_, isStaticSize := st.staticSize()
	if fs, isFixedSize := groups[0].(fixedSizeList); len(groups) == 1 && isFixedSize && isStaticSize {
		mustParse, parseBody := fs.mustParserFunction(context)
		args = append(args, fs.requiredArgs()...)
		body = append(body, parseBody...)

		finalCode := mustParse + "\n\n" + context.parseFunction(args, body)

		return finalCode
	}

	for _, group := range groups {
		body = append(body, group.parser(context)...)
		args = append(args, group.requiredArgs()...)
	}

	finalCode := context.parseFunction(args, body)

	return finalCode
}
