package summary

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

const CheckUsage = "hdrlens check FILE"

// RunCheck validates and summarizes a header source.
func RunCheck(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return 0, err
	}
	if len(parsed.Args) == 0 {
		return 0, errors.New("missing FILE")
	}
	if len(parsed.Args) > 1 {
		return 0, errors.New("too many arguments")
	}

	source, err := base.ReadSource(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}

	fields, err := message.Parse("\uFEFF" + source.Text)
	loaded := message.Loaded{Display: source.Display}
	if err != nil {
		var problem *message.Problem
		if errors.As(err, &problem) {
			io.Out(problem.Error() + "\n1 problem\n")
			return 1, nil
		}
		return 0, loaded.Wrap(err)
	}
	if len(fields) == 0 {
		return 0, loaded.Wrap(errors.New("no header fields"))
	}

	stats := Summarize(fields)
	var output strings.Builder
	output.WriteString("ok, ")
	output.WriteString(base.Plural(stats.Total, "field"))
	output.WriteByte('\n')
	output.WriteString(base.Line("distinct", intString(stats.Distinct)))
	output.WriteByte('\n')
	output.WriteString(base.Line("repeated", intString(stats.Repeated)))
	output.WriteByte('\n')
	output.WriteString(base.Line("longest", intString(stats.Longest)))
	output.WriteByte('\n')
	io.Out(output.String())
	return 0, nil
}

func intString(value int) string {
	return strconv.Itoa(value)
}
