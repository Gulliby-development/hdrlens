package freshness

import (
	"testing"

	"github.com/Gulliby-development/hdrlens/internal/base"
)

func TestRunValidationLiteralCases(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		headers string
		wantOut string
		wantErr string
		status  int
	}{
		{name: "missing now value", args: []string{"a.txt", "--now"}, headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT", wantErr: "--now needs a value"},
		{name: "repeated now", args: []string{"a.txt", "--now", "A", "--now", "B"}, headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT", wantErr: "repeated option --now"},
		{name: "missing now with file", args: []string{"a.txt"}, headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT", wantErr: "missing --now"},
		{name: "missing now without file", args: []string{}, wantErr: "missing --now"},
		{name: "missing file before invalid now", args: []string{"--now", "yesterday"}, wantErr: "missing FILE"},
		{name: "too many files before invalid now", args: []string{"a.txt", "b.txt", "--now", "yesterday"}, headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT", wantErr: "too many arguments"},
		{name: "unknown option first", args: []string{"--bad"}, wantErr: "unknown option \"--bad\""},
		{name: "invalid now after file validation", args: []string{"a.txt", "--now", "yesterday"}, headers: "Age: 1", wantErr: "invalid date \"yesterday\""},
		{name: "empty now is present but invalid", args: []string{"a.txt", "--now", ""}, headers: "Age: 1", wantErr: "invalid date \"\""},
		{name: "missing Date follows valid now", args: []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"}, headers: "Age: 1", wantErr: "\"a.txt\": missing Date field"},
		{name: "file problem precedes invalid now", args: []string{"a.txt", "--now", "yesterday"}, headers: "Age: bad", wantErr: "\"a.txt\": HC006 invalid age \"bad\" at 1:6"},
		{name: "missing file", args: []string{"missing.txt", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"}, wantErr: "\"missing.txt\": cannot read the file"},
		{name: "no header fields", args: []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"}, headers: "# nothing", wantErr: "\"a.txt\": no header fields"},
		{name: "positioned problem is wrapped once", args: []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"}, headers: " x: y", wantErr: "\"a.txt\": HL002 continuation line at 1:1"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Files: map[string][]byte{"a.txt": []byte(test.headers)}}
			status, err := Run(test.args, memory)
			if status != test.status {
				t.Errorf("Run() status = %d, want %d", status, test.status)
			}
			if err == nil || err.Error() != test.wantErr {
				t.Errorf("Run() error = %v, want %q", err, test.wantErr)
			}
			if memory.Stdout != test.wantOut {
				t.Errorf("Run() stdout = %q, want %q", memory.Stdout, test.wantOut)
			}
			if memory.Stderr != "" {
				t.Errorf("Run() stderr = %q, want empty", memory.Stderr)
			}
		})
	}
}

func TestRunFreshnessOutputLiteralCases(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		headers string
		wantOut string
		status  int
	}{
		{
			name:    "max age and stored age are fresh",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: public, max-age=600\nAge: 100",
			wantOut: "verdict     fresh\nlifetime    600 (max-age)\nage         280\nleft        320\n",
			status:  0,
		},
		{
			name:    "age boundary is stale",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: max-age=180",
			wantOut: "verdict     stale\nlifetime    180 (max-age)\nage         180\nover        0\n",
			status:  1,
		},
		{
			name:    "one second above age boundary is fresh",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: max-age=181",
			wantOut: "verdict     fresh\nlifetime    181 (max-age)\nage         180\nleft        1\n",
			status:  0,
		},
		{
			name:    "stale max age",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 09:00:00 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: public, max-age=600\nAge: 100",
			wantOut: "verdict     stale\nlifetime    600 (max-age)\nage         723\nover        123\n",
			status:  1,
		},
		{
			name:    "shared lifetime takes s maxage",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT", "--shared"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: s-maxage=60, max-age=600",
			wantOut: "verdict     stale\nlifetime    60 (s-maxage)\nage         180\nover        120\n",
			status:  1,
		},
		{
			name:    "private lifetime ignores s maxage",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: s-maxage=60, max-age=600",
			wantOut: "verdict     fresh\nlifetime    600 (max-age)\nage         180\nleft        420\n",
			status:  0,
		},
		{
			name:    "heuristic lifetime",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nLast-Modified: Sun, 06 Nov 1994 07:00:00 GMT",
			wantOut: "verdict     fresh\nlifetime    657 (heuristic)\nage         180\nleft        477\n",
			status:  0,
		},
		{
			name:    "expires lifetime",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nExpires: Sun, 06 Nov 1994 09:00:00 GMT",
			wantOut: "verdict     fresh\nlifetime    623 (expires)\nage         180\nleft        443\n",
			status:  0,
		},
		{
			name:    "invalid expiry is zero lifetime",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nExpires: soon",
			wantOut: "verdict     stale\nlifetime    0 (expires)\nage         180\nover        180\n",
			status:  1,
		},
		{
			name:    "no lifetime source",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT",
			wantOut: "verdict     stale\nlifetime    0 (none)\nage         180\nover        180\n",
			status:  1,
		},
		{
			name:    "past expiry is zero",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nExpires: Sun, 06 Nov 1994 07:00:00 GMT",
			wantOut: "verdict     stale\nlifetime    0 (expires)\nage         180\nover        180\n",
			status:  1,
		},
		{
			name:    "future Last Modified clamps heuristic to zero",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nLast-Modified: Sun, 06 Nov 1994 09:00:00 GMT",
			wantOut: "verdict     stale\nlifetime    0 (heuristic)\nage         180\nover        180\n",
			status:  1,
		},
		{
			name:    "present invalid expires outranks heuristic",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nExpires: soon\nLast-Modified: Sun, 06 Nov 1994 07:00:00 GMT",
			wantOut: "verdict     stale\nlifetime    0 (expires)\nage         180\nover        180\n",
			status:  1,
		},
		{
			name:    "max age outranks invalid expires",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: max-age=600\nExpires: soon",
			wantOut: "verdict     fresh\nlifetime    600 (max-age)\nage         180\nleft        420\n",
			status:  0,
		},
		{
			name:    "private cache ignores shared max age",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: s-maxage=60\nExpires: soon",
			wantOut: "verdict     stale\nlifetime    0 (expires)\nage         180\nover        180\n",
			status:  1,
		},
		{
			name:    "future Date does not reduce stored age",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 07:00:00 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nAge: 100",
			wantOut: "verdict     stale\nlifetime    0 (none)\nage         100\nover        100\n",
			status:  1,
		},
		{
			name:    "maximum accepted age and delta do not overflow",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nAge: 9999999999\nCache-Control: max-age=9999999999",
			wantOut: "verdict     stale\nlifetime    9999999999 (max-age)\nage         9999999999\nover        0\n",
			status:  1,
		},
		{
			name:    "no cache forces revalidation without difference row",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: no-cache\nExpires: soon",
			wantOut: "verdict     revalidate\nlifetime    0 (expires)\nage         180\n",
			status:  1,
		},
		{
			name:    "no store policy",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: no-store, max-age=600",
			wantOut: "verdict     uncacheable\nlifetime    600 (max-age)\nage         180\n",
			status:  1,
		},
		{
			name:    "private directive is a shared cache policy",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT", "--shared"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: private=\"a,b\", max-age=600",
			wantOut: "verdict     uncacheable\nlifetime    600 (max-age)\nage         180\n",
			status:  1,
		},
		{
			name:    "valued private directive is fresh privately",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: private=\"a,b\", max-age=600",
			wantOut: "verdict     fresh\nlifetime    600 (max-age)\nage         180\nleft        420\n",
			status:  0,
		},
		{
			name:    "no store outranks no cache",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: no-store, no-cache, max-age=600",
			wantOut: "verdict     uncacheable\nlifetime    600 (max-age)\nage         180\n",
			status:  1,
		},
		{
			name:    "shared private outranks no cache",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT", "--shared"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: private, no-cache, max-age=600",
			wantOut: "verdict     uncacheable\nlifetime    600 (max-age)\nage         180\n",
			status:  1,
		},
		{
			name:    "no cache outranks freshness",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: no-cache, max-age=600",
			wantOut: "verdict     revalidate\nlifetime    600 (max-age)\nage         180\n",
			status:  1,
		},
		{
			name:    "valued no store policy counts",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: no-store=yes, max-age=600",
			wantOut: "verdict     uncacheable\nlifetime    600 (max-age)\nage         180\n",
			status:  1,
		},
		{
			name:    "valued no cache policy counts",
			args:    []string{"a.txt", "--now", "Sun, 06 Nov 1994 08:52:37 GMT"},
			headers: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nCache-Control: no-cache=\"a\", max-age=600",
			wantOut: "verdict     revalidate\nlifetime    600 (max-age)\nage         180\n",
			status:  1,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Files: map[string][]byte{"a.txt": []byte(test.headers)}}
			status, err := Run(test.args, memory)
			if err != nil {
				t.Fatalf("Run() error = %v, want nil", err)
			}
			if status != test.status {
				t.Errorf("Run() status = %d, want %d", status, test.status)
			}
			if memory.Stdout != test.wantOut {
				t.Errorf("Run() stdout = %q, want %q", memory.Stdout, test.wantOut)
			}
			if memory.Stderr != "" {
				t.Errorf("Run() stderr = %q, want empty", memory.Stderr)
			}
		})
	}
}

func TestEvaluateArithmeticLiteralCases(t *testing.T) {
	tests := []struct {
		name    string
		control Control
		now     int64
		shared  bool
		want    Evaluation
	}{
		{
			name: "exact age boundary is stale",
			control: Control{
				Date:       100,
				Directives: map[string]Directive{"max-age": {DeltaSeconds: 180}},
			},
			now:  280,
			want: Evaluation{Lifetime: 180, Source: "max-age", Age: 180, Verdict: "stale"},
		},
		{
			name: "one second more lifetime is fresh",
			control: Control{
				Date:       100,
				Directives: map[string]Directive{"max-age": {DeltaSeconds: 181}},
			},
			now:  280,
			want: Evaluation{Lifetime: 181, Source: "max-age", Age: 180, Verdict: "fresh"},
		},
		{
			name:    "future Date retains stored age",
			control: Control{Date: 300, Age: 100, Directives: map[string]Directive{"max-age": {DeltaSeconds: 101}}},
			now:     200,
			want:    Evaluation{Lifetime: 101, Source: "max-age", Age: 100, Verdict: "fresh"},
		},
		{
			name: "difference and age addition saturate at MaxInt64",
			control: Control{
				Date: -1 << 63,
				Age:  1<<63 - 1,
				Directives: map[string]Directive{
					"max-age": {DeltaSeconds: 1<<63 - 1},
				},
			},
			now: 1<<63 - 1,
			want: Evaluation{
				Lifetime: 1<<63 - 1,
				Source:   "max-age",
				Age:      1<<63 - 1,
				Verdict:  "stale",
			},
		},
		{
			name: "expiry difference saturates at MaxInt64",
			control: Control{
				Date:         -1 << 63,
				Expires:      1<<63 - 1,
				HasExpires:   true,
				ExpiresValid: true,
			},
			now: -1 << 63,
			want: Evaluation{
				Lifetime: 1<<63 - 1,
				Source:   "expires",
				Age:      0,
				Verdict:  "fresh",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := Evaluate(test.control, test.now, test.shared)
			if got != test.want {
				t.Errorf("Evaluate() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestUsage(t *testing.T) {
	if Usage != "hdrlens cache FILE --now DATE [--shared]" {
		t.Errorf("Usage = %q, want %q", Usage, "hdrlens cache FILE --now DATE [--shared]")
	}
}
