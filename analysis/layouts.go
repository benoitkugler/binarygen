package analysis

import "go/types"

// Type is the common interface for struct field types
// supported by the package.
type Type interface {
	Origin() types.Type
}

// Scope defines one step of parsing/writting,
// which may come from several fields
type Scope struct{}

// Table is a Go struct defining the binary layout
// of a table.
type Table struct {
	Origin *types.Named
	Fields []Field
}

// Field is a struct field, where embeded fields have been resolved
type Field struct {
	Type   Type
	Layout Layout
	Name   string
}

func (an *Analyser) handleType(ty types.Type) Type {
	if out, has := an.Types[ty]; has {
		return out
	}

	out := an.createTypeFor(ty)
	an.Types[ty] = out

	return out
}

func (an *Analyser) createTypeFor(ty types.Type) Type {
	return nil // TODO:
}
