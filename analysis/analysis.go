// Package analysis uses go/types and go/packages to extract
// information about the structures to convert to binary form.
package analysis

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Analyser provides information about types,
// shared by the parser and writer code generator
type Analyser struct {
	sourcePath string

	pkg *packages.Package

	// Sources contains only the structs
	// from the source file
	Sources []*types.Named

	// Types contains all the types encoutered when
	// dealing with [Source]
	Types map[types.Type]Type

	// additional information used to retrieve aliases
	forAliases syntaxFieldTypes

	// additional special directives provided by comments
	commentsMap map[*types.Named]commments

	// used to link union member and indicator flag
	unionTags map[*types.Named][]*types.Const

	// get the structs which are member of an interface
	interfaces map[*types.Interface][]*types.Named
}

// NewAnalyser load the package of `path` and
// analyze the defined structs, filling the fields
// [Source] and [Types].
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

	// perform the actual analysis
	an.Types = make(map[types.Type]Type)
	for _, ty := range an.Sources {
		an.handleType(ty)
	}

	return an, nil
}

// load the source go file with go/packages
func importSource(path string) (Analyser, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return Analyser{}, err
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedName | packages.NeedFiles | packages.NeedDeps | packages.NeedImports,
	}
	tmp, err := packages.Load(cfg, "file="+path)
	if err != nil {
		return Analyser{}, err
	}
	if len(tmp) != 1 {
		return Analyser{}, fmt.Errorf("multiple packages not supported")
	}

	pkg := tmp[0]
	out := Analyser{
		sourcePath: path,
		pkg:        pkg,
	}

	return out, nil
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
			if an.pkg.Fset.File(obj.Pos()).Name() == an.sourcePath {
				an.Sources = append(an.Sources, ty.(*types.Named))
			}
		}
	}
}

// look for integer constants with type <...>Version
// and values <...>Version<v>,
// which are mapped to concrete types <interfaceName><v>
func (an *Analyser) fetchUnionFlags() {
	an.unionTags = make(map[*types.Named][]*types.Const)

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

		an.unionTags[named] = append(an.unionTags[named], cst)
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
