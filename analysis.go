package binarygen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

type analyser struct {
	scope *types.Scope // package scope
	ast   *ast.File    // file containing types declaration
	fset  *token.FileSet

	structDefs    map[string]structDef    // by name, filled by fetchStructs()
	structLayouts map[string]structLayout // by name, filled by the analysis

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
			break
		}
	}

	return out, nil
}

// additional meta data included as tags in structs definitions
type fieldTags struct {
	lenSize    string // for array, how big is the length storage
	offsetSize string // for data stored at an offset, how big is the offset storage
}

func newFieldTags(tag string) fieldTags {
	t := reflect.StructTag(tag)
	return fieldTags{
		lenSize:    t.Get("len-size"),
		offsetSize: t.Get("offset-size"),
	}
}

// one struct definition to handle
type structDef struct {
	underlying *types.Struct
	aliases    map[string]ast.Expr // field name -> type expr
	name       string
}

// register the structs in the given input file
func (an *analyser) fetchStructs() {
	an.structDefs = map[string]structDef{}

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
			an.structDefs[name] = str
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

func (an *analyser) performAnalysis() {
	an.fetchStructs()

	an.structLayouts = make(map[string]structLayout, len(an.structDefs))
	for k := range an.structDefs {
		an.getOrAnalyseStruct(k) // write the result into structLayouts
	}
}

// return an empty string for not named type
func typeName(ty types.Type) string {
	if named, ok := ty.(*types.Named); ok {
		return named.Obj().Name()
	}
	return ""
}

type withConstructor struct {
	name_ string
	size_ int

	isMethod bool // fromUint(), toUint() vs xxxFromUint() xxxtoUint()
}

func (wc withConstructor) size() int {
	return wc.size_
}

type basicType struct {
	name_ string // named type

	binaryLayout int // underlying
}

func (bt basicType) size() int { return bt.binaryLayout }

// integer offset to the actual data
type offset struct {
	target fieldType

	size_ int
}

func (o offset) size() int {
	return o.size_
}

func (offset) offsetVariableName(field string) string {
	return fmt.Sprintf("offsetTo%s", strings.Title(field))
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
				if layout, isFunc := layoutFromFunc(meth.Type()); isFunc {
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

func sliceElement(typeDecl ast.Expr) ast.Expr {
	slice := typeDecl.(*ast.ArrayType)
	return slice.Elt
}

func sizeFromTag(tag string) int {
	switch tag {
	case "16":
		return bytes2
	case "32":
		return bytes4
	case "64":
		return bytes8
	case "-": // len must be given in argument
		return 0
	default:
		return -1
	}
}

type fieldType interface {
	staticSize() (int, bool)
	name() string

	// parser expression, with no bounds check
	// <objectName>.<dstSelector> = mustParse(<byteSliceName[<offsetName>:])
	mustParser(cc codeContext, dstSelector string) string
}

func (of offset) name() string          { return of.target.name() }
func (wc withConstructor) name() string { return wc.name_ }
func (bt basicType) name() string       { return bt.name_ }
func (sl slice) name() string           { return "[]" + sl.element.name() }
func (sl structLayout) name() string    { return sl.name_ }

func (of offset) staticSize() (int, bool)          { return of.size_, true }
func (wc withConstructor) staticSize() (int, bool) { return wc.size_, true }
func (bt basicType) staticSize() (int, bool)       { return bt.binaryLayout, true }
func (sl slice) staticSize() (int, bool)           { return 0, false }

// staticSize returns the statically known size of the type
// or false if it is dynamic or requires additional length check
func (sl structLayout) staticSize() (int, bool) {
	totalSize := 0
	for _, field := range sl.fields {
		// special case for offsets : they have a static size
		// but still require additional length check
		if _, isOffset := field.type_.(offset); isOffset {
			return 0, false
		}

		size, ok := field.type_.staticSize()
		if !ok {
			return 0, false
		}
		totalSize += size
	}
	return totalSize, true
}

// []<element> slice type
type slice struct {
	element fieldType
	sizeLen int
}

func (sl slice) requiredArgs(fieldName string) []argument {
	if sl.sizeLen == 0 { // provided as function argument
		return []argument{{sl.externalLengthVariable(fieldName), "int"}}
	}
	return nil
}

type structField struct {
	type_ fieldType
	name  string // name of the field
}

// structLayout is the result of the analysis of a Go struct
type structLayout struct {
	name_ string // name of the type

	fields []structField
}

// as an optimization groups the contiguous fixed-size fields
func (sl structLayout) groups() (out []group) {
	var fixedSize fixedSizeList
	for _, field := range sl.fields {
		if _, isFixedSize := field.type_.staticSize(); isFixedSize {
			fixedSize = append(fixedSize, field)
			continue
		}

		// close the current fixedSize array ...
		if len(fixedSize) != 0 {
			out = append(out, fixedSize)
			fixedSize = nil
		}

		// and add a standalone field
		out = append(out, standaloneField(field))
	}

	// close the current fixedSize array if needed
	if len(fixedSize) != 0 {
		out = append(out, fixedSize)
	}

	return out
}

func (st structLayout) requiredArgs() (args []argument) {
	for _, field := range st.fields {
		switch ty := field.type_.(type) {
		case slice:
			args = append(args, ty.requiredArgs(field.name)...)
		case structLayout: // recurse
			args = append(args, ty.requiredArgs()...)
		}
	}
	return args
}

// returns `true` is the type is referenced in other types
func (an *analyser) isTypeReferenced(st structLayout) bool {
	_, has := an.structDefs[st.name_]
	return has
}

func (an *analyser) handleFieldType(ty types.Type, tags fieldTags, astDecl ast.Expr) fieldType {
	// special case for offsets
	if s := sizeFromTag(tags.offsetSize); s != -1 {
		// the field is an offset to the actual data
		tags.offsetSize = ""
		target := an.handleFieldType(ty, tags, astDecl)
		return offset{target: target, size_: s}
	}

	// basic fixed size types
	if ft := an.handleFixedSize(ty, astDecl); ft != nil {
		return ft
	}

	// and try for slice
	af, ok := an.handleSlice(ty, tags, astDecl)
	if ok {
		return af
	}

	// named struct
	named, ok := ty.(*types.Named)
	if ok {
		return an.getOrAnalyseStruct(named.Obj().Name())
	}

	panic(fmt.Sprintf("unsupported field type in struct: %s", ty))
}

// check is the underlying type as fixed size;
// return nil if not
func (an *analyser) handleFixedSize(ty types.Type, typeDecl ast.Expr) fieldType {
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
	case *types.Array:
		panic("array not supported yet")
	}
	return nil
}

func (an *analyser) handleSlice(ty types.Type, tags fieldTags, typeDecl ast.Expr) (slice, bool) {
	if fieldType, ok := ty.Underlying().(*types.Slice); ok {
		var sl slice

		size := sizeFromTag(tags.lenSize)
		if size == -1 {
			panic(fmt.Sprintf("missing tag 'len-size' for type %s", ty))
		}

		tags.lenSize = ""
		fieldElement := an.handleFieldType(fieldType.Elem(), tags, sliceElement(typeDecl))
		if fieldElement == nil {
			panic("slice of variable length element are not supported")
		}

		sl.element = fieldElement
		sl.sizeLen = size

		return sl, true
	}

	return slice{}, false
}

func (an *analyser) getOrAnalyseStruct(typeName string) structLayout {
	if out, has := an.structLayouts[typeName]; has {
		return out
	}

	if def, ok := an.structDefs[typeName]; ok {
		out := an.analyzeStruct(def)
		an.structLayouts[typeName] = out
		return out
	}

	panic("unknown type name" + typeName)
}

func (an *analyser) analyzeStruct(str structDef) (out structLayout) {
	out.name_ = str.name

	st := str.underlying
	for i := 0; i < st.NumFields(); i++ {
		field, tag := st.Field(i), newFieldTags(st.Tag(i))

		var sf structField
		sf.name = field.Name()
		astDecl := str.aliases[field.Name()]
		sf.type_ = an.handleFieldType(field.Type(), tag, astDecl)

		out.fields = append(out.fields, sf)
	}

	return out
}

// additional arguments required by the parsing and writing functions
type argument struct {
	variableName, typeName string
}

func (arg argument) asSignature() string {
	return fmt.Sprintf("%s %s", arg.variableName, arg.typeName)
}

// groups

type fixedSizeList []structField

// returns the total size needed by the fields
func (fs fixedSizeList) size() int {
	totalSize := 0
	for _, field := range fs {
		s, _ := field.type_.staticSize() // by construction, staticSize returns true
		totalSize += s
	}
	return totalSize
}
