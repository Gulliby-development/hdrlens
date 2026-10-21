package summary

import (
	"testing"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

func TestExportedAPI(t *testing.T) {
	var check func([]string, base.IO) (int, error) = RunCheck
	var list func([]string, base.IO) (int, error) = RunList
	if check == nil || list == nil {
		t.Fatal("command entry point is nil")
	}
	if CheckUsage != "hdrlens check FILE" {
		t.Fatalf("CheckUsage = %q", CheckUsage)
	}
	if ListUsage != "hdrlens list FILE [--by-name]" {
		t.Fatalf("ListUsage = %q", ListUsage)
	}
}

func TestSummaryStatisticsLiteralCases(t *testing.T) {
	tests := []struct {
		name   string
		fields message.Fields
		want   Stats
	}{
		{name: "empty fields", fields: message.Fields{}, want: Stats{}},
		{
			name:   "one empty value",
			fields: message.Fields{{Name: "X", Value: ""}},
			want:   Stats{Total: 1, Distinct: 1, Repeated: 0, Longest: 0},
		},
		{
			name: "case-insensitive repeated name boundary",
			fields: message.Fields{
				{Name: "A", Value: "x"},
				{Name: "a", Value: "yy"},
				{Name: "B", Value: "zzz"},
			},
			want: Stats{Total: 3, Distinct: 2, Repeated: 1, Longest: 3},
		},
		{
			name: "three occurrences are one repeated name",
			fields: message.Fields{
				{Name: "A", Value: "one"},
				{Name: "a", Value: "two"},
				{Name: "A", Value: "longest"},
				{Name: "B", Value: "four"},
			},
			want: Stats{Total: 4, Distinct: 2, Repeated: 1, Longest: 7},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if got := Summarize(test.fields); got != test.want {
				t.Fatalf("Summarize(%#v) = %#v, want %#v", test.fields, got, test.want)
			}
		})
	}
}

func TestCommandOptionAndArgumentErrors(t *testing.T) {
	tests := []struct {
		name    string
		check   bool
		args    []string
		wantErr string
	}{
		{name: "check requires a file", check: true, args: []string{}, wantErr: "missing FILE"},
		{name: "list requires a file", args: []string{}, wantErr: "missing FILE"},
		{name: "check rejects extra arguments", check: true, args: []string{"a", "b"}, wantErr: "too many arguments"},
		{name: "list rejects extra arguments", args: []string{"a", "b"}, wantErr: "too many arguments"},
		{name: "check rejects options before argument checks", check: true, args: []string{"--by-name"}, wantErr: "unknown option \"--by-name\""},
		{name: "list rejects unknown options before reading", args: []string{"resp.txt", "--write"}, wantErr: "unknown option \"--write\""},
		{name: "list by-name does not take a value", args: []string{"resp.txt", "--by-name=1"}, wantErr: "--by-name takes no value"},
		{name: "list rejects a repeated flag", args: []string{"--by-name", "--by-name"}, wantErr: "repeated option --by-name"},
		{name: "check treats help as an unknown option", check: true, args: []string{"--help"}, wantErr: "unknown option \"--help\""},
		{name: "list treats double dash as an unknown option", args: []string{"--"}, wantErr: "unknown option \"--\""},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Files: map[string][]byte{}}
			var status int
			var err error
			if test.check {
				status, err = RunCheck(test.args, memory)
			} else {
				status, err = RunList(test.args, memory)
			}
			if status != 0 || err == nil || err.Error() != test.wantErr {
				t.Fatalf("got status %d and error %v, want status 0 and %q", status, err, test.wantErr)
			}
			if memory.Stdout != "" || memory.Stderr != "" {
				t.Fatalf("error wrote output: stdout %q, stderr %q", memory.Stdout, memory.Stderr)
			}
		})
	}
}

func TestCheckAndListSourceCases(t *testing.T) {
	tests := []struct {
		name       string
		check      bool
		args       []string
		memory     *base.MemIO
		wantStatus int
		wantErr    string
		wantOutput string
	}{
		{
			name:    "check missing file",
			check:   true,
			args:    []string{"nowhere.txt"},
			memory:  &base.MemIO{Files: map[string][]byte{}},
			wantErr: "\"nowhere.txt\": cannot read the file",
		},
		{
			name:    "list missing file",
			args:    []string{"nowhere.txt"},
			memory:  &base.MemIO{Files: map[string][]byte{}},
			wantErr: "\"nowhere.txt\": cannot read the file",
		},
		{
			name:    "check exact malformed stdin bytes",
			check:   true,
			args:    []string{"-"},
			memory:  &base.MemIO{Stdin: []byte{65, 255, 10}},
			wantErr: "<stdin>: input is not valid UTF-8",
		},
		{
			name:    "list exact malformed stdin bytes",
			args:    []string{"-"},
			memory:  &base.MemIO{Stdin: []byte{65, 255, 10}},
			wantErr: "<stdin>: input is not valid UTF-8",
		},
		{
			name:    "check empty header",
			check:   true,
			args:    []string{"empty.txt"},
			memory:  &base.MemIO{Files: map[string][]byte{"empty.txt": []byte("# nothing")}},
			wantErr: "\"empty.txt\": no header fields",
		},
		{
			name:    "list empty header",
			args:    []string{"empty.txt"},
			memory:  &base.MemIO{Files: map[string][]byte{"empty.txt": []byte("# nothing")}},
			wantErr: "\"empty.txt\": no header fields",
		},
		{
			name:       "check removes one leading BOM",
			check:      true,
			args:       []string{"-"},
			memory:     &base.MemIO{Stdin: []byte("\uFEFFX: a\n")},
			wantOutput: "ok, 1 field\ndistinct    1\nrepeated    0\nlongest     1\n",
		},
		{
			name:       "list removes one leading BOM",
			args:       []string{"-"},
			memory:     &base.MemIO{Stdin: []byte("\uFEFFX: a\n")},
			wantOutput: "LINE  NAME  VALUE\n1     X     a\n",
		},
		{
			name:       "check retains a second BOM as a parse problem",
			check:      true,
			args:       []string{"-"},
			memory:     &base.MemIO{Stdin: []byte("\uFEFF\uFEFFX: a\n")},
			wantStatus: 1,
			wantOutput: "HL001 unexpected character \"?\" at 1:1\n1 problem\n",
		},
		{
			name:    "list retains a second BOM as a parse problem",
			args:    []string{"-"},
			memory:  &base.MemIO{Stdin: []byte("\uFEFF\uFEFFX: a\n")},
			wantErr: "<stdin>: HL001 unexpected character \"?\" at 1:1",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var status int
			var err error
			if test.check {
				status, err = RunCheck(test.args, test.memory)
			} else {
				status, err = RunList(test.args, test.memory)
			}
			if status != test.wantStatus {
				t.Fatalf("status = %d, want %d", status, test.wantStatus)
			}
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err == nil || err.Error() != test.wantErr {
				t.Fatalf("error = %v, want %q", err, test.wantErr)
			}
			if test.memory.Stdout != test.wantOutput || test.memory.Stderr != "" {
				t.Fatalf("stdout/stderr = %q/%q, want %q/empty", test.memory.Stdout, test.memory.Stderr, test.wantOutput)
			}
		})
	}
}

func TestParseProblemCommandResults(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		problem string
	}{
		{name: "non-ASCII character", input: "Accept: café", problem: "HL001 unexpected character \"?\" at 1:12"},
		{name: "supplementary character", input: "a: 😀", problem: "HL001 unexpected character \"?\" at 1:4"},
		{name: "embedded carriage return", input: "a: b\rc", problem: "HL001 unexpected character \"?\" at 1:5"},
		{name: "continuation after comments and blank line", input: "# note\n\nHost: a\n  x: y", problem: "HL002 continuation line at 4:1"},
		{name: "missing colon counts trailing spaces", input: "Host  ", problem: "HL003 missing colon at 1:7"},
		{name: "space in field name", input: "Host : a", problem: "HL004 invalid field name \"Host \" at 1:1"},
		{name: "empty field name", input: ": x", problem: "HL004 invalid field name \"\" at 1:1"},
		{name: "character validation precedes continuation", input: " x: é", problem: "HL001 unexpected character \"?\" at 1:5"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			checkMemory := &base.MemIO{Files: map[string][]byte{"bad.txt": []byte(test.input)}}
			status, err := RunCheck([]string{"bad.txt"}, checkMemory)
			wantCheck := test.problem + "\n1 problem\n"
			if status != 1 || err != nil || checkMemory.Stdout != wantCheck || checkMemory.Stderr != "" {
				t.Fatalf("check got status %d, error %v, stdout %q, stderr %q; want 1, nil, %q, empty",
					status, err, checkMemory.Stdout, checkMemory.Stderr, wantCheck)
			}

			listMemory := &base.MemIO{Files: map[string][]byte{"bad.txt": []byte(test.input)}}
			status, err = RunList([]string{"bad.txt"}, listMemory)
			wantList := "\"bad.txt\": " + test.problem
			if status != 0 || err == nil || err.Error() != wantList || listMemory.Stdout != "" || listMemory.Stderr != "" {
				t.Fatalf("list got status %d, error %v, stdout %q, stderr %q; want 0, %q, empty, empty",
					status, err, listMemory.Stdout, listMemory.Stderr, wantList)
			}
		})
	}
}

func TestCheckValidStatisticsLiteralCases(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "singular field and empty value",
			input:  "X:\n",
			output: "ok, 1 field\ndistinct    1\nrepeated    0\nlongest     0\n",
		},
		{
			name:   "case-insensitive repeated names",
			input:  "A: x\na: yy\nA: zzz\nB: q\nb: r\n",
			output: "ok, 5 fields\ndistinct    2\nrepeated    2\nlongest     3\n",
		},
		{
			name:   "comments trimming and CRLF",
			input:  "# note\n\nHost: \t ab \t\r\nX: z\r",
			output: "ok, 2 fields\ndistinct    2\nrepeated    0\nlongest     2\n",
		},
		{
			name:   "request response fixture",
			input:  "# a cached response\nDate: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: public, max-age=600\nLast-Modified: Sun, 06 Nov 1994 07:00:00 GMT\nAge: 100\nSet-Cookie: sid=abc; Path=/app; Secure; HttpOnly; SameSite=Lax; Max-Age=3600\nSet-Cookie: theme=dark; Domain=example.com; Expires=Mon, 07 Nov 1994 08:00:00 GMT\nSet-Cookie: tmp=1; Domain=com\n",
			output: "ok, 7 fields\ndistinct    5\nrepeated    1\nlongest     69\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Stdin: []byte(test.input)}
			status, err := RunCheck([]string{"-"}, memory)
			if status != 0 || err != nil || memory.Stdout != test.output || memory.Stderr != "" {
				t.Fatalf("got status %d, error %v, stdout %q, stderr %q; want 0, nil, %q, empty",
					status, err, memory.Stdout, memory.Stderr, test.output)
			}
		})
	}
}

func TestListLiteralOrderingCases(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		input  string
		output string
	}{
		{
			name:   "default source order with empty value",
			args:   []string{"-"},
			input:  "b: two\nA: first\na: second\nB:\n",
			output: "LINE  NAME  VALUE\n1     b     two\n2     A     first\n3     a     second\n4     B\n",
		},
		{
			name:   "sorted case-insensitive names with numeric line ties",
			args:   []string{"-", "--by-name"},
			input:  "b: two\nA: first\na: second\nB:\n",
			output: "LINE  NAME  VALUE\n2     A     first\n3     a     second\n1     b     two\n4     B\n",
		},
		{
			name:   "request fixture sorted table",
			args:   []string{"-", "--by-name"},
			input:  "Accept: text/html, application/json;q=0.9, */*;q=0.1\nRange: bytes=0-99,500-,-50\nIf-None-Match: W/\"v1\", \"v2\"\nIf-Modified-Since: Sun, 06 Nov 1994 08:49:37 GMT\n",
			output: "LINE  NAME               VALUE\n1     Accept             text/html, application/json;q=0.9, */*;q=0.1\n4     If-Modified-Since  Sun, 06 Nov 1994 08:49:37 GMT\n3     If-None-Match      W/\"v1\", \"v2\"\n2     Range              bytes=0-99,500-,-50\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Stdin: []byte(test.input)}
			status, err := RunList(test.args, memory)
			if status != 0 || err != nil || memory.Stdout != test.output || memory.Stderr != "" {
				t.Fatalf("got status %d, error %v, stdout %q, stderr %q; want 0, nil, %q, empty",
					status, err, memory.Stdout, memory.Stderr, test.output)
			}
		})
	}
}
