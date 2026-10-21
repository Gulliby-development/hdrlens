package byterange

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

// Usage is the command's usage line.
const Usage = "hdrlens range FILE LENGTH"

// Run validates and resolves the byte ranges in FILE for a resource of LENGTH bytes.
func Run(args []string, io base.IO) (int, error) {
	parsedArgs, err := base.Parse(args, nil)
	if err != nil {
		return 0, err
	}
	if len(parsedArgs.Args) == 0 {
		return 0, errors.New("missing FILE")
	}
	if len(parsedArgs.Args) == 1 {
		return 0, errors.New("missing LENGTH")
	}
	if len(parsedArgs.Args) > 2 {
		return 0, errors.New("too many arguments")
	}

	loaded, err := message.Load(io, parsedArgs.Args[0])
	if err != nil {
		return 0, err
	}
	ranges, err := Parse(loaded.Fields)
	if err != nil {
		return 0, loaded.Wrap(err)
	}
	length, err := ParseLength(parsedArgs.Args[1])
	if err != nil {
		return 0, err
	}

	resolution := Resolve(ranges, length)
	if !ranges.Present {
		io.Out(linesOutput(base.Line("status", "200"), base.Line("served", strconv.FormatInt(length, 10))))
		return 0, nil
	}
	if len(resolution.Spans) == 1 &&
		resolution.Spans[0].Start == 0 &&
		resolution.Spans[0].End == length-1 {
		io.Out(linesOutput(base.Line("status", "200"), base.Line("served", strconv.FormatInt(length, 10))))
		return 0, nil
	}
	if len(resolution.Spans) == 0 {
		io.Out(linesOutput(base.Line("status", "416"), base.Line("range", "bytes */"+strconv.FormatInt(length, 10))))
		return 1, nil
	}

	lines := []string{
		base.Line("status", "206"),
		base.Line("ranges", strconv.Itoa(len(resolution.Spans))),
		base.Line("served", strconv.FormatInt(resolution.Total, 10)),
	}
	for _, span := range resolution.Spans {
		lines = append(lines, base.Line("range",
			"bytes "+strconv.FormatInt(span.Start, 10)+"-"+strconv.FormatInt(span.End, 10)+"/"+strconv.FormatInt(length, 10)))
	}
	io.Out(linesOutput(lines...))
	return 0, nil
}

func linesOutput(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
