package byterange

import (
	"reflect"
	"testing"

	"github.com/Gulliby-development/hdrlens/internal/message"
)

func TestParseInvalidRanges(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		code      string
		text      string
		line      int
		col       int
		wantError string
	}{
		{
			name:      "missing equals",
			input:     "Range: 1-2",
			code:      "HR001",
			text:      "missing \"=\"",
			line:      1,
			col:       8,
			wantError: "HR001 missing \"=\" at 1:8",
		},
		{
			name:      "empty unit",
			input:     "Range: =1-2",
			code:      "HR002",
			text:      "unsupported unit \"\"",
			line:      1,
			col:       8,
			wantError: "HR002 unsupported unit \"\" at 1:8",
		},
		{
			name:      "unit is not trimmed",
			input:     "Range: bytes =1-2",
			code:      "HR002",
			text:      "unsupported unit \"bytes \"",
			line:      1,
			col:       8,
			wantError: "HR002 unsupported unit \"bytes \" at 1:8",
		},
		{
			name:      "invalid second spec",
			input:     "Range: bytes=1-2, 3",
			code:      "HR003",
			text:      "invalid range \"3\"",
			line:      1,
			col:       19,
			wantError: "HR003 invalid range \"3\" at 1:19",
		},
		{
			name:      "dash alone",
			input:     "Range: bytes=-",
			code:      "HR003",
			text:      "invalid range \"-\"",
			line:      1,
			col:       14,
			wantError: "HR003 invalid range \"-\" at 1:14",
		},
		{
			name:      "reversed closed range",
			input:     "Range: bytes=5-2",
			code:      "HR004",
			text:      "reversed range \"5-2\"",
			line:      1,
			col:       14,
			wantError: "HR004 reversed range \"5-2\" at 1:14",
		},
		{
			name:      "empty range list",
			input:     "Range: bytes=",
			code:      "HR005",
			text:      "no range",
			line:      1,
			col:       14,
			wantError: "HR005 no range at 1:14",
		},
		{
			name:      "only empty comma pieces",
			input:     "Range: bytes= , ,",
			code:      "HR005",
			text:      "no range",
			line:      1,
			col:       14,
			wantError: "HR005 no range at 1:14",
		},
		{
			name:      "sixteen digit endpoint",
			input:     "Range: bytes=1000000000000000-",
			code:      "HR003",
			text:      "invalid range \"1000000000000000-\"",
			line:      1,
			col:       14,
			wantError: "HR003 invalid range \"1000000000000000-\" at 1:14",
		},
		{
			name:      "leading plus is invalid",
			input:     "Range: bytes=+1-2",
			code:      "HR003",
			text:      "invalid range \"+1-2\"",
			line:      1,
			col:       14,
			wantError: "HR003 invalid range \"+1-2\" at 1:14",
		},
		{
			name:      "space within endpoint is invalid",
			input:     "Range: bytes=1 -2",
			code:      "HR003",
			text:      "invalid range \"1 -2\"",
			line:      1,
			col:       14,
			wantError: "HR003 invalid range \"1 -2\" at 1:14",
		},
		{
			name:      "later field error wins in file order",
			input:     "Range: bytes=0-1\nRange: bytes=5-2",
			code:      "HR004",
			text:      "reversed range \"5-2\"",
			line:      2,
			col:       14,
			wantError: "HR004 reversed range \"5-2\" at 2:14",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fields, err := message.Parse(test.input)
			if err != nil {
				t.Fatalf("message.Parse(%q) returned error: %v", test.input, err)
			}
			parsed, err := Parse(fields)
			if err == nil {
				t.Fatalf("Parse(%q) returned %#v, want a problem", test.input, parsed)
			}
			problem, ok := err.(*message.Problem)
			if !ok {
				t.Fatalf("Parse(%q) error has type %T, want *message.Problem", test.input, err)
			}
			if problem.Code != test.code || problem.Text != test.text ||
				problem.Line != test.line || problem.Col != test.col {
				t.Fatalf("Parse(%q) problem = %#v, want code %q text %q at %d:%d",
					test.input, problem, test.code, test.text, test.line, test.col)
			}
			if err.Error() != test.wantError {
				t.Fatalf("Parse(%q) error = %q, want %q", test.input, err.Error(), test.wantError)
			}
		})
	}
}

func TestParseValidRanges(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ParsedRanges
	}{
		{
			name:  "closed range with leading zeros",
			input: "Range: bytes=01-02",
			want: ParsedRanges{
				Present: true,
				Specs:   []RangeSpec{{Kind: ClosedRange, First: 1, Last: 2}},
			},
		},
		{
			name:  "open range",
			input: "Range: bytes=1-",
			want: ParsedRanges{
				Present: true,
				Specs:   []RangeSpec{{Kind: OpenRange, First: 1}},
			},
		},
		{
			name:  "suffix range",
			input: "Range: bytes=-2",
			want: ParsedRanges{
				Present: true,
				Specs:   []RangeSpec{{Kind: SuffixRange, Suffix: 2}},
			},
		},
		{
			name:  "case insensitive field and unit with empty list items",
			input: "rAnGe: BYTES=,01-02,,",
			want: ParsedRanges{
				Present: true,
				Specs:   []RangeSpec{{Kind: ClosedRange, First: 1, Last: 2}},
			},
		},
		{
			name:  "multiple fields retain spec order",
			input: "Host: a\nRange: bytes=1-\nrange: bytes=-2",
			want: ParsedRanges{
				Present: true,
				Specs: []RangeSpec{
					{Kind: OpenRange, First: 1},
					{Kind: SuffixRange, Suffix: 2},
				},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fields, err := message.Parse(test.input)
			if err != nil {
				t.Fatalf("message.Parse(%q) returned error: %v", test.input, err)
			}
			got, err := Parse(fields)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", test.input, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Parse(%q) = %#v, want %#v", test.input, got, test.want)
			}
		})
	}
}

func TestParseLength(t *testing.T) {
	tests := []struct {
		input     string
		want      int64
		wantError string
	}{
		{input: "0", want: 0},
		{input: "1", want: 1},
		{input: "999999999999999", want: 999999999999999},
		{input: "", wantError: "invalid length \"\""},
		{input: "00", wantError: "invalid length \"00\""},
		{input: "007", wantError: "invalid length \"007\""},
		{input: "-1", wantError: "invalid length \"-1\""},
		{input: "+1", wantError: "invalid length \"+1\""},
		{input: " 1", wantError: "invalid length \" 1\""},
		{input: "1 ", wantError: "invalid length \"1 \""},
		{input: "1.0", wantError: "invalid length \"1.0\""},
		{input: "1000000000000000", wantError: "invalid length \"1000000000000000\""},
	}
	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			got, err := ParseLength(test.input)
			if test.wantError != "" {
				if err == nil {
					t.Fatalf("ParseLength(%q) = %d, want error %q", test.input, got, test.wantError)
				}
				if err.Error() != test.wantError {
					t.Fatalf("ParseLength(%q) error = %q, want %q", test.input, err.Error(), test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseLength(%q) returned error: %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("ParseLength(%q) = %d, want %d", test.input, got, test.want)
			}
		})
	}
}
