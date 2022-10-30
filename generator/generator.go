package generator

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/benoitkugler/binarygen/analysis"
)

// Buffer is used to accumulate
// and de-deduplicate function declarations
type Buffer struct {
	seen  map[string]bool
	decls []Declaration
}

func NewBuffer() Buffer {
	return Buffer{seen: map[string]bool{}}
}

func (db *Buffer) Add(decl Declaration) {
	if db.seen[decl.ID] {
		return
	}
	db.decls = append(db.decls, decl)
	db.seen[decl.ID] = true
}

// remove non exported, unused function declaration
func (db Buffer) filterUnused() []Declaration {
	var filtered []Declaration
	for i, decl := range db.decls {
		if strings.Contains(decl.ID, ".") || decl.IsExported {
			filtered = append(filtered, decl)
			continue // always include methods and exported functions
		}

		isUsed := false
		for j, other := range db.decls {
			if i == j {
				continue
			}
			if strings.Contains(other.Content, decl.ID) {
				// the function is used, keep it
				isUsed = true
				break
			}
		}
		if isUsed {
			filtered = append(filtered, decl)
		}
	}
	return filtered
}

// Code removes the unused declaration and returns
// the final code.
func (db Buffer) Code() string {
	var builder strings.Builder
	for _, decl := range db.filterUnused() {
		builder.WriteString(decl.Content + "\n")
	}
	return builder.String()
}

// Declaration is a chunk of generated go code,
// with an id used to avoid duplication
type Declaration struct {
	ID         string
	Content    string
	IsExported bool
}

// Expression is a Go expression, such as a variable name, a static number, or an expression
type Expression = string

// Context holds the names of the objects used in the
// generated code
type Context struct {
	// <variableName> = parse<Type>(<byteSliceName>)
	// <byteSliceName> = append<Type>To(<variableName>, <byteSliceName>)

	// Type is the name of the type being generated
	Type Expression

	// ObjectVar if the name of the variable being parsed or dumped
	ObjectVar Expression

	// Slice is the name of the []byte being read or written
	Slice Expression

	// Offset holds the variable name for the current offset,
	// and its value when known at compile time
	Offset Offset
}

// ErrReturn returns a "return ..., err" statement
func (cc Context) ErrReturn(errVariable string) string {
	return fmt.Sprintf("return %s{}, 0, %s", cc.Type, errVariable)
}

// Selector returns a "<ObjectVar>.<field>" statement
func (cc Context) Selector(field string) string {
	return fmt.Sprintf("%s.%s", cc.ObjectVar, field)
}

// SubSlice slices the current input slice at the current offset
// and assigns it to `subSlice`.
// It also updates the [Context.Slice] field
func (cc *Context) SubSlice(subSlice Expression) string {
	out := fmt.Sprintf("%s := %s[%s:]", subSlice, cc.Slice, cc.Offset.Name)
	cc.Slice = subSlice
	return out
}

// ParsingFunc adds the context to the given [scopes] and [args]
func (cc Context) ParsingFunc(args, scopes []string) Declaration {
	isExported := unicode.IsUpper([]rune(cc.Type)[0])
	funcTitle := "parse"
	if isExported {
		funcTitle = "Parse"
	}
	funcName := funcTitle + strings.Title(string(cc.Type))
	content := fmt.Sprintf(`func %s(%s) (%s, int, error) {
		var %s %s
		%s
		return %s, %s, nil
	}
	`, funcName, strings.Join(args, ","), cc.Type, cc.ObjectVar,
		cc.Type, strings.Join(scopes, "\n"), cc.ObjectVar, cc.Offset.Name)
	return Declaration{ID: funcName, Content: content, IsExported: isExported}
}

// offset management

// Offset represents an offset in a byte of slice.
// It is designed to produce the optimal output regardless
// of whether its value is known at compile time or not.
type Offset struct {
	// Name is the name of the variable containing the offset
	Name Expression

	// part of the value known at compile time, or -1
	value int
}

func NewOffset(name Expression, initialValue int) Offset {
	return Offset{Name: name, value: initialValue}
}

// Value returns the optimal Go expression for the offset current value.
func (of Offset) Value() Expression {
	if of.value != -1 {
		// use the compile time value
		return strconv.Itoa(of.value)
	}

	return of.Name
}

// With returns the optimal expression for <offset> + <size>
func (of Offset) With(size analysis.BinarySize) Expression {
	of.value += int(size)
	return of.Value()
}

// Increment updates the current value, adding [size]
func (of *Offset) Increment(size analysis.BinarySize) {
	of.value += int(size)
}

// UpdateStatement returns a statement for <offset> += size,
// without changing the tracked value, since
// it has already been done with [Increment] calls.
func (of Offset) UpdateStatement(size analysis.BinarySize) Expression {
	return fmt.Sprintf("%s += %d", of.Name, size)
}

// SetStatement returns the code for <offset> = <value>,
// and remove the tracked value wich is now unknown
func (of *Offset) SetStatement(value Expression) Expression {
	of.value = -1
	return fmt.Sprintf("%s = %s", of.Name, value)
}

// ArrayOffset returns the expression for <offset> + <count> * <elementSize>,
// usable for offsets or array length
func ArrayOffset(offset Expression, count Expression, elementSize int) Expression {
	return arrayOffsetExpr(offset, count, elementSize)
}

func arrayOffsetExpr(offset Expression, count Expression, elementSize int) Expression {
	if elementSize == 1 {
		if offset == "0" || offset == "" {
			return count
		}
		return fmt.Sprintf("%s + %s", offset, count)
	} else {
		if offset == "0" || offset == "" {
			return fmt.Sprintf("%s * %d", count, elementSize)
		}
		return fmt.Sprintf("%s + %s * %d", offset, count, elementSize)
	}
}

// // Code returns the optimal Go expression for the offset.
// func (of NamedOffset) Code() Expression {
// 	if of.Value.resolvedValue != -1 {
// 		// use the compile time value instead of the variable name
// 		return Expression(strconv.Itoa(of.Value.resolvedValue))
// 	}

// 	// else if
// 	// if [resolvedValue] is not available it means [part1] is not empty
// 	if of.part2 == 0 { // only use part1
// 		return of.part1
// 	}

// 	return Expression(fmt.Sprintf("%s + %d", of.part1, of.part2))
// }

// type Offset struct {
// 	part1 Expression // unkwown at compile time
// 	part2 int        // known at compile time

// 	resolvedValue int // -1 if not computable
// }

// // NewOffset returns the offset for <variable> + <static>.
// // If [variable] is actually a fixed, static number, it is added to <static>
// func NewOffset(variable Expression, static int) Offset {
// 	// try to resolve the offset at generation time
// 	if variable == "" {
// 		variable = "0"
// 	}
// 	if offset, err := strconv.Atoi(string(variable)); err == nil {
// 		return Offset{part2: offset + static, resolvedValue: offset + static}
// 	}

// 	return Offset{part1: variable, part2: static, resolvedValue: -1}
// }

// // ArrayOffsetInt is a shortcut for [ArrayOffset] with an integer [count]
// func ArrayOffsetInt(offset Expression, count, elementSize int) Offset {
// 	return NewOffset(offset, count*elementSize)
// }
