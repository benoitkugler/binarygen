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
	"strings"

	"golang.org/x/tools/go/packages"
)

// Analyser provides information about types,
// shared by the parser and writer code generator
type Analyser struct {
	// Source is the path of the origin go source file.
	Source string

	absSourcePath string

	pkg *packages.Package

	// Sources contains only the structs
	// from the source file
	Sources []*types.Named

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

// load the source go file with go/packages
func importSource(path string) (Analyser, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return Analyser{}, err
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedName | packages.NeedFiles | packages.NeedDeps | packages.NeedImports,
	}
	tmp, err := packages.Load(cfg, "file="+absPath)
	if err != nil {
		return Analyser{}, err
	}
	if len(tmp) != 1 {
		return Analyser{}, fmt.Errorf("multiple packages not supported")
	}

	pkg := tmp[0]
	out := Analyser{
		Source:        path,
		absSourcePath: absPath,
		pkg:           pkg,
	}

	return out, nil
}

// NewAnalyser load the package of `path` and
// analyze the defined structs, filling the fields
// [Source] and [Tables].
func NewAnalyser(path string) (Analyser, error) {
	an, err := importSource(path)
	if err != nil {
		return an, err
	}

	an.fetchSource()
	an.fetchStructsComments()
	an.fetchFieldAliases()
	an.fetchUnionFlags()
	an.fetchInterfaces()
	an.fetchConstructors()

	// perform the actual analysis
	an.Tables = make(map[*types.Named]Struct)
	for _, ty := range an.Sources {
		an.handleTable(ty)
	}

	return an, nil
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
			if an.pkg.Fset.File(obj.Pos()).Name() == an.absSourcePath {
				an.Sources = append(an.Sources, ty.(*types.Named))
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
		for _, st := range an.Sources {
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

func (an *Analyser) handleTable(ty *types.Named) {
	if _, has := an.Tables[ty]; has {
		return
	}

	st := an.createTypeFor(ty, parsedTags{}, nil).(Struct)
	an.Tables[ty] = st
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
		return Offset{target: target, size: offset.binary()}
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
		return an.createFromStruct(ty.(*types.Named))
	case *types.Slice:
		elemDecl := sliceElement(decl)
		// recurse on the element
		elem := an.createTypeFor(under.Elem(), parsedTags{}, elemDecl)
		return Slice{origin: ty, Elem: elem, Count: tags.arrayCount, CountExpr: tags.arrayCountField}
	case *types.Interface:
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
		st := an.createFromStruct(member)

		out.Members = append(out.Members, st)
		out.Flags = append(out.Flags, flag)
	}
	return out
}
