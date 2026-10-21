package freshness

import (
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

// Directive is one validated Cache-Control directive. Name and Value preserve
// their spelling; HasValue distinguishes a bare name from an explicit value.
type Directive struct {
	Name         string
	Value        string
	HasValue     bool
	DeltaSeconds int64
}

// Control is the validated freshness metadata from a parsed header block.
// Presence is separate from the value so explicit zero values remain distinct
// from absent fields. ExpiresValid is meaningful only when HasExpires is true.
type Control struct {
	Directives      map[string]Directive
	Date            int64
	HasDate         bool
	Expires         int64
	HasExpires      bool
	ExpiresValid    bool
	Age             int64
	HasAge          bool
	LastModified    int64
	HasLastModified bool
}

// ParseControl validates freshness fields in their original file order.
// fields must be a successfully parsed message.Fields collection.
func ParseControl(fields message.Fields) (Control, error) {
	control := Control{Directives: make(map[string]Directive)}
	seenSingleton := make(map[string]bool)

	for _, field := range fields {
		switch asciiLower(field.Name) {
		case "cache-control":
			if err := parseDirectives(&control, field); err != nil {
				return Control{}, err
			}
		case "date", "expires", "age", "last-modified":
			canonical := canonicalFieldName(asciiLower(field.Name))
			key := asciiLower(canonical)
			if seenSingleton[key] {
				return Control{}, problem("HC004", "repeated field "+base.Show(canonical), field.Line, field.Col)
			}
			seenSingleton[key] = true

			switch key {
			case "date":
				value, ok := message.ParseDate(field.Value)
				if !ok {
					return Control{}, problem("HC005", "invalid date "+base.Show(field.Value), field.Line, field.Col)
				}
				control.Date = value
				control.HasDate = true
			case "expires":
				control.Expires, control.ExpiresValid = message.ParseDate(field.Value)
				control.HasExpires = true
			case "age":
				value, ok := parseDigits(field.Value, 10)
				if !ok {
					return Control{}, problem("HC006", "invalid age "+base.Show(field.Value), field.Line, field.Col)
				}
				control.Age = value
				control.HasAge = true
			case "last-modified":
				value, ok := message.ParseDate(field.Value)
				if !ok {
					return Control{}, problem("HC005", "invalid date "+base.Show(field.Value), field.Line, field.Col)
				}
				control.LastModified = value
				control.HasLastModified = true
			}
		}
	}

	return control, nil
}

func parseDirectives(control *Control, field message.Field) error {
	for _, piece := range message.Split(field.Value, field.Col, ',', true) {
		if piece.Text == "" {
			continue
		}

		name, value, hasValue, valid := directiveParts(piece.Text)
		if !valid {
			return problem("HC001", "invalid directive "+base.Show(piece.Text), field.Line, piece.Col)
		}

		foldedName := asciiLower(name)
		var deltaSeconds int64
		if foldedName == "max-age" || foldedName == "s-maxage" {
			valueCol := piece.Col + len(name)
			if hasValue {
				valueCol++
			}
			delta := value
			if !hasValue {
				delta = ""
			}
			var ok bool
			deltaSeconds, ok = parseDigits(delta, 10)
			if !ok {
				return problem("HC002", "invalid delta-seconds "+base.Show(delta), field.Line, valueCol)
			}
		}

		if _, exists := control.Directives[foldedName]; exists {
			return problem("HC003", "repeated directive "+base.Show(name), field.Line, piece.Col)
		}
		control.Directives[foldedName] = Directive{
			Name:         name,
			Value:        value,
			HasValue:     hasValue,
			DeltaSeconds: deltaSeconds,
		}
	}
	return nil
}

func directiveParts(text string) (name, value string, hasValue, valid bool) {
	equals := strings.IndexByte(text, '=')
	if equals < 0 {
		name = text
	} else {
		name = text[:equals]
		value = text[equals+1:]
		hasValue = true
	}
	if !validDirectiveName(name) {
		return name, value, hasValue, false
	}
	if hasValue && !validDirectiveValue(value) {
		return name, value, hasValue, false
	}
	return name, value, hasValue, true
}

func validDirectiveName(text string) bool {
	if text == "" {
		return false
	}
	for i := 0; i < len(text); i++ {
		char := text[i]
		if isASCIILetter(char) || asciiDigit(char) || char == '.' || char == '_' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func validDirectiveValue(text string) bool {
	if text == "" {
		return false
	}
	if text[0] == '"' {
		if len(text) < 2 || text[len(text)-1] != '"' {
			return false
		}
		for i := 1; i < len(text)-1; i++ {
			if text[i] == '"' {
				return false
			}
		}
		return true
	}
	for i := 0; i < len(text); i++ {
		char := text[i]
		if isASCIILetter(char) || asciiDigit(char) || char == '.' || char == '_' || char == '+' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func parseDigits(text string, maxDigits int) (int64, bool) {
	if len(text) == 0 || len(text) > maxDigits {
		return 0, false
	}
	var value int64
	for i := 0; i < len(text); i++ {
		if !asciiDigit(text[i]) {
			return 0, false
		}
		value = value*10 + int64(text[i]-'0')
	}
	return value, true
}

func asciiDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func isASCIILetter(char byte) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')
}

func asciiLower(text string) string {
	var folded strings.Builder
	folded.Grow(len(text))
	for i := 0; i < len(text); i++ {
		char := text[i]
		if char >= 'A' && char <= 'Z' {
			char += 'a' - 'A'
		}
		folded.WriteByte(char)
	}
	return folded.String()
}

func canonicalFieldName(name string) string {
	switch name {
	case "date":
		return "Date"
	case "expires":
		return "Expires"
	case "age":
		return "Age"
	case "last-modified":
		return "Last-Modified"
	default:
		return name
	}
}

func problem(code, text string, line, col int) *message.Problem {
	return &message.Problem{Code: code, Text: text, Line: line, Col: col}
}
