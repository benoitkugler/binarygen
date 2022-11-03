package parser

import (
	"fmt"
	"strings"

	an "github.com/benoitkugler/binarygen/analysis"
	gen "github.com/benoitkugler/binarygen/generator"
)

// mustParser is only valid for type [ty] with a fixed sized,
// it will panic otherwise
// note thaht th
func mustParser(ty an.Type, cc gen.Context, selector string) string {
	switch ty := ty.(type) {
	case an.Basic:
		return mustParserBasic(ty, cc, selector)
	case an.DerivedFromBasic:
		return mustParserDerived(ty, cc, selector)
	case an.Struct:
		return mustParserStruct(ty, cc, selector)
	case an.Array:
		return mustParserArray(ty, cc, selector)
	default:
		// other types are never fixed sized
		panic(fmt.Sprintf("invalid type %T in mustParser", ty))
	}
}

func mustParserBasic(bt an.Basic, cc gen.Context, selector string) string {
	size, _ := bt.IsFixedSize()
	readCode := readBasicTypeAt(cc, size)

	name := gen.Name(bt)

	switch name {
	case "uint8", "byte", "uint16", "uint32", "uint64": // simplify by removing the unnecessary conversion
		return fmt.Sprintf("%s = %s", cc.Selector(selector), readCode)
	default:
		return fmt.Sprintf("%s = %s(%s)", cc.Selector(selector), name, readCode)
	}
}

func mustParserDerived(de an.DerivedFromBasic, cc gen.Context, selector string) string {
	readCode := readBasicTypeAt(cc, de.Size)
	return fmt.Sprintf("%s = %sFromUint(%s)", cc.Selector(selector), de.Name, readCode)
}

// only valid for fixed size structs, call the `mustParse` method
func mustParserStruct(st an.Struct, cc gen.Context, selector string) string {
	return fmt.Sprintf("%s.mustParse(%s[%s:])", cc.Selector(selector), cc.Slice, cc.Offset.Value())
}

func mustParserArray(ar an.Array, cc gen.Context, selector string) string {
	elemSize, ok := ar.Elem.IsFixedSize()
	if !ok {
		panic("mustParserArray only support fixed size elements")
	}

	statements := make([]string, ar.Len)
	for i := range statements {
		// adjust the selector
		elemSelector := fmt.Sprintf("%s[%d]", selector, i)
		// generate the code
		statements[i] = mustParser(ar.Elem, cc, elemSelector)
		// update the context offset
		cc.Offset.Increment(elemSize)
	}
	return strings.Join(statements, "\n")
}

// extension to a scope

// returns the reading instructions, without bounds check
// it can be used for example when parsing a slice of such fields
func mustParserFields(fs an.StaticSizedFields, cc *gen.Context) string {
	var code []string

	// optimize following slice access
	if len(fs) >= 2 {
		code = append(code, fmt.Sprintf("_ = %s[%s] // early bound checking", cc.Slice, cc.Offset.With(fs.Size()-1)))
	}

	for _, field := range fs {
		code = append(code, mustParser(field.Type, *cc, field.Name))

		fieldSize, _ := field.Type.IsFixedSize()
		// adjust the offset
		cc.Offset.Increment(fieldSize)
	}

	return strings.Join(code, "\n")
}

// return the mustParse method and the body of the parse function
func mustParserFieldsFunction(fs an.StaticSizedFields, cc gen.Context) (mustParse gen.Declaration, parseBody string) {
	contextCopy := cc
	mustParseBody := mustParserFields(fs, &contextCopy) // pass a copy of context not influence the next calls

	mustParse.ID = string(cc.Type) + ".mustParse"
	mustParse.Content = fmt.Sprintf(`func (%s *%s) mustParse(%s []byte) {
		%s
	}
	`, cc.ObjectVar, cc.Type, cc.Slice, mustParseBody)

	// for the parsing function: check length, call mustParse, and update the offset
	check := staticLengthCheckAt(cc, fs.Size())
	mustParseCall := fmt.Sprintf("%s.mustParse(%s)", cc.ObjectVar, cc.Slice)
	updateOffset := cc.Offset.UpdateStatement(fs.Size())

	parseBody = strings.Join([]string{
		check,
		mustParseCall,
		string(updateOffset),
	}, "\n")

	return mustParse, parseBody
}
