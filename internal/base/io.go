package base

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// IO abstracts the input and output streams used by commands.
type IO interface {
	ReadStdin() ([]byte, error)
	ReadFile(path string) ([]byte, error)
	Out(text string)
	Err(text string)
}

// Source is decoded input together with the name used when reporting errors.
type Source struct {
	Display string
	Text    string
}

// ReadSource reads and validates stdin or a named file.
func ReadSource(io IO, arg string) (Source, error) {
	display := Show(arg)
	var data []byte
	var err error
	if arg == "-" {
		display = "<stdin>"
		data, err = io.ReadStdin()
	} else {
		data, err = io.ReadFile(arg)
	}
	if err != nil {
		return Source{}, errors.New(display + ": cannot read the file")
	}
	if !utf8.Valid(data) {
		return Source{}, errors.New(display + ": input is not valid UTF-8")
	}

	text := string(data)
	text = strings.TrimPrefix(text, "\uFEFF")
	return Source{Display: display, Text: text}, nil
}

// MemIO is an in-memory implementation of IO for commands and tests.
type MemIO struct {
	Stdin  []byte
	Files  map[string][]byte
	Stdout string
	Stderr string
}

// ReadStdin returns the configured standard input bytes.
func (m *MemIO) ReadStdin() ([]byte, error) {
	return m.Stdin, nil
}

// ReadFile returns the configured file bytes.
func (m *MemIO) ReadFile(path string) ([]byte, error) {
	data, ok := m.Files[path]
	if !ok {
		return nil, errors.New("file does not exist")
	}
	return data, nil
}

// Out appends text to the in-memory standard output.
func (m *MemIO) Out(text string) {
	m.Stdout += text
}

// Err appends text to the in-memory standard error.
func (m *MemIO) Err(text string) {
	m.Stderr += text
}
