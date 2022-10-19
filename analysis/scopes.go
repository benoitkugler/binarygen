package analysis

// Scope defines one step of parsing/writting,
// which may come from several fields.
// It is an optimisation to reduce length checks
type Scope interface {
	isScope()
}

func (SingleField) isScope()       {}
func (StaticSizedFields) isScope() {}

type SingleField Field

// StaticSizedFields is a list of fields which all have a static size.
type StaticSizedFields []Field

func (st Struct) Scopes() (out []Scope) {
	// as an optimization groups the contiguous fixed-size fields
	var fixedSize StaticSizedFields
	for _, field := range st.Fields {
		// append to the static fields
		if _, isFixedSize := field.Type.IsFixedSize(); isFixedSize {
			fixedSize = append(fixedSize, field)
			continue
		}

		// else, close the current fixedSize array ...
		if len(fixedSize) != 0 {
			out = append(out, fixedSize)
			fixedSize = nil
		}

		// and add a standalone field
		out = append(out, SingleField(field))
	}

	// close the current fixedSize array if needed
	if len(fixedSize) != 0 {
		out = append(out, fixedSize)
	}

	return out
}
