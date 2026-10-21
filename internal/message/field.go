package message

// Field is one parsed header field, including its original name and position.
type Field struct {
	Name, Value string
	Line, Col   int
}

// Fields is an ordered collection of header fields.
type Fields []Field

// Named returns the fields whose names match name using ASCII case folding.
func (fields Fields) Named(name string) Fields {
	matches := make(Fields, 0)
	for _, field := range fields {
		if asciiEqualFold(field.Name, name) {
			matches = append(matches, field)
		}
	}
	return matches
}

func asciiEqualFold(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := 0; i < len(left); i++ {
		l, r := left[i], right[i]
		if l >= 'A' && l <= 'Z' {
			l += 'a' - 'A'
		}
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		if l != r {
			return false
		}
	}
	return true
}
