package freshness

import (
	"errors"
	"testing"

	"github.com/Gulliby-development/hdrlens/internal/message"
)

func TestParseControlLiteralCases(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "invalid delta", text: "Cache-Control: public, max-age=abc", want: "HC002 invalid delta-seconds \"abc\" at 1:32"},
		{name: "missing delta", text: "Cache-Control: max-age", want: "HC002 invalid delta-seconds \"\" at 1:23"},
		{name: "syntax before missing delta", text: "Cache-Control: max-age=", want: "HC001 invalid directive \"max-age=\" at 1:16"},
		{name: "duplicate across directives", text: "Cache-Control: no-store,Max-Age=5,max-age=7", want: "HC003 repeated directive \"max-age\" at 1:35"},
		{name: "quoted comma", text: "Cache-Control: private=\"a,b\", max-age=5", want: ""},
		{name: "quoted delta", text: "Cache-Control: max-age=\"5\"", want: "HC002 invalid delta-seconds \"\"5\"\" at 1:24"},
		{name: "empty quoted extension", text: "Cache-Control: x=\"\"", want: ""},
		{name: "unmatched quote", text: "Cache-Control: x=\"a", want: "HC001 invalid directive \"x=\"a\" at 1:16"},
		{name: "cross-line duplicate", text: "Cache-Control: public\nCache-Control: PUBLIC", want: "HC003 repeated directive \"PUBLIC\" at 2:16"},
		{name: "delta before duplicate", text: "Cache-Control: max-age=1,MAX-AGE=abc", want: "HC002 invalid delta-seconds \"abc\" at 1:34"},
		{name: "empty pieces", text: "Cache-Control: , ,", want: ""},
		{name: "ten digit delta", text: "Cache-Control: s-maxage=9999999999", want: ""},
		{name: "eleven digit delta", text: "Cache-Control: max-age=10000000000", want: "HC002 invalid delta-seconds \"10000000000\" at 1:24"},
		{name: "invalid date", text: "Date: yesterday", want: "HC005 invalid date \"yesterday\" at 1:7"},
		{name: "duplicate age", text: "Age: 5\nAge: 6", want: "HC004 repeated field \"Age\" at 2:6"},
		{name: "duplicate before invalid age", text: "Age: 5\nage: bad", want: "HC004 repeated field \"Age\" at 2:6"},
		{name: "file order", text: "Age: -1\nDate: yesterday", want: "HC006 invalid age \"-1\" at 1:6"},
		{name: "leading zero age", text: "Age: 0000000000", want: ""},
		{name: "eleven digit age", text: "Age: 10000000000", want: "HC006 invalid age \"10000000000\" at 1:6"},
		{name: "invalid expiry allowed", text: "Expires: soon", want: ""},
		{name: "invalid modification date", text: "Last-Modified: yesterday", want: "HC005 invalid date \"yesterday\" at 1:16"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fields, err := message.Parse(test.text)
			if err != nil {
				t.Fatalf("message.Parse() error = %v", err)
			}
			_, err = ParseControl(fields)
			if test.want == "" {
				if err != nil {
					t.Fatalf("ParseControl() error = %v, want success", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ParseControl() error = nil, want %q", test.want)
			}
			var problem *message.Problem
			if !errors.As(err, &problem) {
				t.Fatalf("ParseControl() error type = %T, want *message.Problem", err)
			}
			if got := err.Error(); got != test.want {
				t.Errorf("ParseControl() error = %q, want %q", got, test.want)
			}
		})
	}
}

func TestParseControlPresenceAndZeroValues(t *testing.T) {
	fields, err := message.Parse("Date: Sun, 06 Nov 1994 08:49:37 GMT\nAge: 0\nCache-Control: max-age=0\nExpires: soon\nLast-Modified: Sun, 06 Nov 1994 07:00:00 GMT")
	if err != nil {
		t.Fatal(err)
	}
	control, err := ParseControl(fields)
	if err != nil {
		t.Fatal(err)
	}
	if !control.HasDate || control.Date == 0 {
		t.Errorf("Date presence/value = %v/%d, want present nonzero value", control.HasDate, control.Date)
	}
	if !control.HasAge || control.Age != 0 {
		t.Errorf("Age presence/value = %v/%d, want present zero", control.HasAge, control.Age)
	}
	if !control.HasExpires || control.ExpiresValid || control.Expires != 0 {
		t.Errorf("Expires presence/validity/value = %v/%v/%d, want present invalid zero", control.HasExpires, control.ExpiresValid, control.Expires)
	}
	if !control.HasLastModified || control.LastModified == 0 {
		t.Errorf("Last-Modified presence/value = %v/%d, want present nonzero value", control.HasLastModified, control.LastModified)
	}
	directive, ok := control.Directives["max-age"]
	if !ok || !directive.HasValue || directive.Value != "0" || directive.DeltaSeconds != 0 {
		t.Errorf("max-age = %#v, present = %v, want explicit zero", directive, ok)
	}

	fields, err = message.Parse("Date: Sun, 06 Nov 1994 08:49:37 GMT")
	if err != nil {
		t.Fatal(err)
	}
	control, err = ParseControl(fields)
	if err != nil {
		t.Fatal(err)
	}
	if !control.HasDate || control.HasAge || control.HasExpires || control.HasLastModified {
		t.Errorf("optional field presence = Age:%v Expires:%v Last-Modified:%v, want all absent", control.HasAge, control.HasExpires, control.HasLastModified)
	}
	if _, ok := control.Directives["max-age"]; ok {
		t.Error("max-age is present without a Cache-Control field")
	}
}

func TestParseControlCanonicalDuplicateNames(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "Date", text: "Date: Sun, 06 Nov 1994 08:49:37 GMT\nDATE: yesterday", want: "HC004 repeated field \"Date\" at 2:7"},
		{name: "Expires", text: "Expires: soon\nexpires: soon", want: "HC004 repeated field \"Expires\" at 2:10"},
		{name: "Age", text: "Age: 1\nage: bad", want: "HC004 repeated field \"Age\" at 2:6"},
		{name: "Last-Modified", text: "Last-Modified: Sun, 06 Nov 1994 07:00:00 GMT\nlast-modified: yesterday", want: "HC004 repeated field \"Last-Modified\" at 2:16"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fields, err := message.Parse(test.text)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ParseControl(fields)
			if err == nil || err.Error() != test.want {
				t.Fatalf("ParseControl() error = %v, want %q", err, test.want)
			}
			var problem *message.Problem
			if !errors.As(err, &problem) {
				t.Fatalf("ParseControl() error type = %T, want *message.Problem", err)
			}
		})
	}
}
