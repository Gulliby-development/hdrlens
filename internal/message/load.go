package message

import (
	"errors"

	"github.com/Gulliby-development/hdrlens/internal/base"
)

// Loaded is a parsed header block with its display name for diagnostics.
type Loaded struct {
	Display string
	Fields  Fields
}

// Load reads and parses a header block from stdin or a named file.
func Load(io base.IO, arg string) (Loaded, error) {
	source, err := base.ReadSource(io, arg)
	if err != nil {
		return Loaded{}, err
	}

	loaded := Loaded{Display: source.Display}
	fields, err := parseNormalized(source.Text)
	if err != nil {
		return Loaded{}, loaded.Wrap(err)
	}
	if len(fields) == 0 {
		return Loaded{}, loaded.Wrap(errors.New("no header fields"))
	}
	loaded.Fields = fields
	return loaded, nil
}

// Wrap prefixes err with this source's display name.
func (loaded Loaded) Wrap(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(loaded.Display + ": " + err.Error())
}
