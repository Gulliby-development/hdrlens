package byterange

import (
	"reflect"
	"testing"

	"github.com/Gulliby-development/hdrlens/internal/base"
)

func TestResolveLiteralSpans(t *testing.T) {
	tests := []struct {
		name   string
		parsed ParsedRanges
		length int64
		want   Resolution
	}{
		{
			name: "closed range is clipped and out of bounds ranges are omitted",
			parsed: ParsedRanges{Present: true, Specs: []RangeSpec{
				{Kind: ClosedRange, First: 8, Last: 99},
				{Kind: ClosedRange, First: 10, Last: 12},
			}},
			length: 10,
			want:   Resolution{Spans: []Span{{Start: 8, End: 9}}, Total: 2},
		},
		{
			name: "out of order touching spans merge",
			parsed: ParsedRanges{Present: true, Specs: []RangeSpec{
				{Kind: ClosedRange, First: 5, Last: 9},
				{Kind: ClosedRange, First: 0, Last: 4},
				{Kind: ClosedRange, First: 30, Last: 40},
			}},
			length: 20,
			want:   Resolution{Spans: []Span{{Start: 0, End: 9}}, Total: 10},
		},
		{
			name: "overlap merges while a gap remains",
			parsed: ParsedRanges{Present: true, Specs: []RangeSpec{
				{Kind: ClosedRange, First: 8, Last: 9},
				{Kind: ClosedRange, First: 1, Last: 4},
				{Kind: ClosedRange, First: 3, Last: 6},
			}},
			length: 12,
			want: Resolution{Spans: []Span{
				{Start: 1, End: 6},
				{Start: 8, End: 9},
			}, Total: 8},
		},
		{
			name: "zero suffix selects nothing",
			parsed: ParsedRanges{Present: true, Specs: []RangeSpec{
				{Kind: SuffixRange, Suffix: 0},
			}},
			length: 10,
			want:   Resolution{Spans: []Span{}, Total: 0},
		},
		{
			name: "oversized suffix selects the full resource",
			parsed: ParsedRanges{Present: true, Specs: []RangeSpec{
				{Kind: SuffixRange, Suffix: 20},
			}},
			length: 10,
			want:   Resolution{Spans: []Span{{Start: 0, End: 9}}, Total: 10},
		},
		{
			name:   "absent Range selects no explicit spans",
			parsed: ParsedRanges{Present: false, Specs: []RangeSpec{{Kind: OpenRange, First: 0}}},
			length: 10,
			want:   Resolution{Spans: []Span{}, Total: 0},
		},
		{
			name: "zero length selects nothing",
			parsed: ParsedRanges{Present: true, Specs: []RangeSpec{
				{Kind: ClosedRange, First: 0, Last: 9},
				{Kind: OpenRange, First: 0},
				{Kind: SuffixRange, Suffix: 10},
			}},
			length: 0,
			want:   Resolution{Spans: []Span{}, Total: 0},
		},
		{
			name: "maximum duplicate spans merge without overflowing the total",
			parsed: ParsedRanges{Present: true, Specs: []RangeSpec{
				{Kind: ClosedRange, First: 0, Last: 999999999999999},
				{Kind: ClosedRange, First: 0, Last: 999999999999999},
			}},
			length: 999999999999999,
			want:   Resolution{Spans: []Span{{Start: 0, End: 999999999999998}}, Total: 999999999999999},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := Resolve(test.parsed, test.length)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Resolve(%#v, %d) = %#v, want %#v", test.parsed, test.length, got, test.want)
			}
		})
	}
}

func TestRunLiteralOutput(t *testing.T) {
	var runSignature func([]string, base.IO) (int, error) = Run
	if runSignature == nil {
		t.Fatal("Run has a nil function value")
	}
	if Usage != "hdrlens range FILE LENGTH" {
		t.Fatalf("Usage = %q, want %q", Usage, "hdrlens range FILE LENGTH")
	}

	tests := []struct {
		name   string
		header string
		length string
		code   int
		stdout string
	}{
		{
			name:   "two partial ranges and suffix",
			header: "Range: bytes=0-99,500-,-50",
			length: "1000",
			code:   0,
			stdout: "status      206\nranges      2\nserved      600\nrange       bytes 0-99/1000\nrange       bytes 500-999/1000\n",
		},
		{
			name:   "clipping leaves two partial ranges",
			header: "Range: bytes=0-99,500-,-50",
			length: "300",
			code:   0,
			stdout: "status      206\nranges      2\nserved      150\nrange       bytes 0-99/300\nrange       bytes 250-299/300\n",
		},
		{
			name:   "unsatisfiable open range",
			header: "Range: bytes=1000-",
			length: "1000",
			code:   1,
			stdout: "status      416\nrange       bytes */1000\n",
		},
		{
			name:   "full open range",
			header: "Range: bytes=0-",
			length: "1000",
			code:   0,
			stdout: "status      200\nserved      1000\n",
		},
		{
			name:   "touching and out of bounds spans",
			header: "Range: bytes=5-9,0-4,30-40",
			length: "20",
			code:   0,
			stdout: "status      206\nranges      1\nserved      10\nrange       bytes 0-9/20\n",
		},
		{
			name:   "absent Range and zero length",
			header: "Host: a",
			length: "0",
			code:   0,
			stdout: "status      200\nserved      0\n",
		},
		{
			name:   "absent Range",
			header: "Host: a",
			length: "10",
			code:   0,
			stdout: "status      200\nserved      10\n",
		},
		{
			name:   "present Range with zero length",
			header: "Range: bytes=0-",
			length: "0",
			code:   1,
			stdout: "status      416\nrange       bytes */0\n",
		},
		{
			name:   "zero suffix",
			header: "Range: bytes=-0",
			length: "10",
			code:   1,
			stdout: "status      416\nrange       bytes */10\n",
		},
		{
			name:   "oversized suffix selects full resource",
			header: "Range: bytes=-20",
			length: "10",
			code:   0,
			stdout: "status      200\nserved      10\n",
		},
		{
			name:   "case insensitive fields and units with merged spans",
			header: "rAnGe: BYTES=,01-02,,\nRange: bytes=2-4,3-3",
			length: "10",
			code:   0,
			stdout: "status      206\nranges      1\nserved      4\nrange       bytes 1-4/10\n",
		},
		{
			name:   "clipped closed range",
			header: "Range: bytes=8-99",
			length: "10",
			code:   0,
			stdout: "status      206\nranges      1\nserved      2\nrange       bytes 8-9/10\n",
		},
		{
			name:   "duplicate maximum spans merge",
			header: "Range: bytes=0-999999999999999,0-999999999999999",
			length: "999999999999999",
			code:   0,
			stdout: "status      200\nserved      999999999999999\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			mem := &base.MemIO{Files: map[string][]byte{"a.txt": []byte(test.header)}}
			code, err := runSignature([]string{"a.txt", test.length}, mem)
			if err != nil {
				t.Fatalf("Run returned error: %v", err)
			}
			if code != test.code {
				t.Fatalf("Run exit code = %d, want %d", code, test.code)
			}
			if mem.Stdout != test.stdout {
				t.Fatalf("Run stdout = %q, want %q", mem.Stdout, test.stdout)
			}
			if mem.Stderr != "" {
				t.Fatalf("Run stderr = %q, want empty", mem.Stderr)
			}
		})
	}
}

func TestRunLiteralErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "help is an unsupported option", args: []string{"--help"}, want: "unknown option \"--help\""},
		{name: "missing FILE", args: []string{}, want: "missing FILE"},
		{name: "missing LENGTH", args: []string{"a.txt"}, want: "missing LENGTH"},
		{name: "too many positional arguments", args: []string{"a.txt", "10", "11"}, want: "too many arguments"},
		{name: "option errors precede arity", args: []string{"bad.txt", "007", "--write"}, want: "unknown option \"--write\""},
		{name: "arity follows option validation", args: []string{"bad.txt", "007", "11"}, want: "too many arguments"},
		{name: "missing input file", args: []string{"nowhere.txt", "007"}, want: "\"nowhere.txt\": cannot read the file"},
		{name: "empty header file", args: []string{"empty.txt", "007"}, want: "\"empty.txt\": no header fields"},
		{name: "Range problem precedes invalid LENGTH", args: []string{"bad.txt", "007"}, want: "\"bad.txt\": HR003 invalid range \"3\" at 2:19"},
		{name: "invalid LENGTH", args: []string{"a.txt", "007"}, want: "invalid length \"007\""},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			mem := &base.MemIO{Files: map[string][]byte{
				"a.txt":     []byte("Host: a"),
				"bad.txt":   []byte("# r\nRange: bytes=1-2, 3"),
				"empty.txt": []byte("# nothing"),
			}}
			code, err := Run(test.args, mem)
			if err == nil {
				t.Fatalf("Run(%q) returned code %d without an error", test.args, code)
			}
			if code != 0 {
				t.Fatalf("Run(%q) error exit code = %d, want 0", test.args, code)
			}
			if err.Error() != test.want {
				t.Fatalf("Run(%q) error = %q, want %q", test.args, err.Error(), test.want)
			}
			if mem.Stdout != "" || mem.Stderr != "" {
				t.Fatalf("Run(%q) wrote stdout %q or stderr %q on error", test.args, mem.Stdout, mem.Stderr)
			}
		})
	}
}

func TestRunStdinOrder(t *testing.T) {
	tests := []struct {
		name   string
		stdin  []byte
		length string
		code   int
		stdout string
		err    string
	}{
		{
			name:   "stdin range succeeds",
			stdin:  []byte("Range: bytes=0-"),
			length: "10",
			code:   0,
			stdout: "status      200\nserved      10\n",
		},
		{
			name:   "stdin Range problem precedes invalid LENGTH",
			stdin:  []byte("Range: bytes=5-2"),
			length: "007",
			code:   0,
			err:    "<stdin>: HR004 reversed range \"5-2\" at 1:14",
		},
		{
			name:   "complete header parse precedes Range validation",
			stdin:  []byte("Range: bytes=5-2\nBad"),
			length: "007",
			code:   0,
			err:    "<stdin>: HL003 missing colon at 2:4",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			mem := &base.MemIO{Stdin: test.stdin}
			code, err := Run([]string{"-", test.length}, mem)
			if code != test.code {
				t.Fatalf("Run exit code = %d, want %d", code, test.code)
			}
			if test.err != "" {
				if err == nil || err.Error() != test.err {
					t.Fatalf("Run error = %v, want %q", err, test.err)
				}
				if mem.Stdout != "" || mem.Stderr != "" {
					t.Fatalf("Run wrote stdout %q or stderr %q on error", mem.Stdout, mem.Stderr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Run returned error: %v", err)
			}
			if mem.Stdout != test.stdout {
				t.Fatalf("Run stdout = %q, want %q", mem.Stdout, test.stdout)
			}
			if mem.Stderr != "" {
				t.Fatalf("Run stderr = %q, want empty", mem.Stderr)
			}
		})
	}
}
