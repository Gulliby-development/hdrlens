package summary

import "github.com/Gulliby-development/hdrlens/internal/message"

// Stats contains the summary counts for a collection of header fields.
type Stats struct {
	Total    int
	Distinct int
	Repeated int
	Longest  int
}

// Summarize counts fields, case-insensitive distinct and repeated names, and
// the byte length of the longest value. Parsed header values are ASCII, so
// their byte length is also their character length.
func Summarize(fields message.Fields) Stats {
	result := Stats{Total: len(fields)}
	counts := make(map[string]int, len(fields))
	for _, field := range fields {
		name := asciiLower(field.Name)
		counts[name]++
		if valueLength := len(field.Value); valueLength > result.Longest {
			result.Longest = valueLength
		}
	}
	result.Distinct = len(counts)
	for _, count := range counts {
		if count > 1 {
			result.Repeated++
		}
	}
	return result
}

// asciiLower folds only ASCII uppercase letters, matching header-name rules.
func asciiLower(name string) string {
	folded := []byte(name)
	for i, char := range folded {
		if char >= 'A' && char <= 'Z' {
			folded[i] = char + ('a' - 'A')
		}
	}
	return string(folded)
}
