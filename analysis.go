package binarygen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"

	"golang.org/x/tools/go/packages"
)

type analyser struct {
	scope *types.Scope // package scope
	ast   *ast.File    // file containing types declaration
	fset  *token.FileSet

	structs map[string]structDef // by name

	pkgName, filePath string
}

// load the source go file with go/packages
func importSource(path string) (analyser, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return analyser{}, err
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedName | packages.NeedFiles,
	}
	tmp, err := packages.Load(cfg, "file="+path)
	if err != nil {
		return analyser{}, err
	}
	if len(tmp) != 1 {
		return analyser{}, fmt.Errorf("multiple packages not supported")
	}

	pkg := tmp[0]
	out := analyser{
		scope:    pkg.Types.Scope(),
		fset:     pkg.Fset,
		pkgName:  pkg.Name,
		filePath: path,
	}
	for _, file := range pkg.Syntax {
		if out.fset.File(file.Pos()).Name() == path {
			out.ast = file
		}
	}

	out.loadStructs()

	return out, nil
}

// one struct definition to handle
type structDef struct {
	underlying *types.Struct
	aliases    map[string]ast.Expr // field name -> type expr
	name       string
}

// return the structs in the given input file
func (an *analyser) loadStructs() {
	an.structs = map[string]structDef{}

	for _, name := range an.scope.Names() {
		obj := an.scope.Lookup(name)

		// filter by input file
		if an.fset.File(obj.Pos()).Name() != an.filePath {
			continue
		}

		switch under := obj.Type().Underlying().(type) {
		case *types.Struct:
			str := structDef{
				underlying: under,
				name:       name,
				aliases:    an.fetchAliases(obj),
			}
			an.structs[name] = str
		}
	}
}

// since go/types remove alias, we use ast to detect them
// return field -> type name
func (an *analyser) fetchAliases(obj types.Object) map[string]ast.Expr {
	pos := obj.Pos()

	var node ast.Node
	ast.Inspect(an.ast, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		if n.Pos() == pos {
			node = n
			return false
		}

		return true
	})

	if node == nil {
		return nil
	}

	out := make(map[string]ast.Expr)
	ast.Inspect(node, func(n ast.Node) bool {
		f, ok := n.(*ast.Field)
		if ok {
			for _, name := range f.Names {
				out[name.Name] = f.Type
			}
		}
		return true
	})
	return out
}

// one or many field whose parsing (or writting)
// is grouped to reduce length checks and allocations
type structChunk interface {
	generateParser(fieldIndex int, srcVar, returnVars, offsetExpression string) string

	generateAppender(fieldIndex int, srcVar, dstSlice string) string
}

// either a basic type or a struct with fixed sized fields
type fixedSizeType interface {
	name() string
	size() int

	// how to implement
	// dstVar = parse(dataSrcVar[offset:])
	generateParser(dstVar, srcSlice string, offset int) string

	// how to implement
	// put srcVar into dstSlice[offset:]
	generateWriter(srcVar, dstSlice string, offset int) string
}

// return an empty string for not named type
func typeName(ty types.Type) string {
	if named, ok := ty.(*types.Named); ok {
		return named.Obj().Name()
	}
	return ""
}

// check is the underlying type as fixed size;
// return nil if not
func (an *analyser) isFixedSize(ty types.Type, typeDecl ast.Expr) fixedSizeType {
	// first check for custom constructor
	// if present, only the constructor type matters
	if wc, ok := an.newWithConstructor(ty, typeDecl); ok { // overide underlying basic info
		return wc
	}

	switch underlying := ty.Underlying().(type) {
	case *types.Basic:
		name := underlying.Name()
		if n := typeName(ty); n != "" {
			name = n
		}
		if L, ok := getBinaryLayout(underlying); ok {
			return basicType{name_: name, binaryLayout: L}
		}
	case *types.Struct:
		named, ok := ty.(*types.Named)
		if !ok {
			panic("anonymous struct not supported")
		}

		fields, ok := an.fixedSizeFromStruct(an.structs[named.Obj().Name()])
		if ok {
			return fixedSizeStruct{
				type_: underlying,
				name_: named.Obj().Name(),
				size_: fields.size(),
			}
		}
	case *types.Array:
		panic("array not supported yet")

	}
	return nil
}

type withConstructor struct {
	name_ string
	size_ int

	isMethod bool // fromUint(), toUint() vs xxxFromUint() xxxtoUint()
}

func (wc withConstructor) name() string {
	return wc.name_
}

func (wc withConstructor) size() int {
	return wc.size_
}

type basicType struct {
	name_ string // named type

	binaryLayout int // underlying
}

func (bt basicType) name() string {
	return bt.name_
}

func (bt basicType) size() int { return bt.binaryLayout }

// a struct with fixed size
type fixedSizeStruct struct {
	type_ *types.Struct // underlying type

	name_ string
	size_ int
}

func (f fixedSizeStruct) name() string {
	return f.name_
}

func (f fixedSizeStruct) size() int {
	return f.size_
}

// how the type is written as binary
type fixedSizeField struct {
	field *types.Var

	type_ fixedSizeType
}

const (
	bytes1 = 1 // byte, int8
	bytes2 = 2 // int16, uint16
	bytes4 = 4 // uint32
	bytes8 = 8 // uint32
)

func getBinaryLayout(t *types.Basic) (int, bool) {
	switch t.Kind() {
	case types.Bool, types.Int8, types.Uint8:
		return bytes1, true
	case types.Int16, types.Uint16:
		return bytes2, true
	case types.Int32, types.Uint32, types.Float32:
		return bytes4, true
	case types.Int64, types.Uint64, types.Float64:
		return bytes8, true
	default:
		return 0, false
	}
}

// return the TypeName or nil if `typeDecl` is not an alias
func (an *analyser) isAlias(typeDecl ast.Expr) *types.TypeName {
	if ident, ok := typeDecl.(*ast.Ident); ok {
		alias := an.scope.Lookup(ident.Name)
		if named, ok := alias.(*types.TypeName); ok && named.IsAlias() {
			return named
		}
	}
	return nil
}

// return the new binary layout, or 0
// if always returns 0 if ty is not a *types.Named
// check for method fromUint() or function xxxFromUint()
func (an *analyser) newWithConstructor(ty types.Type, typeDecl ast.Expr) (withConstructor, bool) {
	layoutFromFunc := func(fnType types.Type) (int, bool) {
		sig, ok := fnType.(*types.Signature)
		if !ok {
			return 0, false
		}
		arg := sig.Params().At(0).Type().Underlying()
		if basic, ok := arg.(*types.Basic); ok {
			return getBinaryLayout(basic)
		}
		return 0, false
	}

	// a type with a method is a named type
	named, ok := ty.(*types.Named)
	if ok {
		for i := 0; i < named.NumMethods(); i++ {
			meth := named.Method(i)
			if meth.Name() == "fromUint" {
				if layout, ok := layoutFromFunc(meth.Type()); ok {
					return withConstructor{name_: typeName(ty), size_: layout, isMethod: true}, true
				}
			}
		}
	}

	// check for aliases
	alias := an.isAlias(typeDecl)
	var tn *types.TypeName
	if alias != nil { // use the alias instead of the underlying type
		tn = alias
	} else if ok {
		tn = named.Obj()
	} else {
		return withConstructor{}, false
	}

	functionName := tn.Name() + "FromUint"
	fn := an.scope.Lookup(functionName)
	if fn != nil {
		if layout, ok := layoutFromFunc(fn.Type()); ok {
			return withConstructor{name_: tn.Name(), size_: layout, isMethod: false}, true
		}
	}

	return withConstructor{}, false
}

type fixedSizeFields []fixedSizeField

// return true is all fields are with fixed size
func (an *analyser) fixedSizeFromStruct(str structDef) (fixedSizeFields, bool) {
	var fixedSize fixedSizeFields
	for i := 0; i < str.underlying.NumFields(); i++ {
		field := str.underlying.Field(i)

		if ft := an.isFixedSize(field.Type(), str.aliases[field.Name()]); ft != nil {
			fixedSize = append(fixedSize, fixedSizeField{field: field, type_: ft})
		} else {
			return fixedSize, false
		}
	}
	return fixedSize, true
}

// returns the total size needed by the fields
func (fs fixedSizeFields) size() int {
	totalSize := 0
	for _, field := range fs {
		totalSize += int(field.type_.size())
	}
	return totalSize
}

type arrayField struct {
	field *types.Var

	sizeLen int
	element fixedSizeType
}

func sliceElement(typeDecl ast.Expr) ast.Expr {
	slice := typeDecl.(*ast.ArrayType)
	return slice.Elt
}

func (an *analyser) newSliceField(field *types.Var, tag string, typeDecl ast.Expr) (arrayField, bool) {
	if fieldType, ok := field.Type().Underlying().(*types.Slice); ok {
		var af arrayField
		af.field = field

		fieldElement := an.isFixedSize(fieldType.Elem(), sliceElement(typeDecl)) // TODO: extract typeDecl
		if fieldElement == nil {
			panic("slice of variable length element are not supported")
		}

		af.element = fieldElement

		tag := reflect.StructTag(tag)
		switch tag.Get("len-size") {
		case "16":
			af.sizeLen = bytes2
		case "32":
			af.sizeLen = bytes4
		case "64":
			af.sizeLen = bytes8
		default:
			panic(fmt.Sprintf("missing tag 'len-size' for %s", field.String()))
		}

		return af, true
	}

	return arrayField{}, false
}

func (an *analyser) analyseStruct(str structDef) (out []structChunk) {
	var fixedSize fixedSizeFields
	st := str.underlying
	for i := 0; i < st.NumFields(); i++ {
		field, tag := st.Field(i), st.Tag(i)
		typeDecl := str.aliases[field.Name()]
		// basic types
		if ft := an.isFixedSize(field.Type(), typeDecl); ft != nil {
			fixedSize = append(fixedSize, fixedSizeField{field: field, type_: ft})
			continue
		}

		// close the current fixedSize array
		if len(fixedSize) != 0 {
			out = append(out, fixedSize)
			fixedSize = nil
		}

		// and try for slice
		af, ok := an.newSliceField(field, tag, typeDecl)
		if ok {
			out = append(out, af)
			continue
		}

		panic(fmt.Sprintf("unsupported field in struct %s", field))
	}

	// close the current fixedSize array
	if len(fixedSize) != 0 {
		out = append(out, fixedSize)
	}

	return out
}
