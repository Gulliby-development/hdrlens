package message

import "testing"

func TestParseDateAcceptsTheObsoleteForms(t *testing.T) {
	tests := []struct {
		text string
		want int64
	}{
		{"Sun, 06 Nov 1994 08:49:37 GMT", 784111777},
		{"Sunday, 06-Nov-94 08:49:37 GMT", 784111777},
		{"Sun Nov  6 08:49:37 1994", 784111777},
		{"Thursday, 01-Jan-70 00:00:00 GMT", 0},
		{"Tuesday, 01-Jan-69 00:00:00 GMT", 3124224000},
		{"Tuesday, 29-Feb-00 00:00:00 GMT", 951782400},
		{"Fri Dec 31 23:59:59 1999", 946684799},
		{"Sun Jul 31 12:00:00 2016", 1469966400},
	}

	for _, test := range tests {
		got, ok := ParseDate(test.text)
		if !ok || got != test.want {
			t.Errorf("ParseDate(%q) = (%d, %t), want (%d, true)", test.text, got, ok, test.want)
		}
	}
}

func TestParseDateRejectsMalformedObsoleteForms(t *testing.T) {
	tests := []struct {
		text string
	}{
		{"Sun, 06-Nov-94 08:49:37 GMT"},
		{"Sunday, 06-Nov-1994 08:49:37 GMT"},
		{"Sunday, 6-Nov-94 08:49:37 GMT"},
		{"Sun Nov 06 08:49:37 1994"},
		{"Sun Nov 6 08:49:37 1994"},
		{"Sun Nov  6 08:49:37 1994 GMT"},
		{"Sunday, 06-nov-94 08:49:37 GMT"},
		{"Monday, 06-Nov-94 08:49:37 GMT"},
		{"Mon Nov  6 08:49:37 1994"},
		{"Sunday, 06-Nov-94 08:49:37 UTC"},
		{"Sunday, 06-Nov-94 08:49:37 GMT "},
		{"Sunday, 06 Nov 1994 08:49:37 GMT"},
		{"Sun Nov  0 08:49:37 1994"},
		{"Mon Feb 30 08:49:37 1994"},
		{"Wed Dec 31 23:59:59 1969"},
	}

	for _, test := range tests {
		if got, ok := ParseDate(test.text); ok {
			t.Errorf("ParseDate(%q) = %d, want rejection", test.text, got)
		}
	}
}

func TestFormatDateKeepsTheFixdateForm(t *testing.T) {
	const want = "Sun, 06 Nov 1994 08:49:37 GMT"
	if got := FormatDate(784111777); got != want {
		t.Errorf("FormatDate(784111777) = %q, want %q", got, want)
	}
}
