package message

import "unicode/utf8"

// Piece is one trimmed part of a split header value.
type Piece struct {
	Text string
	Col  int
}

// Split divides text at each separator, except separators inside quoted text
// when quoted is true. Columns count UTF-16 code units.
func Split(text string, col int, sep byte, quoted bool) []Piece {
	pieces := make([]Piece, 0, 1)
	start := 0
	startCol := col
	column := col
	inQuotes := false

scan:
	for offset := 0; offset < len(text); {
		current := text[offset]
		if current == sep && (!quoted || !inQuotes) {
			pieces = append(pieces, trimmedPiece(text[start:offset], startCol, column))
			offset++
			column = addColumn(column, 1)
			start = offset
			startCol = column
			continue
		}
		if quoted && current == '"' {
			inQuotes = !inQuotes
		}

		r, size := utf8.DecodeRuneInString(text[offset:])
		if size < 1 {
			size = 1
		}
		if (!quoted || !inQuotes) && size > 1 {
			for inner := offset + 1; inner < offset+size; inner++ {
				if text[inner] == sep {
					column = addColumn(column, utf16Units(text[offset:inner]))
					pieces = append(pieces, trimmedPiece(text[start:inner], startCol, column))
					offset = inner + 1
					column = addColumn(column, 1)
					start = offset
					startCol = column
					continue scan
				}
			}
		}

		units := 1
		if r > 0xFFFF {
			units = 2
		}
		column = addColumn(column, units)
		offset += size
	}

	pieces = append(pieces, trimmedPiece(text[start:], startCol, column))
	return pieces
}

func trimmedPiece(text string, startCol, endCol int) Piece {
	start := 0
	end := len(text)
	for start < end && (text[start] == ' ' || text[start] == '\t') {
		start++
	}
	for end > start && (text[end-1] == ' ' || text[end-1] == '\t') {
		end--
	}
	if start == end {
		return Piece{Text: "", Col: endCol}
	}
	return Piece{Text: text[start:end], Col: addColumn(startCol, start)}
}

func utf16Units(text string) int {
	units := 0
	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		if size < 1 {
			size = 1
		}
		if r > 0xFFFF {
			units += 2
		} else {
			units++
		}
		text = text[size:]
	}
	return units
}

func addColumn(column, amount int) int {
	maxInt := int(^uint(0) >> 1)
	if amount > 0 && column > maxInt-amount {
		return maxInt
	}
	return column + amount
}
