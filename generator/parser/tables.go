package parser

import (
	"fmt"
	"go/types"

	an "github.com/benoitkugler/binarygen/analysis"
	gen "github.com/benoitkugler/binarygen/generator"
)

// AllParsers write all the parsing functions in [dst]
func AllParsers(ana an.Analyser, dst *gen.Buffer) {
	for _, table := range ana.Tables {
		for _, decl := range ParserForTable(table) {
			dst.Add(decl)
		}
	}
}

// ParserForTable returns the parsing function for the given table.
// The required methods for fields shall be generated in a separated step.
// TODO: support offset shift
func ParserForTable(ta an.Struct) []gen.Declaration {
	context := &gen.Context{
		Type:      ta.Origin().(*types.Named).Obj().Name(),
		ObjectVar: "item",
		Slice:     "src",                 // defined in args
		Offset:    gen.NewOffset("n", 0), // defined later
	}

	scopes := ta.Scopes()
	if len(scopes) == 0 {
		// empty struct are useful : generate the trivial parser
		return []gen.Declaration{context.ParsingFunc([]string{"[]byte"}, []string{"n := 0"})}
	}

	body, args := []string{"n := 0"}, []string{"src []byte"}
	for _, arg := range requiredArgs(ta) {
		args = append(args, arg.asSignature())
	}

	// important special case when all fields have fixed size (with no offset) :
	// generate a mustParse method
	if _, isFixedSize := ta.IsFixedSize(); isFixedSize {
		fs := scopes[0].(an.StaticSizedFields)
		mustParse, parseBody := mustParserFieldsFunction(fs, *context)
		body = append(body, parseBody)

		return []gen.Declaration{mustParse, context.ParsingFunc(args, body)}
	}

	for _, scope := range scopes {
		body = append(body, parser(scope, context))
	}

	finalCode := context.ParsingFunc(args, body)

	return []gen.Declaration{finalCode}
}

func parser(scope an.Scope, cc *gen.Context) string {
	var code string
	switch scope := scope.(type) {
	case an.StaticSizedFields:
		code = parserForFixedSize(scope, cc)
	case an.SingleField:
		code = parserForSingleField(scope, cc)
	default:
		panic("exhaustive type switch")
	}
	return fmt.Sprintf(`{
		%s
	}`, code)
}

// add the length check
func parserForFixedSize(fs an.StaticSizedFields, cc *gen.Context) string {
	totalSize := fs.Size()
	fields := mustParserFields(fs, cc)
	updateOffset := cc.Offset.UpdateStatement(totalSize)
	return fmt.Sprintf(`%s
		%s
		%s
	`,
		staticLengthCheckAt(totalSize, *cc),
		fields,
		updateOffset)
}

// delegate to the type
func parserForSingleField(field an.SingleField, cc *gen.Context) string {
	return parserForVariableSize(an.Field(field), cc)
}
