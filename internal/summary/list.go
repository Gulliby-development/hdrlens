package summary

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

const ListUsage = "hdrlens list FILE [--by-name]"

// RunList prints the header fields in source order or sorted by name.
func RunList(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, []base.Spec{{Name: "--by-name"}})
	if err != nil {
		return 0, err
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

	fields := loaded.Fields
	if parsed.Flags["--by-name"] {
		fields = append(message.Fields(nil), loaded.Fields...)
		sortFieldsByName(fields)
	}

	rows := make([][]string, 0, len(fields))
	for _, field := range fields {
		rows = append(rows, []string{
			strconv.Itoa(field.Line),
			field.Name,
			field.Value,
		})
	}

	lines := base.Table([]string{"LINE", "NAME", "VALUE"}, rows)
	io.Out(strings.Join(lines, "\n") + "\n")
	return 0, nil
}

// sortFieldsByName orders fields by their ASCII-lowercase names, then by line.
func sortFieldsByName(fields message.Fields) {
	for i := 1; i < len(fields); i++ {
		for j := i; j > 0 && fieldNameLess(fields[j], fields[j-1]); j-- {
			fields[j], fields[j-1] = fields[j-1], fields[j]
		}
	}
}

func fieldNameLess(left, right message.Field) bool {
	leftName := strings.ToLower(left.Name)
	rightName := strings.ToLower(right.Name)
	if leftName != rightName {
		return leftName < rightName
	}
	return left.Line < right.Line
}
