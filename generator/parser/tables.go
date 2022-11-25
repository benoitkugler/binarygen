package parser

import (
	"fmt"
	"go/types"

	an "github.com/benoitkugler/binarygen/analysis"
	gen "github.com/benoitkugler/binarygen/generator"
)

// ParsersForFile write the parsing functions required by [ana.Tables] in [dst]
func ParsersForFile(ana an.Analyser, dst *gen.Buffer) {
	for _, table := range ana.Tables {
		for _, decl := range parserForTable(table) {
			dst.Add(decl)
		}
	}

	for _, standaloneUnion := range ana.StandaloneUnions {
		dst.Add(parserForStanaloneUnion(standaloneUnion))
	}
}

// parserForTable returns the parsing function for the given table.
// The required methods for fields shall be generated in a separated step.
func parserForTable(ta an.Struct) []gen.Declaration {
	context := &gen.Context{
		Type:      ta.Origin().(*types.Named).Obj().Name(),
		ObjectVar: "item",
		Slice:     "src",                                 // defined in args
		Offset:    gen.NewOffset("n", ta.StartingOffset), // defined later
	}

	scopes := ta.Scopes()
	if len(scopes) == 0 {
		// empty struct are useful : generate the trivial parser
		return []gen.Declaration{context.ParsingFunc([]string{"[]byte"}, []string{"n := 0"})}
	}

	body, args := []string{fmt.Sprintf("n := %s", context.Offset.Value())}, []string{"src []byte"}
	for _, arg := range requiredArgs(ta, "") {
		args = append(args, arg.asSignature())
	}
	comment := ""
	if ta.StartingOffset != 0 {
		comment = fmt.Sprintf("the actual data starts at %s[%d:]", context.Slice, ta.StartingOffset)
	}

	// important special case when all fields have fixed size (with no offset) :
	// generate a mustParse method
	if _, isFixedSize := ta.IsFixedSize(); isFixedSize {
		fs := scopes[0].(an.StaticSizedFields)
		mustParse, parseBody := mustParserFieldsFunction(fs, *context)
		body = append(body, parseBody)

		return []gen.Declaration{mustParse, context.ParsingFuncComment(args, body, comment)}
	}

	for _, scope := range scopes {
		body = append(body, parser(scope, ta, context))
	}

	finalCode := context.ParsingFuncComment(args, body, comment)

	return []gen.Declaration{finalCode}
}

// parserForStanaloneUnion returns the parsing function for the given union.
func parserForStanaloneUnion(un an.Union) gen.Declaration {
	context := &gen.Context{
		Type:      un.Origin().(*types.Named).Obj().Name(),
		ObjectVar: "item",
		Slice:     "src",                    // defined in args
		Offset:    gen.NewOffset("read", 0), // defined later
	}

	body, args := []string{}, []string{"src []byte"}

	cases := unionCases(un, an.AtCurrent, context, nil, context.ObjectVar)
	code := standaloneUnionBody(un, context, cases)
	body = append(body, code)

	finalCode := context.ParsingFunc(args, body)

	return finalCode
}

func parser(scope an.Scope, parent an.Struct, cc *gen.Context) string {
	var code string
	switch scope := scope.(type) {
	case an.StaticSizedFields:
		code = parserForFixedSize(scope, cc)
	case an.SingleField:
		code = parserForSingleField(scope, parent, cc)
	default:
		panic("exhaustive type switch")
	}
	return code
}

// add the length check
func parserForFixedSize(fs an.StaticSizedFields, cc *gen.Context) string {
	totalSize := fs.Size()
	return fmt.Sprintf(`%s
		%s
		%s
		`,
		staticLengthCheckAt(*cc, totalSize),
		mustParserFields(fs, cc),
		cc.Offset.UpdateStatement(totalSize),
	)
}

// delegate to the type
func parserForSingleField(field an.SingleField, parent an.Struct, cc *gen.Context) string {
	code := parserForVariableSize(an.Field(field), parent, cc)
	return fmt.Sprintf(`{
		%s}`, code)
}
