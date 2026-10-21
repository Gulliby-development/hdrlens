package byterange

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

// RangeKind identifies the syntax used by a byte-range specification.
type RangeKind int

const (
	// ClosedRange has both a first and last byte.
	ClosedRange RangeKind = iota
	// OpenRange starts at First and continues to the end of the resource.
	OpenRange
	// SuffixRange selects the last Suffix bytes.
	SuffixRange
)

// RangeSpec is one parsed byte-range specification.
type RangeSpec struct {
	Kind   RangeKind
	First  int64
	Last   int64
	Suffix int64
}

// ParsedRanges preserves whether a Range field was present separately from its
// specifications, which can all resolve to no bytes.
type ParsedRanges struct {
	Present bool
	Specs   []RangeSpec
}

// Parse validates every Range field in file order.
func Parse(fields message.Fields) (ParsedRanges, error) {
	parsed := ParsedRanges{Specs: make([]RangeSpec, 0)}
	for _, field := range fields.Named("Range") {
		equals := strings.IndexByte(field.Value, '=')
		if equals < 0 {
			return ParsedRanges{}, rangeProblem(field, "HR001", "missing \"=\"", field.Col)
		}

		unit := field.Value[:equals]
		if !asciiEqualFold(unit, "bytes") {
			return ParsedRanges{}, rangeProblem(field, "HR002", "unsupported unit "+base.Show(unit), field.Col)
		}

		parsed.Present = true
		list := field.Value[equals+1:]
		listCol := field.Col + equals + 1
		pieces := message.Split(list, listCol, ',', false)
		validCount := 0
		for _, piece := range pieces {
			if piece.Text == "" {
				continue
			}
			spec, ok := parseRangeSpec(piece.Text)
			if !ok {
				return ParsedRanges{}, rangeProblem(field, "HR003", "invalid range "+base.Show(piece.Text), piece.Col)
			}
			if spec.Kind == ClosedRange && spec.First > spec.Last {
				return ParsedRanges{}, rangeProblem(field, "HR004", "reversed range "+base.Show(piece.Text), piece.Col)
			}
			parsed.Specs = append(parsed.Specs, spec)
			validCount++
		}
		if validCount == 0 {
			return ParsedRanges{}, rangeProblem(field, "HR005", "no range", listCol)
		}
	}
	return parsed, nil
}

// ParseLength validates LENGTH and returns its numeric value.
func ParseLength(value string) (int64, error) {
	if value == "0" {
		return 0, nil
	}
	if len(value) == 0 || len(value) > 15 || value[0] < '1' || value[0] > '9' {
		return 0, errors.New("invalid length " + base.Show(value))
	}
	number, ok := boundedDecimal(value)
	if !ok {
		return 0, errors.New("invalid length " + base.Show(value))
	}
	return number, nil
}

func rangeProblem(field message.Field, code, text string, col int) error {
	return &message.Problem{Code: code, Text: text, Line: field.Line, Col: col}
}

func parseRangeSpec(text string) (RangeSpec, bool) {
	if strings.HasPrefix(text, "-") {
		suffix, ok := boundedDecimal(text[1:])
		if !ok {
			return RangeSpec{}, false
		}
		return RangeSpec{Kind: SuffixRange, Suffix: suffix}, true
	}

	dash := strings.IndexByte(text, '-')
	if dash < 0 {
		return RangeSpec{}, false
	}
	first, ok := boundedDecimal(text[:dash])
	if !ok {
		return RangeSpec{}, false
	}
	if dash == len(text)-1 {
		return RangeSpec{Kind: OpenRange, First: first}, true
	}
	last, ok := boundedDecimal(text[dash+1:])
	if !ok {
		return RangeSpec{}, false
	}
	return RangeSpec{Kind: ClosedRange, First: first, Last: last}, true
}

// boundedDecimal accepts one to fifteen ASCII digits, including leading zeros.
func boundedDecimal(text string) (int64, bool) {
	if len(text) == 0 || len(text) > 15 {
		return 0, false
	}
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return 0, false
		}
	}
	value, err := strconv.ParseInt(text, 10, 64)
	return value, err == nil
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
