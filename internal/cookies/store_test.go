package cookies

import (
	"testing"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

func storeCookiesForTest(t *testing.T, text, responseText, dateText string) []StoredCookie {
	t.Helper()
	fields, err := message.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	cookies, err := ParseCookies(fields)
	if err != nil {
		t.Fatal(err)
	}
	response, err := ParseURL(responseText)
	if err != nil {
		t.Fatal(err)
	}
	now, ok := message.ParseDate(dateText)
	if !ok {
		t.Fatalf("invalid literal test date %q", dateText)
	}
	return StoreCookies(cookies, response, now)
}

func TestSelectCookieLiteralSendScenarios(t *testing.T) {
	const file = "Set-Cookie: a=1; Path=/\n" +
		"Set-Cookie: b=2; Path=/shop\n" +
		"Set-Cookie: c=3; Path=/shop/\n" +
		"Set-Cookie: d=4; Path=/shopping\n" +
		"Set-Cookie: dup=first; Path=/shop/cart\n" +
		"Set-Cookie: dup=second; Path=/shop/cart\n" +
		"Set-Cookie: secure=ok; Secure; Path=/shop\n" +
		"Set-Cookie: gone=gone; Max-Age=0; Path=/shop\n" +
		"Set-Cookie: e=5; Domain=shop.example.com; Path=/shop/cart"
	tests := []struct {
		name    string
		request string
		want    string
	}{
		{
			name:    "longest paths first and equal paths retain file order",
			request: "https://shop.example.com/shop/cart",
			want:    "dup=first; dup=second; e=5; c=3; b=2; secure=ok; a=1",
		},
		{
			name:    "cookie path exact match",
			request: "https://shop.example.com/shop",
			want:    "b=2; secure=ok; a=1",
		},
		{
			name:    "path prefix does not match shopping",
			request: "https://shop.example.com/shopping",
			want:    "d=4; a=1",
		},
		{
			name:    "slash boundary matches nested request path",
			request: "https://shop.example.com/shop/x",
			want:    "c=3; b=2; secure=ok; a=1",
		},
		{
			name:    "explicit domain matches a subdomain but host only does not",
			request: "https://sub.shop.example.com/shop/cart/x",
			want:    "e=5",
		},
		{
			name:    "domain and host-only cookies reject other hosts",
			request: "https://other.example.com/shop/cart",
			want:    "-",
		},
		{
			name:    "secure cookie is omitted from http request",
			request: "http://shop.example.com/shop/cart",
			want:    "dup=first; dup=second; e=5; c=3; b=2; a=1",
		},
	}
	stored := storeCookiesForTest(
		t,
		file,
		"https://shop.example.com/shop/list",
		"Sun, 06 Nov 1994 08:49:37 GMT",
	)
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			request, err := ParseURL(test.request)
			if err != nil {
				t.Fatal(err)
			}
			got := SendValue(SelectCookies(stored, request))
			if got != test.want {
				t.Fatalf("SendValue = %q, want %q", got, test.want)
			}
		})
	}
}

func TestCookieCommandExampleTablesAndSendRows(t *testing.T) {
	const response = "# a cached response\n" +
		"Date: Sun, 06 Nov 1994 08:49:37 GMT\n" +
		"Cache-Control: public, max-age=600\n" +
		"Last-Modified: Sun, 06 Nov 1994 07:00:00 GMT\n" +
		"Age: 100\n" +
		"Set-Cookie: sid=abc; Path=/app; Secure; HttpOnly; SameSite=Lax; Max-Age=3600\n" +
		"Set-Cookie: theme=dark; Domain=example.com; Expires=Mon, 07 Nov 1994 08:00:00 GMT\n" +
		"Set-Cookie: tmp=1; Domain=com"
	const jar = "Set-Cookie: a=1; Path=/\n" +
		"Set-Cookie: b=2; Path=/shop/cart; Max-Age=0\n" +
		"Set-Cookie: c=3; Secure\n" +
		"Set-Cookie: e=5; Domain=shop.example.com\n" +
		"Set-Cookie: f=6; Path=/shop"
	const responseOutput = "LINE  NAME   STATUS                                DOMAIN           PATH  EXPIRES\n" +
		"6     sid    stored                                www.example.com  /app  Sun, 06 Nov 1994 09:49:37 GMT\n" +
		"7     theme  stored                                example.com      /app  Mon, 07 Nov 1994 08:00:00 GMT\n" +
		"8     tmp    rejected (domain has only one label)  -                -     -\n" +
		"send        sid=abc; theme=dark\n"
	const jarOutput = "LINE  NAME  STATUS                                     DOMAIN            PATH        EXPIRES\n" +
		"1     a     stored                                     shop.example.com  /           session\n" +
		"2     b     expired                                    shop.example.com  /shop/cart  -\n" +
		"3     c     rejected (Secure cookie from an http URL)  -                 -           -\n" +
		"4     e     stored                                     shop.example.com  /shop       session\n" +
		"5     f     stored                                     shop.example.com  /shop       session\n" +
		"send        e=5; f=6; a=1\n"
	const subdomainOutput = "LINE  NAME  STATUS                                     DOMAIN            PATH        EXPIRES\n" +
		"1     a     stored                                     shop.example.com  /           session\n" +
		"2     b     expired                                    shop.example.com  /shop/cart  -\n" +
		"3     c     rejected (Secure cookie from an http URL)  -                 -           -\n" +
		"4     e     stored                                     shop.example.com  /shop       session\n" +
		"5     f     stored                                     shop.example.com  /shop       session\n" +
		"send        e=5\n"
	tests := []struct {
		name   string
		args   []string
		files  map[string][]byte
		code   int
		stdout string
	}{
		{
			name: "response example",
			args: []string{
				"resp.txt", "https://www.example.com/app/home", "--now",
				"Sun, 06 Nov 1994 08:49:37 GMT", "--send",
				"https://www.example.com/app/x",
			},
			files:  map[string][]byte{"resp.txt": []byte(response)},
			code:   1,
			stdout: responseOutput,
		},
		{
			name: "jar example",
			args: []string{
				"jar.txt", "http://shop.example.com/shop/list", "--now",
				"Sun, 06 Nov 1994 08:49:37 GMT", "--send",
				"http://shop.example.com/shop/cart",
			},
			files:  map[string][]byte{"jar.txt": []byte(jar)},
			code:   1,
			stdout: jarOutput,
		},
		{
			name: "jar explicit domain sends to subdomain",
			args: []string{
				"jar.txt", "http://shop.example.com/shop/list", "--now",
				"Sun, 06 Nov 1994 08:49:37 GMT", "--send",
				"https://sub.shop.example.com/shop/x",
			},
			files:  map[string][]byte{"jar.txt": []byte(jar)},
			code:   1,
			stdout: subdomainOutput,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Files: test.files}
			code, err := Run(test.args, memory)
			if err != nil || code != test.code {
				t.Fatalf("Run = (%d, %v), want (%d, nil)", code, err, test.code)
			}
			if memory.Stdout != test.stdout || memory.Stderr != "" {
				t.Fatalf("stdout/stderr = %q / %q, want %q / empty", memory.Stdout, memory.Stderr, test.stdout)
			}
		})
	}
}

func TestCookieCommandSessionAndExpiredOnlyResults(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "session-only cookie with no send row",
			text: "Set-Cookie: a=1",
			want: "LINE  NAME  STATUS  DOMAIN  PATH  EXPIRES\n" +
				"1     a     stored  a.com   /     session\n",
		},
		{
			name: "expired-only cookies do not return a finding",
			text: "Set-Cookie: a=1; Max-Age=0",
			want: "LINE  NAME  STATUS   DOMAIN  PATH  EXPIRES\n" +
				"1     a     expired  a.com   /     -\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Files: map[string][]byte{"a.txt": []byte(test.text)}}
			code, err := Run(
				[]string{"a.txt", "https://a.com/", "--now", "Sun, 06 Nov 1994 08:49:37 GMT"},
				memory,
			)
			if err != nil || code != 0 {
				t.Fatalf("Run = (%d, %v), want (0, nil)", code, err)
			}
			if memory.Stdout != test.want || memory.Stderr != "" {
				t.Fatalf("stdout/stderr = %q / %q, want %q / empty", memory.Stdout, memory.Stderr, test.want)
			}
		})
	}
}

func TestCookieCommandRawPathDisplayAndMatching(t *testing.T) {
	const table = "LINE  NAME  STATUS  DOMAIN  PATH  EXPIRES\n" +
		"1     a     stored  a.com   /?    session\n"
	tests := []struct {
		name     string
		response string
		send     string
		want     string
	}{
		{
			name:     "non-ascii raw prefix matches and displays as question mark",
			response: "https://a.com/é/x",
			send:     "https://a.com/é/y",
			want:     table + "send        a=1\n",
		},
		{
			name:     "different raw path does not match the same display cell",
			response: "https://a.com/é/x",
			send:     "https://a.com/?/y",
			want:     table + "send        -\n",
		},
		{
			name:     "tab raw prefix matches and displays as question mark",
			response: "https://a.com/\t/x",
			send:     "https://a.com/\t/y",
			want:     table + "send        a=1\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Files: map[string][]byte{"a.txt": []byte("Set-Cookie: a=1")}}
			code, err := Run(
				[]string{
					"a.txt", test.response, "--now",
					"Sun, 06 Nov 1994 08:49:37 GMT", "--send", test.send,
				},
				memory,
			)
			if err != nil || code != 0 {
				t.Fatalf("Run = (%d, %v), want (0, nil)", code, err)
			}
			if memory.Stdout != test.want || memory.Stderr != "" {
				t.Fatalf("stdout/stderr = %q / %q, want %q / empty", memory.Stdout, memory.Stderr, test.want)
			}
		})
	}
}

func TestCookieCommandValidationPrecedence(t *testing.T) {
	const date = "Sun, 06 Nov 1994 08:49:37 GMT"
	const cookie = "Set-Cookie: a=1"
	tests := []struct {
		name  string
		args  []string
		files map[string][]byte
		want  string
	}{
		{
			name: "unknown option before required option and arguments",
			args: []string{"--bogus"},
			want: "unknown option \"--bogus\"",
		},
		{
			name: "missing option value before missing arguments",
			args: []string{"--now"},
			want: "--now needs a value",
		},
		{
			name: "repeated now precedes missing arguments",
			args: []string{"--now", date, "--now", date},
			want: "repeated option --now",
		},
		{
			name: "missing now precedes positional arguments",
			args: []string{"a.txt"},
			want: "missing --now",
		},
		{
			name: "missing file follows now",
			args: []string{"--now", date},
			want: "missing FILE",
		},
		{
			name: "missing url precedes file loading",
			args: []string{"absent.txt", "--now", date},
			want: "missing URL",
		},
		{
			name: "too many arguments precede file loading",
			args: []string{"absent.txt", "https://a.com/", "extra", "--now", date},
			want: "too many arguments",
		},
		{
			name:  "response URL precedes invalid date and send URL",
			args:  []string{"a.txt", "ftp://a.com/", "--now", "yesterday", "--send", "ftp://b.com/"},
			files: map[string][]byte{"a.txt": []byte(cookie)},
			want:  "invalid URL \"ftp://a.com/\"",
		},
		{
			name:  "date precedes send URL",
			args:  []string{"a.txt", "https://a.com/", "--now", "yesterday", "--send", "ftp://b.com/"},
			files: map[string][]byte{"a.txt": []byte(cookie)},
			want:  "invalid date \"yesterday\"",
		},
		{
			name:  "explicit empty date is validated",
			args:  []string{"a.txt", "https://a.com/", "--now=", "--send", "ftp://b.com/"},
			files: map[string][]byte{"a.txt": []byte(cookie)},
			want:  "invalid date \"\"",
		},
		{
			name:  "send URL follows a valid date",
			args:  []string{"a.txt", "https://a.com/", "--now", date, "--send", "ftp://b.com/"},
			files: map[string][]byte{"a.txt": []byte(cookie)},
			want:  "invalid URL \"ftp://b.com/\"",
		},
		{
			name:  "explicit empty send URL is validated",
			args:  []string{"a.txt", "https://a.com/", "--now", date, "--send", ""},
			files: map[string][]byte{"a.txt": []byte(cookie)},
			want:  "invalid URL \"\"",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Files: test.files}
			code, err := Run(test.args, memory)
			if err == nil || err.Error() != test.want || code != 0 {
				t.Fatalf("Run = (%d, %v), want (0, %q)", code, err, test.want)
			}
			if memory.Stdout != "" || memory.Stderr != "" {
				t.Fatalf("error wrote stdout/stderr %q / %q, want both empty", memory.Stdout, memory.Stderr)
			}
		})
	}
}

func TestCookieCommandHeaderAndCookieProblemsPrecedeValues(t *testing.T) {
	const date = "Sun, 06 Nov 1994 08:49:37 GMT"
	tests := []struct {
		name string
		text string
		args []string
		want string
	}{
		{
			name: "complete header parsing precedes cookie syntax",
			text: "Set-Cookie: sid\n continuation",
			args: []string{"bad.txt", "ftp://a.com/", "--now", "yesterday", "--send", "ftp://b.com/"},
			want: "\"bad.txt\": HL002 continuation line at 2:1",
		},
		{
			name: "cookie syntax precedes response URL and date",
			text: "Set-Cookie: sid",
			args: []string{"bad.txt", "ftp://a.com/", "--now", "yesterday", "--send", "ftp://b.com/"},
			want: "\"bad.txt\": HK001 missing \"=\" at 1:13",
		},
		{
			name: "missing set-cookie precedes response URL and date",
			text: "Accept: text/html",
			args: []string{"req.txt", "ftp://a.com/", "--now", "yesterday", "--send", "ftp://b.com/"},
			want: "\"req.txt\": no Set-Cookie field",
		},
		{
			name: "response url is only checked after a valid cookie file",
			text: "Set-Cookie: a=1",
			args: []string{"a.txt", "ftp://a.com/", "--now", date},
			want: "invalid URL \"ftp://a.com/\"",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			memory := &base.MemIO{Files: map[string][]byte{
				test.args[0]: []byte(test.text),
			}}
			code, err := Run(test.args, memory)
			if err == nil || err.Error() != test.want || code != 0 {
				t.Fatalf("Run = (%d, %v), want (0, %q)", code, err, test.want)
			}
			if memory.Stdout != "" || memory.Stderr != "" {
				t.Fatalf("error wrote stdout/stderr %q / %q, want both empty", memory.Stdout, memory.Stderr)
			}
		})
	}
}
