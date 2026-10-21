<p align="center"><img src="docs/logo.svg" width="96" alt="" /></p>

# hdrlens

**hdrlens** is an offline command-line tool for web developers who inspect HTTP headers in saved `curl -i` dumps, logs, or test fixtures. Give it a text file containing header fields and it can validate and list them, choose a media type from `Accept`, resolve `Range` requests, evaluate cache freshness, or show which cookies are stored and sent. It reads local files or standard input; it does not make network requests.

The input is header-only: use header field lines such as `Accept: text/html`, not an HTTP status line or response body. The examples below can be saved as `resp.txt` and `req.txt`. Pass `-` in place of a file path to read standard input. Cache and cookie evaluations use an explicit `--now` date, so their results do not depend on the system clock.

[![CI](https://github.com/Gulliby-development/hdrlens/actions/workflows/ci.yml/badge.svg)](https://github.com/Gulliby-development/hdrlens/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Why

Header blocks are easy to read and hard to interpret: which type does that `Accept` line pick, which bytes does that `Range` select, is that response still fresh, which cookies would a client keep? `hdrlens` answers these from a saved file with exact, documented rules and a fixed `--now` date, so the output is the same on every machine and fits in a bug report or a test fixture.

## Install

Build from the project source with the Go toolchain version declared in `go.mod` (Go 1.27.0). The project uses only the Go standard library; building and running the commands needs no third-party packages or network access.

On macOS or Linux:

```sh
go build -o hdrlens .
./hdrlens --help
```

On Windows PowerShell:

```powershell
go build -o hdrlens.exe .
.\hdrlens.exe --help
```

The resulting executable is standalone. Keep it in the current directory or put it on your `PATH` to use `hdrlens` in the command examples.

## Usage

```text
usage: hdrlens check FILE
       hdrlens list FILE [--by-name]
       hdrlens accept FILE OFFER...
       hdrlens range FILE LENGTH
       hdrlens cache FILE --now DATE [--shared]
       hdrlens cookie FILE URL --now DATE [--send URL]
```

`FILE` names a UTF-8 text file of header fields; `-` reads the same format from standard input. A single leading UTF-8 byte-order mark is accepted. Header names are case-insensitive for command processing. Dates use the HTTP-date form shown in the examples, for example `Sun, 06 Nov 1994 08:49:37 GMT`.

## Commands

- **`check FILE`** validates the header fields. On valid input it prints the field count and distinct, repeated, and longest-value statistics. If a field line is malformed, it prints the first problem and a problem count.
- **`list FILE [--by-name]`** prints a table of line number, field name, and value in file order. `--by-name` sorts by lower-case field name, then by line number.
- **`accept FILE OFFER...`** compares each offered `TYPE/SUB` media type with the file's `Accept` fields and prints its quality and deciding range, followed by the selected offer. If there is no `Accept` field, all offers have quality 1. The first offer wins a tie; no positively rated offer is a negative result.
- **`range FILE LENGTH`** resolves any `Range` fields against a non-negative resource length. It reports status 200 for the full resource, 206 and the merged byte spans for a partial result, or 416 when no requested span is satisfiable.
- **`cache FILE --now DATE [--shared]`** evaluates cache directives and date fields at the supplied time. It reports whether the response is fresh, stale, requires revalidation, or is uncacheable. `--shared` applies shared-cache rules, including `s-maxage`.
- **`cookie FILE URL --now DATE [--send URL]`** evaluates `Set-Cookie` fields for the response URL at the supplied time, showing stored, expired, and rejected cookies. With `--send URL`, it also lists the stored cookies that match that request URL.

For example, save the following response fields as `resp.txt`:

```text
# a cached response
Date: Sun, 06 Nov 1994 08:49:37 GMT
Cache-Control: public, max-age=600
Last-Modified: Sun, 06 Nov 1994 07:00:00 GMT
Age: 100
Set-Cookie: sid=abc; Path=/app; Secure; HttpOnly; SameSite=Lax; Max-Age=3600
Set-Cookie: theme=dark; Domain=example.com; Expires=Mon, 07 Nov 1994 08:00:00 GMT
Set-Cookie: tmp=1; Domain=com
```

Save these request fields as `req.txt`:

```text
Accept: text/html, application/json;q=0.9, */*;q=0.1
Range: bytes=0-99,500-,-50
If-None-Match: W/"v1", "v2"
If-Modified-Since: Sun, 06 Nov 1994 08:49:37 GMT
```

Checking the response:

```console
$ hdrlens check resp.txt
ok, 7 fields
distinct    5
repeated    1
longest     69
```

Negotiating a type from the request's `Accept` field:

```console
$ hdrlens accept req.txt application/json text/plain image/png
OFFER             Q    RANGE
application/json  0.9  application/json
text/plain        0.1  */*
image/png         0.1  */*
selected    application/json
```

Other command forms are shown in the usage block above. For instance, cache and cookie commands require the explicit `--now DATE` value; cookie also takes the response `URL`, and `--send URL` is optional.

### Options

| Option | Used by | Description |
|---|---|---|
| `--by-name` | `list` | Sort by lower-case field name, then by line. |
| `--now DATE` | `cache`, `cookie` | The current time, an HTTP date such as `Sun, 06 Nov 1994 08:49:37 GMT` (required). |
| `--shared` | `cache` | Evaluate as a shared cache (`s-maxage`, `private`). |
| `--send URL` | `cookie` | Also list the stored cookies a request to `URL` would send. |
| `--help` | (alone) | Print the usage lines. |

## Exit codes

- **0** means the command completed successfully. For `accept`, this means an offer was selected; for `range`, a satisfiable response was produced; for `cache`, the verdict is fresh.
- **1** is a finding or negative answer: `check` found a malformed field, `accept` selected no offer, `range` produced 416, `cache` did not find a fresh response, or `cookie` rejected at least one cookie. The command prints its result to standard output.
- **2** means an error, such as invalid command arguments, an unreadable or malformed input file, or an invalid command value. Standard output is empty and standard error contains one line prefixed with `hdrlens:`.

## Development

The application and tests use only the Go standard library. From the repository root, run:

```sh
gofmt -l .
go vet ./...
go test ./...
```

`gofmt -l .` should print no paths. `go vet ./...` checks the packages, and `go test ./...` runs the automated tests, including the executable's malformed-standard-input check.

## License

MIT, see [LICENSE](LICENSE).

---

Built with BAKED
