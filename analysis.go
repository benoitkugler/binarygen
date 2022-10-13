package binarygen

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type analyser struct {
	scope *types.Scope // package scope
	ast   *ast.File    // file containing types declaration
	fset  *token.FileSet

	unionTags  map[*types.Named][]*types.Const     // tags for each tag type
	interfaces map[*types.Interface][]*types.Named // members for each interface

	structDefs    map[string]structDef // by name, filled by fetchTables()
	fileTypeNames []string             // filled by fetchTables

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
	// for arrays, how big is specified the length:
	// either _first8, _first16, _first32, _first64
	// or the name of a field, or "-" when provided externally
	len        string
	offsetSize string // for data stored at an offset, how big is the offset storage

	versionField string       // the name of the field containing the union kind
	versionType  *types.Named // the type of the version flag
}

func newFieldTags(st *types.Struct, tag string) fieldTags {
	t := reflect.StructTag(tag)
	out := fieldTags{
		len:          t.Get("len"),
		offsetSize:   t.Get("offset-size"),
		versionField: t.Get("version-field"),
	}
	if out.versionField != "" {
		for i := 0; i < st.NumFields(); i++ {
			if fi := st.Field(i); fi.Name() == out.versionField {
				out.versionType = fi.Type().(*types.Named)
				return out
			}
		}
		panic("unknow field " + out.versionField)
	}
	return out
}

// one struct definition to handle
type structDef struct {
	underlying *types.Struct
	aliases    map[string]ast.Expr // field name -> type expr
	name       string
}

// register the structs in the given input file
func (an *analyser) fetchTables() {
	an.structDefs = map[string]structDef{}
	an.fileTypeNames = nil

	for _, name := range an.scope.Names() {
		obj := an.scope.Lookup(name)

		if tn, isTypeName := obj.(*types.TypeName); isTypeName && tn.IsAlias() {
			// ignore top level aliases
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

			// filter by input file
			if an.fset.File(obj.Pos()).Name() == an.filePath {
				an.fileTypeNames = append(an.fileTypeNames, name)
			}
		}
	}
}

// look for integer constants with type <...>Version
// and values <...>Version<v>,
// which are mapped to concrete types <interfaceName><v>
func (an *analyser) fetchUnionFlags() {
	an.unionTags = make(map[*types.Named][]*types.Const)
	for _, name := range an.scope.Names() {
		obj := an.scope.Lookup(name)

		cst, isConst := obj.(*types.Const)
		if !isConst {
			continue
		}
		if cst.Val().Kind() != constant.Int {
			continue
		}

		named, ok := cst.Type().(*types.Named)
		if !ok {
			continue
		}

		if !strings.HasSuffix(named.Obj().Name(), "Version") {
			continue
		}

		an.unionTags[named] = append(an.unionTags[named], cst)
	}
}

func (an *analyser) fetchInterfaces() {
	an.interfaces = make(map[*types.Interface][]*types.Named)

	names := an.scope.Names()
	var structs []*types.Named
	for _, name := range names {
		obj := an.scope.Lookup(name)

		_, isStruct := obj.Type().Underlying().(*types.Struct)
		if !isStruct {
			continue
		}

		structs = append(structs, obj.Type().(*types.Named))
	}

	for _, name := range names {
		obj := an.scope.Lookup(name)

		itf, isItf := obj.Type().Underlying().(*types.Interface)
		if !isItf {
			continue
		}

		// find the members of this interface
		for _, st := range structs {
			if types.Implements(st, itf) {
				an.interfaces[itf] = append(an.interfaces[itf], st)
			}
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
	an.fetchTables()
	an.fetchUnionFlags()
	an.fetchInterfaces()

	an.structLayouts = make(map[string]structLayout, len(an.structDefs))
	for _, k := range an.fileTypeNames {
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

// if returns false if ty is not a *types.Named
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
	if slice, ok := typeDecl.(*ast.ArrayType); ok {
		return slice.Elt
	}
	return nil
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

type fixedSizeType interface {
	oType

	mustParser(cc codeContext, dstSelector string) string
}

func (sl slice) requiredArgs(fieldName string) []argument {
	if sl.lengthLocation == "" { // provided as function argument
		return []argument{{sl.externalLengthVariable(fieldName), "int"}}
	}
	return nil
}

// return the union of the arguments for each member
func (u union) requiredArgs() []argument {
	all := map[argument]bool{}
	for _, member := range u.members {
		for _, arg := range member.requiredArgs() {
			all[arg] = true
		}
	}
	out := make([]argument, 0, len(all))
	for arg := range all {
		out = append(out, arg)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].variableName < out[j].variableName })
	return out
}

type structField struct {
	type_ oType
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
		_, hasFixedType := field.type_.(fixedSizeType)
		if _, isFixedSize := field.type_.staticSize(); isFixedSize && hasFixedType {
			fixedSize = append(fixedSize, field)
			continue
		}

		// close the current fixedSize array ...
		if len(fixedSize) != 0 {
			out = append(out, fixedSize)
			fixedSize = nil
		}

		// and add a standalone field
		out = append(out, field)
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
		case union:
			args = append(args, ty.requiredArgs()...)
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

func (an *analyser) handleFieldType(ty types.Type, tags fieldTags, astDecl ast.Expr) oType {
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

	// union types
	_, isItf := ty.Underlying().(*types.Interface)
	if isItf {
		return an.handleInterface(ty, tags)
	}

	// named struct
	named, ok := ty.(*types.Named)
	if ok {
		return an.getOrAnalyseStruct(named.Obj().Name())
	}

	panic(fmt.Sprintf("unsupported field type in struct: %s", ty))
}

// check if the underlying type as fixed size;
// return nil if not
func (an *analyser) handleFixedSize(ty types.Type, typeDecl ast.Expr) oType {
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
			return basicType{name_: name, binarySize: L}
		}
	case *types.Array:
		elem := underlying.Elem()
		resolvedElem := an.handleFixedSize(elem, sliceElement(typeDecl))
		if resolvedElem, isElemBasic := resolvedElem.(basicType); isElemBasic {
			return array{element: resolvedElem, length: int(underlying.Len())}
		}
		panic("array with elements of variable size is not supported")
	}
	return nil
}

func (an *analyser) handleSlice(ty types.Type, tags fieldTags, typeDecl ast.Expr) (slice, bool) {
	if fieldType, ok := ty.Underlying().(*types.Slice); ok {
		var sl slice

		sl.lengthLocation = tags.len

		tags.len = ""
		fieldElement := an.handleFieldType(fieldType.Elem(), tags, sliceElement(typeDecl))
		elementTyp, ok := fieldElement.(fixedSizeType)
		if !ok {
			panic("slice of variable length element are not supported")
		}

		sl.element = elementTyp

		return sl, true
	}

	return slice{}, false
}

func (an *analyser) handleInterface(ty types.Type, tags fieldTags) union {
	itf := ty.Underlying().(*types.Interface)
	named, ok := ty.(*types.Named)
	if !ok {
		panic("anonymous interfaces not supported")
	}
	itfName := named.Obj().Name()

	if tags.versionField == "" {
		panic("missing tag version-field for field with type " + itfName)
	}

	out := union{type_: named, flagFieldName: tags.versionField}

	flags := an.unionTags[tags.versionType]
	byVersion := map[string]*types.Const{}
	for _, flag := range flags {
		_, version, _ := strings.Cut(flag.Name(), "Version")
		byVersion[version] = flag
	}

	for _, member := range an.interfaces[itf] {
		memberName := member.Obj().Name()
		version := strings.TrimPrefix(memberName, itfName)
		st := an.getOrAnalyseStruct(memberName)
		flag, ok := byVersion[version]
		if !ok {
			panic(fmt.Sprintf("union flag %sVersion%s not defined", itfName, version))
		}
		out.members = append(out.members, st)
		out.flags = append(out.flags, flag)
	}
	return out
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

	panic("unknown type name " + typeName)
}

func (an *analyser) analyzeStruct(str structDef) (out structLayout) {
	out.name_ = str.name

	st := str.underlying
	for i := 0; i < st.NumFields(); i++ {
		field, tags := st.Field(i), newFieldTags(st, st.Tag(i))

		var sf structField
		sf.name = field.Name()
		astDecl := str.aliases[field.Name()]
		sf.type_ = an.handleFieldType(field.Type(), tags, astDecl)

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
