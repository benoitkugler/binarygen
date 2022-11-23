package parser

import (
	"fmt"
	"sort"
	"strings"

	an "github.com/benoitkugler/binarygen/analysis"
	gen "github.com/benoitkugler/binarygen/generator"
)

// read a basic value at the current offset,
// do not perform bounds check
func readBasicTypeAt(cc gen.Context, size an.BinarySize) string {
	sliceName, offset := cc.Slice, cc.Offset.Value()
	switch size {
	case an.Byte:
		return fmt.Sprintf("%s[%s]", sliceName, offset)
	case an.Uint16:
		return fmt.Sprintf("binary.BigEndian.Uint16(%s[%s:])", sliceName, offset)
	case an.Uint32:
		return fmt.Sprintf("binary.BigEndian.Uint32(%s[%s:])", sliceName, offset)
	case an.Uint64:
		return fmt.Sprintf("binary.BigEndian.Uint64(%s[%s:])", sliceName, offset)
	default:
		panic(fmt.Sprintf("size not supported %d", size))
	}
}

// instruction to check the length of <sliceName>
// the `Context` is used to generate the proper error return statement,
// and to identify the input slice
// there are 3 cases :
//	- static length
//	- length dependent on the runtime length of an array
//	- length depends on external condition (optional fields)

// check for <length> (from the start of the slice)
func lengthCheck(cc gen.Context, length gen.Expression) string {
	errReturn := cc.ErrReturn(gen.ErrFormated(fmt.Sprintf(`"EOF: expected length: %%d, got %%d", %s, L`, length)))
	return fmt.Sprintf(`if L := len(%s); L < %s {
		%s
	}
	`, cc.Slice, length, errReturn)
}

// check for <offset> + <size>, where size is known at compile time
func staticLengthCheckAt(cc gen.Context, size an.BinarySize) string {
	errReturn := cc.ErrReturn(gen.ErrFormated(fmt.Sprintf(`"EOF: expected length: %s, got %%d", L`, cc.Offset.With(size))))
	return fmt.Sprintf(`if L := len(%s); L < %s {
		%s
	}`, cc.Slice, cc.Offset.With(size), errReturn)
}

// check for <offset> + <count>*<size>, where size is known at compile time
func affineLengthCheckAt(cc gen.Context, count gen.Expression, size an.BinarySize) string {
	lengthExpr := cc.Offset.WithAffine(count, size)
	return lengthCheck(cc, lengthExpr)
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

func conditionalLengthCheck(args conditionalLength, cc gen.Context) string {
	out := fmt.Sprintf(`{
		expectedLength := %s
	`, args.baseLength)
	for _, cd := range args.conditions {
		out += fmt.Sprintf(`if %s {
			expectedLength += %d
		}
		`, cd.variableName(), cd.size)
	}
	errReturn := cc.ErrReturn(gen.ErrFormated(fmt.Sprintf(`"EOF: expected length: %%d, got %%d", expectedLength, L`)))
	out += fmt.Sprintf(`if L := len(%s); L < expectedLength {
		%s
		}
	}
	`, cc.Slice, errReturn)
	return out
}

// additional arguments

// additional arguments required by the parsing functions
type argument struct {
	variableName gen.Expression
	typeName     string
}

func (arg argument) asSignature() string {
	return fmt.Sprintf("%s %s", arg.variableName, arg.typeName)
}

// select which arguments to pass to the child function,
// among arguments provided by the parent struct or external
func resolveArguments(itemName string, fi an.Field, requiredArguments []argument) string {
	var args []string

	if len(fi.ArgumentsProvidedByFields) != 0 {
		for i, arg := range fi.ArgumentsProvidedByFields {
			requiredType := requiredArguments[i].typeName
			args = append(args, fmt.Sprintf("%s(%s.%s)", requiredType, itemName, arg))
		}
	} else {
		for _, arg := range requiredArguments {
			args = append(args, arg.variableName)
		}
	}

	return strings.Join(args, ", ")
}

// func argumentsList(arguments []argument) string {
// 	var args []string
// 	for _, arg := range arguments {
// 		args = append(args, arg.variableName)
// 	}
// 	return strings.Join(args, ", ")
// }

func requiredArgs(st an.Struct) (args []argument) {
	for _, field := range st.Fields {
		switch ty := field.Type.(type) {
		case an.Slice:
			if ty.Count == an.NoLength {
				args = append(args, argument{
					variableName: externalCountVariable(field.Name),
					typeName:     "int",
				})
			}
			if elem, isStruct := ty.Elem.(an.Struct); isStruct { // recurse for the child
				args = append(args, requiredArgs(elem)...)
			}
		case an.Union:
			// return the union of the arguments for each member
			all := map[argument]bool{}
			for _, member := range ty.Members {
				for _, arg := range requiredArgs(member) {
					all[arg] = true
				}
			}
			out := make([]argument, 0, len(all))
			for arg := range all {
				out = append(out, arg)
			}
			sort.Slice(out, func(i, j int) bool { return out[i].variableName < out[j].variableName })

			args = append(args, out...)
		case an.Struct: // recurse
			args = append(args, requiredArgs(ty)...)
		}
	}
	// add the user provided one
	for _, arg := range st.Arguments {
		args = append(args, argument{variableName: arg.VariableName, typeName: arg.TypeName})
	}
	return args
}

func externalCountVariable(fieldName string) gen.Expression {
	return strings.ToLower(string(fieldName[0])) + fieldName[1:] + "Count"
}
