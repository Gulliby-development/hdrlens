package freshness

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

// Usage is the command's usage line.
const Usage = "hdrlens cache FILE --now DATE [--shared]"

// Run validates and evaluates a stored response's cache freshness.
func Run(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, []base.Spec{
		{Name: "--now", Takes: true},
		{Name: "--shared"},
	})
	if err != nil {
		return 0, err
	}

	nowText, hasNow := parsed.Values["--now"]
	if !hasNow {
		return 0, errors.New("missing --now")
	}
	if len(parsed.Args) == 0 {
		return 0, errors.New("missing FILE")
	}
	if len(parsed.Args) > 1 {
		return 0, errors.New("too many arguments")
	}

	loaded, err := message.Load(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	control, err := ParseControl(loaded.Fields)
	if err != nil {
		return 0, loaded.Wrap(err)
	}

	now, validNow := message.ParseDate(nowText)
	if !validNow {
		return 0, errors.New("invalid date " + base.Show(nowText))
	}
	if !control.HasDate {
		return 0, loaded.Wrap(errors.New("missing Date field"))
	}

	result := Evaluate(control, now, parsed.Flags["--shared"])
	lines := []string{
		base.Line("verdict", result.Verdict),
		base.Line("lifetime", strconv.FormatInt(result.Lifetime, 10)+" ("+result.Source+")"),
		base.Line("age", strconv.FormatInt(result.Age, 10)),
	}
	switch result.Verdict {
	case "fresh":
		lines = append(lines, base.Line("left", strconv.FormatInt(result.Lifetime-result.Age, 10)))
	case "stale":
		lines = append(lines, base.Line("over", strconv.FormatInt(result.Age-result.Lifetime, 10)))
	}
	io.Out(strings.Join(lines, "\n") + "\n")
	if result.Verdict == "fresh" {
		return 0, nil
	}
	return 1, nil
}
