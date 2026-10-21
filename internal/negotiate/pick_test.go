package negotiate

import (
	"testing"

	"github.com/Gulliby-development/hdrlens/internal/base"
)

func TestAcceptRunErrorsAndOrdering(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdin    []byte
		files    map[string][]byte
		wantCode int
		wantErr  string
	}{
		{
			name:     "help is not a command option",
			args:     []string{"--help"},
			wantCode: 0,
			wantErr:  "unknown option \"--help\"",
		},
		{
			name:     "missing file",
			args:     []string{},
			wantCode: 0,
			wantErr:  "missing FILE",
		},
		{
			name:     "missing offer",
			args:     []string{"req.txt"},
			wantCode: 0,
			wantErr:  "missing OFFER",
		},
		// AC-007 arguments: ["nowhere.txt","text/*"]
		{
			name:     "load error precedes offer validation",
			args:     []string{"nowhere.txt", "text/*"},
			wantCode: 0,
			wantErr:  "\"nowhere.txt\": cannot read the file",
		},
		{
			name:     "empty header file",
			args:     []string{"empty.txt", "a/b"},
			files:    map[string][]byte{"empty.txt": []byte("# nothing")},
			wantCode: 0,
			wantErr:  "\"empty.txt\": no header fields",
		},
		// AC-007 arguments: ["bad.txt","text/*"]
		{
			name:     "invalid accept range is wrapped once",
			args:     []string{"bad.txt", "text/*"},
			files:    map[string][]byte{"bad.txt": []byte("Accept: x")},
			wantCode: 0,
			wantErr:  "\"bad.txt\": HA001 invalid media range \"x\" at 1:9",
		},
		{
			name:     "all accept fields are validated before offer validation",
			args:     []string{"a.txt", "a/b", "text/*"},
			files:    map[string][]byte{"a.txt": []byte("Accept: x")},
			wantCode: 0,
			wantErr:  "\"a.txt\": HA001 invalid media range \"x\" at 1:9",
		},
		// AC-007 arguments: ["a.txt","a/b","text/*","-"]
		{
			name:     "invalid offer is reported after a valid header",
			args:     []string{"a.txt", "a/b", "text/*", "-"},
			files:    map[string][]byte{"a.txt": []byte("Host: a")},
			wantCode: 0,
			wantErr:  "invalid offer \"text/*\"",
		},
		{
			name:     "all header syntax is parsed before negotiation",
			args:     []string{"a.txt", "a/b"},
			files:    map[string][]byte{"a.txt": []byte("Accept: x\nBad")},
			wantCode: 0,
			wantErr:  "\"a.txt\": HL003 missing colon at 2:4",
		},
		{
			name:     "later invalid accept field is not bypassed by a match",
			args:     []string{"a.txt", "text/html"},
			files:    map[string][]byte{"a.txt": []byte("Accept: text/html\nAccept: x")},
			wantCode: 0,
			wantErr:  "\"a.txt\": HA001 invalid media range \"x\" at 2:9",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Stdin: test.stdin, Files: test.files}
			code, err := Run(test.args, memory)
			if code != test.wantCode {
				t.Fatalf("Run code = %d, want %d", code, test.wantCode)
			}
			if err == nil || err.Error() != test.wantErr {
				t.Fatalf("Run error = %v, want %q", err, test.wantErr)
			}
			if memory.Stdout != "" || memory.Stderr != "" {
				t.Fatalf("Run wrote output on error: stdout %q, stderr %q", memory.Stdout, memory.Stderr)
			}
		})
	}
}

func TestAcceptRunOutputs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdin    []byte
		files    map[string][]byte
		wantCode int
		wantOut  string
	}{
		{
			name:     "stdin selects the first best offer",
			args:     []string{"-", "application/json", "text/plain", "image/png"},
			stdin:    []byte("Accept: text/html, application/json;q=0.9, */*;q=0.1\n"),
			wantCode: 0,
			wantOut:  "OFFER             Q    RANGE\napplication/json  0.9  application/json\ntext/plain        0.1  */*\nimage/png         0.1  */*\nselected    application/json\n",
		},
		{
			name:     "no positive rating",
			args:     []string{"a.txt", "text/html", "image/png"},
			files:    map[string][]byte{"a.txt": []byte("Accept: text/html;q=0")},
			wantCode: 1,
			wantOut:  "OFFER      Q  RANGE\ntext/html  0  text/html\nimage/png  0  -\nselected    none\n",
		},
		{
			name:     "absent accept defaults to the first offer",
			args:     []string{"a.txt", "a/b", "c/d"},
			files:    map[string][]byte{"a.txt": []byte("Host: a")},
			wantCode: 0,
			wantOut:  "OFFER  Q  RANGE\na/b    1  -\nc/d    1  -\nselected    a/b\n",
		},
		{
			name:     "case is ignored and range output is normalized",
			args:     []string{"a.txt", "APPLICATION/json"},
			files:    map[string][]byte{"a.txt": []byte("aCcEpT: Application/JSON;Q=0.85")},
			wantCode: 0,
			wantOut:  "OFFER             Q     RANGE\nAPPLICATION/json  0.85  application/json\nselected    APPLICATION/json\n",
		},
		{
			name:     "specificity wins before quality",
			args:     []string{"a.txt", "text/html", "text/css"},
			files:    map[string][]byte{"a.txt": []byte("Accept: text/*;q=0.5, text/html;q=0.2")},
			wantCode: 0,
			wantOut:  "OFFER      Q    RANGE\ntext/html  0.2  text/html\ntext/css   0.5  text/*\nselected    text/css\n",
		},
		{
			name:     "equal ratings preserve offer order",
			args:     []string{"a.txt", "image/png", "text/plain"},
			files:    map[string][]byte{"a.txt": []byte("Accept: */*;q=0.5")},
			wantCode: 0,
			wantOut:  "OFFER       Q    RANGE\nimage/png   0.5  */*\ntext/plain  0.5  */*\nselected    image/png\n",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Stdin: test.stdin, Files: test.files}
			code, err := Run(test.args, memory)
			if err != nil {
				t.Fatalf("Run returned error: %v", err)
			}
			if code != test.wantCode {
				t.Fatalf("Run code = %d, want %d", code, test.wantCode)
			}
			if memory.Stdout != test.wantOut {
				t.Fatalf("Run output = %q, want %q", memory.Stdout, test.wantOut)
			}
			if memory.Stderr != "" {
				t.Fatalf("Run wrote stderr %q", memory.Stderr)
			}
		})
	}
}

func TestAcceptUsage(t *testing.T) {
	if Usage != "hdrlens accept FILE OFFER..." {
		t.Fatalf("Usage = %q, want %q", Usage, "hdrlens accept FILE OFFER...")
	}
}
