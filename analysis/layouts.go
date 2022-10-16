package analysis

import "go/types"

// Type is the common interface for struct field types
// supported by the package.
type Type interface {
	Origin() types.Type
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
