package cli

import (
	"errors"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/byterange"
	"github.com/Gulliby-development/hdrlens/internal/cookies"
	"github.com/Gulliby-development/hdrlens/internal/freshness"
	"github.com/Gulliby-development/hdrlens/internal/negotiate"
	"github.com/Gulliby-development/hdrlens/internal/summary"
)

// Usage is the help text for the hdrlens commands.
const Usage = "usage: hdrlens check FILE\n" +
	"       hdrlens list FILE [--by-name]\n" +
	"       hdrlens accept FILE OFFER...\n" +
	"       hdrlens range FILE LENGTH\n" +
	"       hdrlens cache FILE --now DATE [--shared]\n" +
	"       hdrlens cookie FILE URL --now DATE [--send URL]\n"

// Run dispatches a command and reports command errors consistently.
func Run(argv []string, io base.IO) int {
	if len(argv) == 0 {
		io.Err("hdrlens: missing command (check, list, accept, range, cache or cookie)\n")
		return 2
	}
	if len(argv) == 1 && argv[0] == "--help" {
		io.Out(Usage)
		return 0
	}

	var status int
	var err error
	switch argv[0] {
	case "check":
		status, err = summary.RunCheck(argv[1:], io)
	case "list":
		status, err = summary.RunList(argv[1:], io)
	case "accept":
		status, err = negotiate.Run(argv[1:], io)
	case "range":
		status, err = byterange.Run(argv[1:], io)
	case "cache":
		status, err = freshness.Run(argv[1:], io)
	case "cookie":
		status, err = cookies.Run(argv[1:], io)
	default:
		err = errors.New("unknown command " + base.Show(argv[0]))
	}

	if err != nil {
		io.Err("hdrlens: " + err.Error() + "\n")
		return 2
	}
	return status
}
