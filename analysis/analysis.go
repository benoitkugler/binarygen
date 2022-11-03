// Package analysis uses go/types and go/packages to extract
// information about the structures to convert to binary form.
package analysis

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Analyser provides information about types,
// shared by the parser and writer code generator
type Analyser struct {
	// Source is the path of the origin go source file.
	Source string

	sourceAbsPath string

	pkg *packages.Package

	// Sources contains only the structs
	// from the source file
	sources []*types.Named

	// Tables contains the resolved struct definitions, coming from [Sources]
	Tables map[*types.Named]Struct

	// additional information used to retrieve aliases
	forAliases syntaxFieldTypes

	// additional special directives provided by comments
	commentsMap map[*types.Named]commments

	// used to link union member and indicator flag :
	// constant type -> constant values
	unionFlags map[*types.Named][]*types.Const

	// get the structs which are member of an interface
	interfaces map[*types.Interface][]*types.Named

	// map type string to data storage
	constructors map[string]*types.Basic
}

// ImportSource loads the source go file with go/packages,
// also returning the absolute path.
func ImportSource(path string) (*packages.Package, string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, "", err
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedName | packages.NeedFiles | packages.NeedDeps | packages.NeedImports,
	}
	tmp, err := packages.Load(cfg, "file="+absPath)
	if err != nil {
		return nil, "", err
	}
	if len(tmp) != 1 {
		return nil, "", fmt.Errorf("multiple packages not supported")
	}

	return tmp[0], absPath, nil
}

// NewAnalyserFromPkg uses [pkg] to analyse the tables defined in
// [sourcePath].
func NewAnalyserFromPkg(pkg *packages.Package, sourcePath, sourceAbsPath string) Analyser {
	an := Analyser{
		Source:        sourcePath,
		sourceAbsPath: sourceAbsPath,
		pkg:           pkg,
	}

	an.fetchSource()
	an.fetchStructsComments()
	an.fetchFieldAliases()
	an.fetchUnionFlags()
	an.fetchInterfaces()
	an.fetchConstructors()

	// perform the actual analysis
	an.Tables = make(map[*types.Named]Struct)
	for _, ty := range an.sources {
		an.handleTable(ty)
	}

	return an
}

// NewAnalyser load the package of `path` and
// analyze the defined structs, filling the fields
// [Source] and [Tables].
func NewAnalyser(path string) (Analyser, error) {
	pkg, absPath, err := ImportSource(path)
	if err != nil {
		return Analyser{}, err
	}

	return NewAnalyserFromPkg(pkg, path, absPath), nil
}

type syntaxFieldTypes = map[*types.Named]map[string]ast.Expr

func getSyntaxFields(scope *types.Scope, ty *ast.TypeSpec, st *ast.StructType) (*types.Named, map[string]ast.Expr) {
	named := scope.Lookup(ty.Name.Name).Type().(*types.Named)
	fieldTypes := make(map[string]ast.Expr)
	for _, field := range st.Fields.List {
		for _, fieldName := range field.Names {
			fieldTypes[fieldName.Name] = field.Type
		}
	}
	return named, fieldTypes
}

func (an *Analyser) PackageName() string { return an.pkg.Name }

// go/types erase alias information, so we add it in a preliminary step
func (an *Analyser) fetchFieldAliases() {
	an.forAliases = make(syntaxFieldTypes)
	scope := an.pkg.Types.Scope()
	for _, file := range an.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			if ty, isType := n.(*ast.TypeSpec); isType {
				if st, isStruct := ty.Type.(*ast.StructType); isStruct {
					named, tys := getSyntaxFields(scope, ty, st)
					an.forAliases[named] = tys
					return false
				}
			}
			return true
		})
	}
}

func (an *Analyser) fetchStructsComments() {
	an.commentsMap = make(map[*types.Named]commments)
	scope := an.pkg.Types.Scope()
	for _, file := range an.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			if decl, isDecl := n.(*ast.GenDecl); isDecl {
				if len(decl.Specs) != 1 {
					return true
				}
				n = decl.Specs[0]
				if ty, isType := n.(*ast.TypeSpec); isType {
					typ := scope.Lookup(ty.Name.Name).Type()
					if named, ok := typ.(*types.Named); ok {
						an.commentsMap[named] = parseComments(decl.Doc)
					}
					return false
				}
			}
			return true
		})
	}
}

// register the structs in the given input file
func (an *Analyser) fetchSource() {
	scope := an.pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)

		if tn, isTypeName := obj.(*types.TypeName); isTypeName && tn.IsAlias() {
			// ignore top level aliases
			continue
		}
		ty := obj.Type()
		if _, isStruct := ty.Underlying().(*types.Struct); isStruct {
			// filter by input file
			if an.pkg.Fset.File(obj.Pos()).Name() == an.sourceAbsPath {
				an.sources = append(an.sources, ty.(*types.Named))
			}
		}
	}
}

// look for integer constants with type <...>Version
// and values <...>Version<v>,
// which are mapped to concrete types <interfaceName><v>
func (an *Analyser) fetchUnionFlags() {
	an.unionFlags = make(map[*types.Named][]*types.Const)

	scope := an.pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)

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

		an.unionFlags[named] = append(an.unionFlags[named], cst)
	}
}

func (an *Analyser) fetchInterfaces() {
	an.interfaces = make(map[*types.Interface][]*types.Named)

	scope := an.pkg.Types.Scope()
	names := scope.Names()

	for _, name := range names {
		obj := scope.Lookup(name)

		itf, isItf := obj.Type().Underlying().(*types.Interface)
		if !isItf {
			continue
		}

		// find the members of this interface
		for _, st := range an.sources {
			if types.Implements(st, itf) {
				an.interfaces[itf] = append(an.interfaces[itf], st)
			}
		}
	}
}

func (an *Analyser) fetchConstructors() {
	an.constructors = make(map[string]*types.Basic)

	scope := an.pkg.Types.Scope()
	names := scope.Names()

	for _, name := range names {
		obj := scope.Lookup(name)

		fn, isFunction := obj.(*types.Func)
		if !isFunction {
			continue
		}

		sig := fn.Type().(*types.Signature)

		// look for <...>FromUint and patterns
		if typeName, _, ok := strings.Cut(fn.Name(), "FromUint"); ok {
			if sig.Params().Len() != 1 {
				panic("invalid signature for constructor " + fn.Name())
			}
			arg := sig.Params().At(0).Type().(*types.Basic)
			an.constructors[typeName] = arg
		}
	}
}

// handle table wraps [createFromStruct] by registering
// the type in [Tables]
func (an *Analyser) handleTable(ty *types.Named) Struct {
	if st, has := an.Tables[ty]; has {
		return st
	}

	st := an.createFromStruct(ty)
	an.Tables[ty] = st
	return st
}

// resolveName returns a string for the name of the given type,
// using [decl] to preserve aliases.
func (an *Analyser) resolveName(ty types.Type, decl ast.Expr) string {
	// check if we have an alias
	if ident, ok := decl.(*ast.Ident); ok {
		alias := an.pkg.Types.Scope().Lookup(ident.Name)
		if named, ok := alias.(*types.TypeName); ok && named.IsAlias() {
			return named.Name()
		}
	}
	// otherwise use the short name for Named types
	if named, ok := ty.(*types.Named); ok {
		return named.Obj().Name()
	}
	// defaut to the general string representation
	return ty.String()
}

// sliceElement returns the ast for the declaration of a slice or array element,
// or nil if the slice is for instance defined by a named type
func sliceElement(typeDecl ast.Expr) ast.Expr {
	if slice, ok := typeDecl.(*ast.ArrayType); ok {
		return slice.Elt
	}
	return nil
}

// createTypeFor analyse the given type `ty`.
// When it is found on a struct field, `tags` gives additional metadata.
// `decl` matches the syntax declaration of `ty` so that aliases
// can be retrieved.
func (an *Analyser) createTypeFor(ty types.Type, tags parsedTags, decl ast.Expr) Type {
	// first deals with special cases, defined by tags
	if tags.isOpaque {
		return Opaque{origin: ty}
	}

	if offset := tags.offsetSize; offset != 0 {
		// adjust the tags and "recurse" to the actual type
		tags.offsetSize = NoOffset
		target := an.createTypeFor(ty, tags, decl)
		if _, isFixedSize := target.IsFixedSize(); isFixedSize {
			panic("offset to fixed size type is not supported")
		}
		return Offset{Target: target, Size: offset.binary()}
	}

	// now inspect the actual go type
	switch under := ty.Underlying().(type) {
	case *types.Basic:
		return an.createFromBasic(ty, decl)
	case *types.Array:
		elemDecl := sliceElement(decl)
		// recurse on the element
		elem := an.createTypeFor(under.Elem(), parsedTags{}, elemDecl)
		return Array{origin: ty, Len: int(under.Len()), Elem: elem}
	case *types.Struct:
		// anonymous structs are not supported
		return an.handleTable(ty.(*types.Named))
	case *types.Slice:
		elemDecl := sliceElement(decl)
		// recurse on the element
		elem := an.createTypeFor(under.Elem(), parsedTags{}, elemDecl)
		return Slice{origin: ty, Elem: elem, Count: tags.arrayCount, CountExpr: tags.arrayCountField}
	case *types.Interface:
		if tags.unionField == nil {
			panic(fmt.Sprintf("union field with type %s is missing unionField tag", ty))
		}
		// anonymous interface are not supported
		return an.createFromInterface(ty.(*types.Named), tags.unionField)
	default:
		panic(fmt.Sprintf("unsupported type %s", under))
	}
}

// [ty] has underlying type Basic
func (an *Analyser) createFromBasic(ty types.Type, decl ast.Expr) Type {
	// check for custom constructors
	name := an.resolveName(ty, decl)
	if binaryType, hasConstructor := an.constructors[name]; hasConstructor {
		size, _ := newBinarySize(binaryType)
		return DerivedFromBasic{origin: ty, Name: name, Size: size}
	}

	return Basic{origin: ty}
}

func (an *Analyser) createFromStruct(ty *types.Named) Struct {
	st := ty.Underlying().(*types.Struct)
	out := Struct{
		origin: ty,
		Fields: make([]Field, st.NumFields()),
	}

	cm := an.commentsMap[ty]
	if cm.startingOffset != "" {
		// we only support integer shift for now
		so, err := strconv.Atoi(cm.startingOffset)
		if err != nil {
			panic(fmt.Sprintf("unsupported startingOffset %s: %s", cm.startingOffset, err))
		}
		out.StartingOffset = so
	}

	for i := range out.Fields {
		field := st.Field(i)

		// process the struct tags
		tags := newTags(st, reflect.StructTag(st.Tag(i)))

		astDecl := an.forAliases[ty][field.Name()]

		fieldType := an.createTypeFor(field.Type(), tags, astDecl)

		out.Fields[i] = Field{
			Name:   field.Name(),
			Type:   fieldType,
			Layout: Layout{SubsliceStart: tags.subsliceStart},
		}
	}

	return out
}

func (an *Analyser) createFromInterface(ty *types.Named, unionField *types.Var) Union {
	itfName := ty.Obj().Name()
	itf := ty.Underlying().(*types.Interface)
	flags := an.unionFlags[unionField.Type().(*types.Named)]
	members := an.interfaces[itf]

	// this can't be correct in practice
	if len(members) == 0 {
		panic(fmt.Sprintf("interface %s does not have any member", itfName))
	}
	// match flags and members
	byVersion := map[string]*types.Const{}
	for _, flag := range flags {
		_, version, _ := strings.Cut(flag.Name(), "Version")
		byVersion[version] = flag
	}

	out := Union{origin: ty, FlagField: unionField.Name()}
	for _, member := range members {
		memberName := member.Obj().Name()
		// fetch the associated flag
		version := strings.TrimPrefix(memberName, itfName)
		flag, ok := byVersion[version]
		if !ok {
			panic(fmt.Sprintf("union flag %sVersion%s not defined", itfName, version))
		}
		// analyse the concrete type
		st := an.handleTable(member)

		out.Members = append(out.Members, st)
		out.Flags = append(out.Flags, flag)
	}
	return out
}
