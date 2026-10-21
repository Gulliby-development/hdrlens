package cli

import (
	"os"
	"strings"
	"testing"
)

func TestReadmeDocumentsContract(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	contents := []struct {
		name string
		want string
	}{
		{name: "install heading", want: "## Install"},
		{name: "usage heading", want: "## Usage"},
		{name: "commands heading", want: "## Commands"},
		{name: "exit codes heading", want: "## Exit codes"},
		{name: "development heading", want: "## Development"},
		{name: "check usage", want: "usage: hdrlens check FILE"},
		{name: "list usage", want: "       hdrlens list FILE [--by-name]"},
		{name: "accept usage", want: "       hdrlens accept FILE OFFER..."},
		{name: "range usage", want: "       hdrlens range FILE LENGTH"},
		{name: "cache usage", want: "       hdrlens cache FILE --now DATE [--shared]"},
		{name: "cookie usage", want: "       hdrlens cookie FILE URL --now DATE [--send URL]"},
		{name: "full help block", want: `usage: hdrlens check FILE
       hdrlens list FILE [--by-name]
       hdrlens accept FILE OFFER...
       hdrlens range FILE LENGTH
       hdrlens cache FILE --now DATE [--shared]
       hdrlens cookie FILE URL --now DATE [--send URL]`},
		{name: "resp.txt input block", want: `# a cached response
Date: Sun, 06 Nov 1994 08:49:37 GMT
Cache-Control: public, max-age=600
Last-Modified: Sun, 06 Nov 1994 07:00:00 GMT
Age: 100
Set-Cookie: sid=abc; Path=/app; Secure; HttpOnly; SameSite=Lax; Max-Age=3600
Set-Cookie: theme=dark; Domain=example.com; Expires=Mon, 07 Nov 1994 08:00:00 GMT
Set-Cookie: tmp=1; Domain=com`},
		{name: "req.txt input block", want: `Accept: text/html, application/json;q=0.9, */*;q=0.1
Range: bytes=0-99,500-,-50
If-None-Match: W/"v1", "v2"
If-Modified-Since: Sun, 06 Nov 1994 08:49:37 GMT`},
		{name: "check output block", want: `ok, 7 fields
distinct    5
repeated    1
longest     69`},
		{name: "accept output block", want: `OFFER             Q    RANGE
application/json  0.9  application/json
text/plain        0.1  */*
image/png         0.1  */*
selected    application/json`},
	}

	for _, content := range contents {
		t.Run(content.name, func(t *testing.T) {
			if !strings.Contains(string(readme), content.want) {
				t.Errorf("README.md does not contain required %s", content.name)
			}
		})
	}
}
