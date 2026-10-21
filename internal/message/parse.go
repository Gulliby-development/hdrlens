package message

import (
	"strings"
	"unicode/utf16"

	"github.com/Gulliby-development/hdrlens/internal/base"
)

// Parse parses header fields and returns the first positioned syntax problem.
func Parse(text string) (Fields, error) {
	return parseNormalized(strings.TrimPrefix(text, "\uFEFF"))
}

// parseNormalized parses text after the caller has handled any input BOM.
func parseNormalized(text string) (Fields, error) {
	fields := make(Fields, 0)
	if text == "" {
		return fields, nil
	}

	lineNumber := 1
	lineStart := 0
	for index := 0; index < len(text); index++ {
		if text[index] != '\n' {
			continue
		}
		field, include, err := parseFieldLine(text[lineStart:index], lineNumber)
		if err != nil {
			return nil, err
		}
		if include {
			fields = append(fields, field)
		}
		lineStart = index + 1
		lineNumber++
	}
	if lineStart < len(text) {
		field, include, err := parseFieldLine(text[lineStart:], lineNumber)
		if err != nil {
			return nil, err
		}
		if include {
			fields = append(fields, field)
		}
	}
	return fields, nil
}

func parseFieldLine(line string, lineNumber int) (Field, bool, error) {
	line = strings.TrimSuffix(line, "\r")
	first := 0
	for first < len(line) && (line[first] == ' ' || line[first] == '\t') {
		first++
	}
	if first == len(line) || line[first] == '#' {
		return Field{}, false, nil
	}

	column := 1
	for _, char := range line {
		if char != '\t' && (char < ' ' || char > '~') {
			return Field{}, false, &Problem{
				Code: "HL001",
				Text: "unexpected character " + base.Show(string(char)),
				Line: lineNumber,
				Col:  column,
			}
		}
		column += len(utf16.Encode([]rune{char}))
	}

	if line[0] == ' ' || line[0] == '\t' {
		return Field{}, false, &Problem{
			Code: "HL002",
			Text: "continuation line",
			Line: lineNumber,
			Col:  1,
		}
	}

	colon := strings.IndexByte(line, ':')
	if colon < 0 {
		return Field{}, false, &Problem{
			Code: "HL003",
			Text: "missing colon",
			Line: lineNumber,
			Col:  utf16ColumnCount(line) + 1,
		}
	}

	name := line[:colon]
	if name == "" || !validFieldName(name) {
		return Field{}, false, &Problem{
			Code: "HL004",
			Text: "invalid field name " + base.Show(name),
			Line: lineNumber,
			Col:  1,
		}
	}

	valueStart := colon + 1
	for valueStart < len(line) && (line[valueStart] == ' ' || line[valueStart] == '\t') {
		valueStart++
	}
	valueEnd := len(line)
	for valueEnd > valueStart && (line[valueEnd-1] == ' ' || line[valueEnd-1] == '\t') {
		valueEnd--
	}
	valueCol := valueStart + 1
	if valueStart == valueEnd {
		valueCol = len(line) + 1
	}
	return Field{
		Name:  name,
		Value: line[valueStart:valueEnd],
		Line:  lineNumber,
		Col:   valueCol,
	}, true, nil
}

func validFieldName(name string) bool {
	for index := 0; index < len(name); index++ {
		char := name[index]
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-' {
			continue
		}
		return false
	}
	return true
}

func utf16ColumnCount(text string) int {
	count := 0
	for _, char := range text {
		count += len(utf16.Encode([]rune{char}))
	}
	return count
}
