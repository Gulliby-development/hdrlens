package cli

// BEGIN acceptance rows imports: host-written from the audited spec table. Edit acceptanceActual, not this block.
import (
	acceptancejson "encoding/json"
	acceptancereflect "reflect"
	acceptancetesting "testing"
)

// END acceptance rows imports

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Gulliby-development/hdrlens/internal/base"
)

// BEGIN acceptance rows: host-written from the audited spec table. Edit acceptanceActual, not this block.
func TestAcceptanceRowsCli(t *acceptancetesting.T) {
	var rows []struct {
		ID       string `json:"id"`
		Input    any    `json:"input"`
		Expected any    `json:"expected"`
	}
	if err := acceptancejson.Unmarshal([]byte("[{\"id\":\"AC-003:1\",\"input\":{\"argv\":[\"check\",\"-\"],\"stdin\":\"X: y\\n\"},\"expected\":{\"status\":0,\"stdout\":\"ok, 1 field\\ndistinct    1\\nrepeated    0\\nlongest     1\\n\"}},{\"id\":\"AC-003:2\",\"input\":{\"argv\":[\"list\",\"-\"],\"stdin\":\"X: y\\n\"},\"expected\":{\"status\":0,\"stdout\":\"LINE  NAME  VALUE\\n1     X     y\\n\"}},{\"id\":\"AC-003:3\",\"input\":{\"argv\":[\"accept\",\"-\",\"a/b\"],\"stdin\":\"X: y\\n\"},\"expected\":{\"status\":0,\"stdout\":\"OFFER  Q  RANGE\\na/b    1  -\\nselected    a/b\\n\"}},{\"id\":\"AC-003:4\",\"input\":{\"argv\":[\"range\",\"-\",\"10\"],\"stdin\":\"Range: bytes=10-\\n\"},\"expected\":{\"status\":1,\"stdout\":\"status      416\\nrange       bytes */10\\n\"}},{\"id\":\"AC-003:5\",\"input\":{\"argv\":[\"cache\",\"-\",\"--now\",\"Sun, 06 Nov 1994 08:49:37 GMT\"],\"stdin\":\"Date: Sun, 06 Nov 1994 08:49:37 GMT\\n\"},\"expected\":{\"status\":1,\"stdout\":\"verdict     stale\\nlifetime    0 (none)\\nage         0\\nover        0\\n\"}},{\"id\":\"AC-003:6\",\"input\":{\"argv\":[\"cookie\",\"-\",\"http://a.com/\",\"--now\",\"Sun, 06 Nov 1994 08:49:37 GMT\"],\"stdin\":\"Set-Cookie: a=1\\n\"},\"expected\":{\"status\":0,\"stdout\":\"LINE  NAME  STATUS  DOMAIN  PATH  EXPIRES\\n1     a     stored  a.com   /     session\\n\"}},{\"id\":\"AC-003:7\",\"input\":{\"argv\":[\"check\",\"-\"],\"stdin\":\"Host\\n\"},\"expected\":{\"status\":1,\"stdout\":\"HL003 missing colon at 1:5\\n1 problem\\n\"}}]"), &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		row := row
		t.Run(row.ID, func(t *acceptancetesting.T) {
			got, err := acceptanceActual(row.Input)
			if err != nil {
				t.Fatal(err)
			}
			text, err := acceptancejson.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var value any
			if err := acceptancejson.Unmarshal(text, &value); err != nil {
				t.Fatal(err)
			}
			if !acceptancereflect.DeepEqual(value, row.Expected) {
				want, _ := acceptancejson.Marshal(row.Expected)
				t.Fatalf("got %s, want %s", text, want)
			}
		})
	}
}

// END acceptance rows

// acceptanceActual runs a literal acceptance row through the in-memory CLI.

func acceptanceActual(input any) (any, error) {
	object, ok := input.(map[string]any)
	if !ok {
		return nil, errors.New("acceptance input is not an object")
	}
	rawArgs, ok := object["argv"].([]any)
	if !ok {
		return nil, errors.New("acceptance argv is not an array")
	}
	args := make([]string, 0, len(rawArgs))
	for _, rawArg := range rawArgs {
		arg, ok := rawArg.(string)
		if !ok {
			return nil, errors.New("acceptance argv item is not a string")
		}
		args = append(args, arg)
	}
	stdin, ok := object["stdin"].(string)
	if !ok {
		return nil, errors.New("acceptance stdin is not a string")
	}

	memory := newCliMemIO([]byte(stdin))
	status := Run(args, memory)
	if memory.Stderr != "" {
		return nil, errors.New("acceptance command wrote unexpected stderr")
	}
	return map[string]any{"status": status, "stdout": memory.Stdout}, nil
}

func newCliMemIO(stdin []byte) *base.MemIO {
	return &base.MemIO{
		Stdin: stdin,
		Files: map[string][]byte{
			"resp.txt":  []byte("# a cached response\nDate: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: public, max-age=600\nLast-Modified: Sun, 06 Nov 1994 07:00:00 GMT\nAge: 100\nSet-Cookie: sid=abc; Path=/app; Secure; HttpOnly; SameSite=Lax; Max-Age=3600\nSet-Cookie: theme=dark; Domain=example.com; Expires=Mon, 07 Nov 1994 08:00:00 GMT\nSet-Cookie: tmp=1; Domain=com\n"),
			"req.txt":   []byte("Accept: text/html, application/json;q=0.9, */*;q=0.1\nRange: bytes=0-99,500-,-50\nIf-None-Match: W/\"v1\", \"v2\"\nIf-Modified-Since: Sun, 06 Nov 1994 08:49:37 GMT\n"),
			"empty.txt": []byte("# nothing\n"),
			"bad.txt":   []byte("# r\nRange: bytes=1-2, 3\n"),
			"a.txt":     []byte("Age: 1\n"),
		},
	}
}

func TestUsageAndHelp(t *testing.T) {
	var usage string = Usage
	var run func([]string, base.IO) int = Run
	if run == nil {
		t.Fatal("Run is nil")
	}
	want := "usage: hdrlens check FILE\n" +
		"       hdrlens list FILE [--by-name]\n" +
		"       hdrlens accept FILE OFFER...\n" +
		"       hdrlens range FILE LENGTH\n" +
		"       hdrlens cache FILE --now DATE [--shared]\n" +
		"       hdrlens cookie FILE URL --now DATE [--send URL]\n"
	if usage != want {
		t.Fatalf("Usage = %q, want %q", usage, want)
	}

	memory := newCliMemIO(nil)
	if status := Run([]string{"--help"}, memory); status != 0 {
		t.Fatalf("help status = %d, want 0", status)
	}
	if memory.Stdout != want || memory.Stderr != "" {
		t.Fatalf("help stdout/stderr = %q/%q, want exact usage and empty stderr", memory.Stdout, memory.Stderr)
	}
}

func TestWorkedCliErrors(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		stdin []byte
		want  string
	}{
		{name: "no command", args: []string{}, want: "hdrlens: missing command (check, list, accept, range, cache or cookie)\n"},
		{name: "unknown command", args: []string{"fold", "x"}, want: "hdrlens: unknown command \"fold\"\n"},
		{name: "help with arguments is not help", args: []string{"--help", "x"}, want: "hdrlens: unknown command \"--help\"\n"},
		{name: "unknown option", args: []string{"list", "resp.txt", "--write"}, want: "hdrlens: unknown option \"--write\"\n"},
		{name: "flag rejects a value", args: []string{"list", "resp.txt", "--by-name=1"}, want: "hdrlens: --by-name takes no value\n"},
		{name: "option value is missing", args: []string{"cache", "resp.txt", "--now"}, want: "hdrlens: --now needs a value\n"},
		{name: "option cannot repeat", args: []string{"cache", "resp.txt", "--now", "A", "--now", "B"}, want: "hdrlens: repeated option --now\n"},
		{name: "required option is missing", args: []string{"cache", "resp.txt"}, want: "hdrlens: missing --now\n"},
		{name: "required offer is missing", args: []string{"accept", "req.txt"}, want: "hdrlens: missing OFFER\n"},
		{name: "extra range argument", args: []string{"range", "req.txt", "10", "11"}, want: "hdrlens: too many arguments\n"},
		{name: "unreadable file", args: []string{"check", "nowhere.txt"}, want: "hdrlens: \"nowhere.txt\": cannot read the file\n"},
		{name: "malformed stdin bytes through memory IO", args: []string{"check", "-"}, stdin: []byte{0x41, 0xFF, 0x0A}, want: "hdrlens: <stdin>: input is not valid UTF-8\n"},
		{name: "header file has no fields", args: []string{"list", "empty.txt"}, want: "hdrlens: \"empty.txt\": no header fields\n"},
		{name: "header problem precedes length validation", args: []string{"range", "bad.txt", "10"}, want: "hdrlens: \"bad.txt\": HR003 invalid range \"3\" at 2:19\n"},
		{name: `["accept","req.txt","text/*"] (invalid offer)`, args: []string{"accept", "req.txt", "text/*"}, want: "hdrlens: invalid offer \"text/*\"\n"},
		{name: "invalid length", args: []string{"range", "req.txt", "007"}, want: "hdrlens: invalid length \"007\"\n"},
		{name: "missing date field", args: []string{"cache", "a.txt", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"}, want: "hdrlens: \"a.txt\": missing Date field\n"},
		{name: "invalid now date", args: []string{"cache", "resp.txt", "--now", "yesterday"}, want: "hdrlens: invalid date \"yesterday\"\n"},
		{name: "missing cookie URL", args: []string{"cookie", "resp.txt", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"}, want: "hdrlens: missing URL\n"},
		{name: "invalid cookie URL", args: []string{"cookie", "resp.txt", "ftp://a.com/", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"}, want: "hdrlens: invalid URL \"ftp://a.com/\"\n"},
		{name: "missing Set-Cookie field", args: []string{"cookie", "req.txt", "http://a.com/", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"}, want: "hdrlens: \"req.txt\": no Set-Cookie field\n"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assertCliError(t, test.args, test.stdin, test.want)
		})
	}
}

func TestCliPrecedenceErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "option error before required option", args: []string{"cache", "--write"}, want: "hdrlens: unknown option \"--write\"\n"},
		{name: "required option before ordinary arguments", args: []string{"cache"}, want: "hdrlens: missing --now\n"},
		{name: "too many arguments before file read", args: []string{"range", "nowhere.txt", "10", "11"}, want: "hdrlens: too many arguments\n"},
		{name: "file read before length validation", args: []string{"range", "nowhere.txt", "007"}, want: "hdrlens: \"nowhere.txt\": cannot read the file\n"},
		{name: "file problem before length validation", args: []string{"range", "bad.txt", "007"}, want: "hdrlens: \"bad.txt\": HR003 invalid range \"3\" at 2:19\n"},
		{name: "file read before option value validation", args: []string{"cache", "nowhere.txt", "--now", "yesterday"}, want: "hdrlens: \"nowhere.txt\": cannot read the file\n"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assertCliError(t, test.args, nil, test.want)
		})
	}
}

func TestUnknownCommandTokenEdges(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "empty command token", args: []string{""}, want: "hdrlens: unknown command \"\"\n"},
		{name: "non-ASCII command token", args: []string{"😀"}, want: "hdrlens: unknown command \"?\"\n"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assertCliError(t, test.args, nil, test.want)
		})
	}
}

func assertCliError(t *testing.T, args []string, stdin []byte, want string) {
	t.Helper()
	memory := newCliMemIO(stdin)
	if status := Run(args, memory); status != 2 {
		t.Fatalf("Run(%q) status = %d, want 2", args, status)
	}
	if memory.Stdout != "" || memory.Stderr != want {
		t.Fatalf("Run(%q) stdout/stderr = %q/%q, want empty/%q", args, memory.Stdout, memory.Stderr, want)
	}
}

func TestMalformedStdinProcess(t *testing.T) {
	tempDir := t.TempDir()
	binaryName := "hdrlens"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(tempDir, binaryName)

	buildContext, cancelBuild := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelBuild()
	build := exec.CommandContext(buildContext, "go", "build", "-o", binaryPath, "../..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	runContext, cancelRun := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelRun()
	command := exec.CommandContext(runContext, binaryPath, "check", "-")
	command.Stdin = bytes.NewReader([]byte{0x41, 0xFF, 0x0A})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("process error = %v, want exit status 2", err)
	}
	if exitErr.ExitCode() != 2 {
		t.Fatalf("process exit code = %d, want 2", exitErr.ExitCode())
	}
	if stdout.String() != "" {
		t.Fatalf("process stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "hdrlens: <stdin>: input is not valid UTF-8\n" {
		t.Fatalf("process stderr = %q, want exact malformed-stdin diagnostic", stderr.String())
	}
}
