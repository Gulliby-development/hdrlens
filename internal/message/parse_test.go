package message

import (
	"reflect"
	"testing"
)

func TestParseLiteralFields(t *testing.T) {
	tests := []struct {
		name string
		text string
		want Fields
	}{
		{
			name: "one field",
			text: "Host: a",
			want: Fields{{Name: "Host", Value: "a", Line: 1, Col: 7}},
		},
		{
			name: "comments blank lines CRLF and final CR",
			text: "# note\n\nHost: a\r\nX:\t b \t\r",
			want: Fields{
				{Name: "Host", Value: "a", Line: 3, Col: 7},
				{Name: "X", Value: "b", Line: 4, Col: 5},
			},
		},
		{
			name: "one leading BOM",
			text: "\uFEFFHost: a\n",
			want: Fields{{Name: "Host", Value: "a", Line: 1, Col: 7}},
		},
		{
			name: "empty value after space and tab",
			text: "A: \t",
			want: Fields{{Name: "A", Value: "", Line: 1, Col: 5}},
		},
		{
			name: "first colon separates name and value",
			text: "A:b:c",
			want: Fields{{Name: "A", Value: "b:c", Line: 1, Col: 3}},
		},
		{
			name: "unicode comment and whitespace-only lines",
			text: " \t# café\n \t\n",
			want: Fields{},
		},
		{
			name: "empty text",
			text: "",
			want: Fields{},
		},
		{
			name: "final LF does not create another line",
			text: "A: x\nB: y\n",
			want: Fields{
				{Name: "A", Value: "x", Line: 1, Col: 4},
				{Name: "B", Value: "y", Line: 2, Col: 4},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(test.text)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", test.text, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Parse(%q) = %#v, want %#v", test.text, got, test.want)
			}
		})
	}
}

func TestParseLiteralProblems(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		code      string
		problem   string
		line      int
		column    int
		wantError string
	}{
		{
			name:      "accented character in value",
			text:      "Accept: café",
			code:      "HL001",
			problem:   "unexpected character \"?\"",
			line:      1,
			column:    12,
			wantError: "HL001 unexpected character \"?\" at 1:12",
		},
		{
			name:      "supplementary character column",
			text:      "a: 😀",
			code:      "HL001",
			problem:   "unexpected character \"?\"",
			line:      1,
			column:    4,
			wantError: "HL001 unexpected character \"?\" at 1:4",
		},
		{
			name:      "embedded carriage return",
			text:      "a: b\rc",
			code:      "HL001",
			problem:   "unexpected character \"?\"",
			line:      1,
			column:    5,
			wantError: "HL001 unexpected character \"?\" at 1:5",
		},
		{
			name:      "continuation after skipped lines",
			text:      "# note\n\nHost: a\n  x: y",
			code:      "HL002",
			problem:   "continuation line",
			line:      4,
			column:    1,
			wantError: "HL002 continuation line at 4:1",
		},
		{
			name:      "missing colon includes trailing spaces",
			text:      "Host  ",
			code:      "HL003",
			problem:   "missing colon",
			line:      1,
			column:    7,
			wantError: "HL003 missing colon at 1:7",
		},
		{
			name:      "space before colon is invalid name",
			text:      "Host : a",
			code:      "HL004",
			problem:   "invalid field name \"Host \"",
			line:      1,
			column:    1,
			wantError: "HL004 invalid field name \"Host \" at 1:1",
		},
		{
			name:      "empty field name",
			text:      ": x",
			code:      "HL004",
			problem:   "invalid field name \"\"",
			line:      1,
			column:    1,
			wantError: "HL004 invalid field name \"\" at 1:1",
		},
		{
			name:      "character validation precedes continuation detection",
			text:      " café",
			code:      "HL001",
			problem:   "unexpected character \"?\"",
			line:      1,
			column:    5,
			wantError: "HL001 unexpected character \"?\" at 1:5",
		},
		{
			name:      "ASCII continuation line",
			text:      " bad",
			code:      "HL002",
			problem:   "continuation line",
			line:      1,
			column:    1,
			wantError: "HL002 continuation line at 1:1",
		},
		{
			name:      "missing colon precedes name validation",
			text:      "Bad_Name",
			code:      "HL003",
			problem:   "missing colon",
			line:      1,
			column:    9,
			wantError: "HL003 missing colon at 1:9",
		},
		{
			name:      "invalid name character",
			text:      "Bad_Name: x",
			code:      "HL004",
			problem:   "invalid field name \"Bad_Name\"",
			line:      1,
			column:    1,
			wantError: "HL004 invalid field name \"Bad_Name\" at 1:1",
		},
		{
			name:      "first problem across lines wins",
			text:      "Host\n café",
			code:      "HL003",
			problem:   "missing colon",
			line:      1,
			column:    5,
			wantError: "HL003 missing colon at 1:5",
		},
		{
			name:      "only first of two BOMs is removed",
			text:      "\uFEFF\uFEFFA:b",
			code:      "HL001",
			problem:   "unexpected character \"?\"",
			line:      1,
			column:    1,
			wantError: "HL001 unexpected character \"?\" at 1:1",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fields, err := Parse(test.text)
			if err == nil {
				t.Fatalf("Parse(%q) returned fields %#v, want a problem", test.text, fields)
			}
			problem, ok := err.(*Problem)
			if !ok {
				t.Fatalf("Parse(%q) error has type %T, want *Problem", test.text, err)
			}
			if problem.Code != test.code || problem.Text != test.problem ||
				problem.Line != test.line || problem.Col != test.column {
				t.Fatalf("Parse(%q) problem = %#v, want code %q text %q at %d:%d",
					test.text, problem, test.code, test.problem, test.line, test.column)
			}
			if err.Error() != test.wantError {
				t.Fatalf("Parse(%q) error = %q, want %q", test.text, err.Error(), test.wantError)
			}
		})
	}
}
