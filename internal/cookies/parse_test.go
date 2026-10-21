package cookies

import (
	"errors"
	"testing"

	"github.com/Gulliby-development/hdrlens/internal/message"
)

func parseCookieHeader(text string) (Cookie, error) {
	fields, err := message.Parse(text)
	if err != nil {
		return Cookie{}, err
	}
	cookies, err := ParseCookies(fields)
	if err != nil {
		return Cookie{}, err
	}
	if len(cookies) != 1 {
		return Cookie{}, errors.New("expected one Set-Cookie field")
	}
	return cookies[0], nil
}

func assertCookieProblem(t *testing.T, err error, code, text string, line, col int, want string) {
	t.Helper()
	if err == nil || err.Error() != want {
		t.Fatalf("error = %v, want %q", err, want)
	}
	problem, ok := err.(*message.Problem)
	if !ok {
		t.Fatalf("error type = %T, want *message.Problem", err)
	}
	if problem.Code != code || problem.Text != text || problem.Line != line || problem.Col != col {
		t.Fatalf("problem = %#v, want code %q text %q at %d:%d", problem, code, text, line, col)
	}
}

func TestParseCookieLiteralDiagnostics(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
		code string
		body string
		line int
		col  int
	}{
		{
			name: "missing equals",
			text: "Set-Cookie: sid",
			want: "HK001 missing \"=\" at 1:13",
			code: "HK001",
			body: "missing \"=\"",
			line: 1,
			col:  13,
		},
		{
			name: "invalid name",
			text: "Set-Cookie: a b=1",
			want: "HK002 invalid cookie name \"a b\" at 1:13",
			code: "HK002",
			body: "invalid cookie name \"a b\"",
			line: 1,
			col:  13,
		},
		{
			name: "invalid value",
			text: "Set-Cookie: sid=a b",
			want: "HK003 invalid cookie value \"a b\" at 1:17",
			code: "HK003",
			body: "invalid cookie value \"a b\"",
			line: 1,
			col:  17,
		},
		{
			name: "empty name",
			text: "Set-Cookie: =x",
			want: "HK002 invalid cookie name \"\" at 1:13",
			code: "HK002",
			body: "invalid cookie name \"\"",
			line: 1,
			col:  13,
		},
		{
			name: "quoted value characters are prohibited",
			text: "Set-Cookie: a=\"b\"",
			want: "HK003 invalid cookie value \"\"b\"\" at 1:15",
			code: "HK003",
			body: "invalid cookie value \"\"b\"\"",
			line: 1,
			col:  15,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			_, err := parseCookieHeader(test.text)
			assertCookieProblem(t, err, test.code, test.body, test.line, test.col, test.want)
		})
	}
}

func TestParseCookieLiteralPairs(t *testing.T) {
	tests := []struct {
		name string
		text string
		want Cookie
	}{
		{
			name: "empty value",
			text: "Set-Cookie: a=",
			want: Cookie{Line: 1, Name: "a", Value: ""},
		},
		{
			name: "first equals separates pair",
			text: "Set-Cookie: a=b=c",
			want: Cookie{Line: 1, Name: "a", Value: "b=c"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := parseCookieHeader(test.text)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("cookie = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseCookieLiteralAttributes(t *testing.T) {
	tests := []struct {
		name string
		text string
		want Cookie
	}{
		{
			name: "invalid last max age does not erase valid value",
			text: "Set-Cookie: a=1; Max-Age=4; MAX-AGE=bad",
			want: Cookie{Line: 1, Name: "a", Value: "1", MaxAge: 4, HasMaxAge: true},
		},
		{
			name: "last valid max age wins",
			text: "Set-Cookie: a=1; Max-Age=4; max-age=-0002",
			want: Cookie{Line: 1, Name: "a", Value: "1", MaxAge: -2, HasMaxAge: true},
		},
		{
			name: "invalid numeric forms are ignored",
			text: "Set-Cookie: a=1; Max-Age=+2; Max-Age=10000000000",
			want: Cookie{Line: 1, Name: "a", Value: "1"},
		},
		{
			name: "invalid last domain does not erase valid domain",
			text: "Set-Cookie: a=1; Domain=.EXAMPLE.COM; Domain=.",
			want: Cookie{Line: 1, Name: "a", Value: "1", Domain: "example.com", HasDomain: true},
		},
		{
			name: "invalid path leaves prior path and secure counts with a value",
			text: "Set-Cookie: a=1; Path=/a; Path=x; Secure=no",
			want: Cookie{Line: 1, Name: "a", Value: "1", Path: "/a", HasPath: true, Secure: true},
		},
		{
			name: "valid expires survives invalid expires and ignores other attributes",
			text: "Set-Cookie: a=1; Expires=Sun, 06 Nov 1994 08:49:37 GMT; Expires=soon; HttpOnly; SameSite=Lax;;",
			want: Cookie{Line: 1, Name: "a", Value: "1", Expires: 784111777, HasExpires: true},
		},
		{
			name: "trimmed attributes and one leading domain dot",
			text: "Set-Cookie: a=1; Domain=..EXAMPLE.COM; Path = /b; Max-Age = 0002",
			want: Cookie{Line: 1, Name: "a", Value: "1", MaxAge: 2, HasMaxAge: true, Domain: ".example.com", HasDomain: true, Path: "/b", HasPath: true},
		},
		{
			name: "one digit max age is accepted",
			text: "Set-Cookie: a=1; Max-Age=0",
			want: Cookie{Line: 1, Name: "a", Value: "1", MaxAge: 0, HasMaxAge: true},
		},
		{
			name: "ten digits max age are accepted",
			text: "Set-Cookie: a=1; Max-Age=9999999999",
			want: Cookie{Line: 1, Name: "a", Value: "1", MaxAge: 9999999999, HasMaxAge: true},
		},
		{
			name: "invalid max age does not replace ten digit value",
			text: "Set-Cookie: a=1; Max-Age=9999999999; Max-Age=10000000000",
			want: Cookie{Line: 1, Name: "a", Value: "1", MaxAge: 9999999999, HasMaxAge: true},
		},
		{
			name: "empty expires does not set an expiry",
			text: "Set-Cookie: a=1; Expires=",
			want: Cookie{Line: 1, Name: "a", Value: "1"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := parseCookieHeader(test.text)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("cookie = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseURLLiteralCases(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		want        URL
		defaultPath string
		wantErr     string
	}{
		{
			name:        "host only has slash path",
			text:        "http://A.com",
			want:        URL{Scheme: "http", Host: "a.com", Path: "/"},
			defaultPath: "/",
		},
		{
			name:        "port and query fragment",
			text:        "https://a.com:00000/app/home?x#y",
			want:        URL{Scheme: "https", Host: "a.com", Path: "/app/home"},
			defaultPath: "/app",
		},
		{
			name:        "port has no numeric range limit",
			text:        "http://a.com:99999/a",
			want:        URL{Scheme: "http", Host: "a.com", Path: "/a"},
			defaultPath: "/",
		},
		{
			name:        "query without path",
			text:        "http://a.com?x",
			want:        URL{Scheme: "http", Host: "a.com", Path: "/"},
			defaultPath: "/",
		},
		{
			name:    "unsupported scheme",
			text:    "ftp://a.com/",
			wantErr: "invalid URL \"ftp://a.com/\"",
		},
		{
			name:    "empty host label",
			text:    "http://a..com/",
			wantErr: "invalid URL \"http://a..com/\"",
		},
		{
			name:    "host label starts with hyphen",
			text:    "http://-a.com/",
			wantErr: "invalid URL \"http://-a.com/\"",
		},
		{
			name:    "port exceeds five digits",
			text:    "http://a.com:123456/",
			wantErr: "invalid URL \"http://a.com:123456/\"",
		},
		{
			name:    "empty port",
			text:    "http://a.com:/",
			wantErr: "invalid URL \"http://a.com:/\"",
		},
		{
			name:    "credentials are not allowed",
			text:    "http://u@a.com/",
			wantErr: "invalid URL \"http://u@a.com/\"",
		},
		{
			name:        "year-independent raw non-ascii path",
			text:        "https://a.com/é/x",
			want:        URL{Scheme: "https", Host: "a.com", Path: "/é/x"},
			defaultPath: "/é",
		},
		{
			name:        "raw tab path is preserved",
			text:        "https://a.com/\t/x",
			want:        URL{Scheme: "https", Host: "a.com", Path: "/\t/x"},
			defaultPath: "/\t",
		},
		{
			name:        "one digit port",
			text:        "https://a.com:1/",
			want:        URL{Scheme: "https", Host: "a.com", Path: "/"},
			defaultPath: "/",
		},
		{
			name:        "fragment without path",
			text:        "http://a.com#fragment",
			want:        URL{Scheme: "http", Host: "a.com", Path: "/"},
			defaultPath: "/",
		},
		{
			name:    "scheme is exact lowercase",
			text:    "HTTPS://a.com/",
			wantErr: "invalid URL \"HTTPS://a.com/\"",
		},
		{
			name:    "host label ends with hyphen",
			text:    "http://a-.com/",
			wantErr: "invalid URL \"http://a-.com/\"",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseURL(test.text)
			if test.wantErr != "" {
				if err == nil || err.Error() != test.wantErr {
					t.Fatalf("ParseURL(%q) error = %v, want %q", test.text, err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("ParseURL(%q) = %#v, want %#v", test.text, got, test.want)
			}
			if gotPath := DefaultPath(got.Path); gotPath != test.defaultPath {
				t.Fatalf("DefaultPath(%q) = %q, want %q", got.Path, gotPath, test.defaultPath)
			}
		})
	}
}

func TestStoreCookieLiteralCases(t *testing.T) {
	tests := []struct {
		name       string
		cookieText string
		response   string
		date       string
		status     string
		reason     string
		domain     string
		path       string
		expires    string
	}{
		{
			name:       "single label domain rejected first",
			cookieText: "a=1; Domain=com; Secure",
			response:   "http://a.example.com/x",
			date:       "Sun, 06 Nov 1994 08:49:37 GMT",
			status:     "rejected",
			reason:     "domain has only one label",
			expires:    "-",
		},
		{
			name:       "domain mismatch precedes insecure secure rejection",
			cookieText: "a=1; Domain=other.com; Secure",
			response:   "http://a.example.com/x",
			date:       "Sun, 06 Nov 1994 08:49:37 GMT",
			status:     "rejected",
			reason:     "domain does not match the URL host",
			expires:    "-",
		},
		{
			name:       "secure cookie rejected from http",
			cookieText: "a=1; Secure",
			response:   "http://a.example.com/x",
			date:       "Sun, 06 Nov 1994 08:49:37 GMT",
			status:     "rejected",
			reason:     "Secure cookie from an http URL",
			expires:    "-",
		},
		{
			name:       "max age zero overrides future expires",
			cookieText: "a=1; Max-Age=0; Expires=Mon, 07 Nov 1994 08:00:00 GMT",
			response:   "https://a.com/app/x",
			date:       "Sun, 06 Nov 1994 08:49:37 GMT",
			status:     "expired",
			domain:     "a.com",
			path:       "/app",
			expires:    "-",
		},
		{
			name:       "expires equality is expired",
			cookieText: "a=1; Expires=Sun, 06 Nov 1994 08:49:37 GMT",
			response:   "https://a.com/",
			date:       "Sun, 06 Nov 1994 08:49:37 GMT",
			status:     "expired",
			domain:     "a.com",
			path:       "/",
			expires:    "-",
		},
		{
			name:       "positive max age takes precedence over expired expires",
			cookieText: "a=1; Max-Age=3600; Expires=Thu, 01 Jan 1970 00:00:00 GMT",
			response:   "https://a.com/app/x",
			date:       "Sun, 06 Nov 1994 08:49:37 GMT",
			status:     "stored",
			domain:     "a.com",
			path:       "/app",
			expires:    "Sun, 06 Nov 1994 09:49:37 GMT",
		},
		{
			name:       "expiry can pass accepted input year range",
			cookieText: "a=1; Max-Age=1",
			response:   "https://a.com/",
			date:       "Tue, 31 Dec 2999 23:59:59 GMT",
			status:     "stored",
			domain:     "a.com",
			path:       "/",
			expires:    "Wed, 01 Jan 3000 00:00:00 GMT",
		},
		{
			name:       "session cookie",
			cookieText: "a=1",
			response:   "https://a.com/",
			date:       "Sun, 06 Nov 1994 08:49:37 GMT",
			status:     "stored",
			domain:     "a.com",
			path:       "/",
			expires:    "session",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			cookie, err := parseCookieHeader("Set-Cookie: " + test.cookieText)
			if err != nil {
				t.Fatal(err)
			}
			response, err := ParseURL(test.response)
			if err != nil {
				t.Fatal(err)
			}
			now, ok := message.ParseDate(test.date)
			if !ok {
				t.Fatalf("invalid test date %q", test.date)
			}
			got := StoreCookies([]Cookie{cookie}, response, now)
			if len(got) != 1 {
				t.Fatalf("stored cookies = %#v, want one result", got)
			}
			result := got[0]
			if result.Status != test.status || result.Reason != test.reason ||
				result.Domain != test.domain || result.Path != test.path {
				t.Fatalf("storage result = %#v, want status %q reason %q domain %q path %q", result, test.status, test.reason, test.domain, test.path)
			}
			switch test.expires {
			case "-", "session":
				if result.HasExpiry {
					t.Fatalf("HasExpiry = true with expiry %d, want no expiry (%q)", result.Expires, test.expires)
				}
				if test.expires == "session" && result.Status != "stored" {
					t.Fatalf("session result status = %q, want stored", result.Status)
				}
			default:
				if !result.HasExpiry {
					t.Fatal("HasExpiry = false, want expiry")
				}
				if formatted := message.FormatDate(result.Expires); formatted != test.expires {
					t.Fatalf("formatted expiry = %q, want %q", formatted, test.expires)
				}
			}
		})
	}
}

func TestStoreCookieMaxAgeAdditionBoundaries(t *testing.T) {
	tests := []struct {
		name string
		now  int64
		age  int64
		want int64
	}{
		{
			name: "exact maximum boundary",
			now:  9223372036854775806,
			age:  1,
			want: 9223372036854775807,
		},
		{
			name: "positive overflow saturates",
			now:  9223372036854775807,
			age:  1,
			want: 9223372036854775807,
		},
		{
			name: "negative starting value remains safe",
			now:  -9223372036854775808,
			age:  1,
			want: -9223372036854775807,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result := StoreCookies(
				[]Cookie{{Name: "a", Value: "1", HasMaxAge: true, MaxAge: test.age}},
				URL{Scheme: "https", Host: "a.com", Path: "/"},
				test.now,
			)
			if len(result) != 1 || result[0].Status != "stored" || !result[0].HasExpiry {
				t.Fatalf("StoreCookies result = %#v, want one stored cookie with expiry", result)
			}
			if result[0].Expires != test.want {
				t.Fatalf("expiry = %d, want %d", result[0].Expires, test.want)
			}
		})
	}
}
