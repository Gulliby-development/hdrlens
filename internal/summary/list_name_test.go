package summary

import (
	"testing"

	"github.com/Gulliby-development/hdrlens/internal/base"
)

const nameInput = "Set-Cookie: a=1\nAccept: text/html\nset-cookie: b=2\nHost: example.com\nSET-COOKIE: c=3\n"

func TestListNameFilter(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		output string
	}{
		{
			name:   "keeps the fields of one name in file order ignoring case",
			args:   []string{"-", "--name", "Set-Cookie"},
			output: "LINE  NAME        VALUE\n1     Set-Cookie  a=1\n3     set-cookie  b=2\n5     SET-COOKIE  c=3\n",
		},
		{
			name:   "matches a lower-case argument against mixed-case fields",
			args:   []string{"-", "--name=set-cookie"},
			output: "LINE  NAME        VALUE\n1     Set-Cookie  a=1\n3     set-cookie  b=2\n5     SET-COOKIE  c=3\n",
		},
		{
			name:   "combines with --by-name",
			args:   []string{"--by-name", "-", "--name", "HOST"},
			output: "LINE  NAME  VALUE\n4     Host  example.com\n",
		},
		{
			name:   "prints only the header when no field matches",
			args:   []string{"-", "--name", "Cookie"},
			output: "LINE  NAME  VALUE\n",
		},
		{
			name:   "does not match a name that only contains the argument",
			args:   []string{"-", "--name", "Set"},
			output: "LINE  NAME  VALUE\n",
		},
		{
			name:   "keeps the unfiltered table without the option",
			args:   []string{"-"},
			output: "LINE  NAME        VALUE\n1     Set-Cookie  a=1\n2     Accept      text/html\n3     set-cookie  b=2\n4     Host        example.com\n5     SET-COOKIE  c=3\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Stdin: []byte(nameInput)}
			status, err := RunList(test.args, memory)
			if status != 0 || err != nil || memory.Stdout != test.output || memory.Stderr != "" {
				t.Fatalf("got status %d, error %v, stdout %q, stderr %q; want 0, nil, %q, empty",
					status, err, memory.Stdout, memory.Stderr, test.output)
			}
		})
	}
}

func TestListNameErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "empty value", args: []string{"-", "--name="}, wantErr: "invalid --name value \"\""},
		{name: "empty separate value", args: []string{"-", "--name", ""}, wantErr: "invalid --name value \"\""},
		{name: "missing value", args: []string{"-", "--name"}, wantErr: "--name needs a value"},
		{name: "repeated option", args: []string{"-", "--name", "a", "--name", "b"}, wantErr: "repeated option --name"},
		{name: "missing file is reported first", args: []string{"--name", ""}, wantErr: "missing FILE"},
		{name: "too many arguments is reported first", args: []string{"-", "x", "--name", ""}, wantErr: "too many arguments"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Stdin: []byte(nameInput)}
			status, err := RunList(test.args, memory)
			if err == nil || err.Error() != test.wantErr || status != 0 || memory.Stdout != "" {
				t.Fatalf("got status %d, error %v, stdout %q; want error %q", status, err, memory.Stdout, test.wantErr)
			}
		})
	}
}

func TestCheckRejectsTheNameOption(t *testing.T) {
	memory := &base.MemIO{Stdin: []byte(nameInput)}
	_, err := RunCheck([]string{"-", "--name", "Host"}, memory)
	if err == nil || err.Error() != "unknown option \"--name\"" {
		t.Fatalf("RunCheck error = %v, want unknown option \"--name\"", err)
	}
}
