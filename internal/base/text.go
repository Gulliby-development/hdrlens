package base

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Show returns s in double quotes, replacing each non-printable-ASCII rune
// with one question mark.
func Show(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		if r < ' ' || r > '~' {
			b.WriteByte('?')
		} else {
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// Line formats a key and value with the key padded to at least ten runes.
func Line(key, value string) string {
	padding := 10 - utf8.RuneCountInString(key)
	if padding < 0 {
		padding = 0
	}
	var b strings.Builder
	b.WriteString(key)
	for i := 0; i < padding; i++ {
		b.WriteByte(' ')
	}
	b.WriteString("  ")
	b.WriteString(value)
	return b.String()
}

// Table formats a header and rows as left-aligned columns separated by two
// spaces. Missing cells in ragged rows are treated as empty strings.
func Table(header []string, rows [][]string) []string {
	columnCount := len(header)
	for _, row := range rows {
		if len(row) > columnCount {
			columnCount = len(row)
		}
	}
	widths := make([]int, columnCount)
	for i, cell := range header {
		widths[i] = utf8.RuneCountInString(cell)
	}
	for _, row := range rows {
		for i, cell := range row {
			if width := utf8.RuneCountInString(cell); width > widths[i] {
				widths[i] = width
			}
		}
	}

	lines := make([]string, 0)
	lines = append(lines, tableLine(header, widths))
	for _, row := range rows {
		lines = append(lines, tableLine(row, widths))
	}
	return lines
}

func tableLine(cells []string, widths []int) string {
	var b strings.Builder
	for i, width := range widths {
		if i > 0 {
			b.WriteString("  ")
		}
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteString(cell)
		for padding := width - utf8.RuneCountInString(cell); padding > 0; padding-- {
			b.WriteByte(' ')
		}
	}
	return strings.TrimRight(b.String(), " ")
}

// Plural formats n and appends the plural suffix unless n is exactly one.
func Plural(n int, word string) string {
	if n == 1 {
		return strconv.Itoa(n) + " " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}
