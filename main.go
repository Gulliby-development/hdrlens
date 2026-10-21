package main

import (
	"io"
	"os"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/cli"
)

type realIO struct{}

var _ base.IO = realIO{}

func (realIO) ReadStdin() ([]byte, error) {
	return io.ReadAll(os.Stdin)
}

func (realIO) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (realIO) Out(text string) {
	_, _ = os.Stdout.Write([]byte(text))
}

func (realIO) Err(text string) {
	_, _ = os.Stderr.Write([]byte(text))
}

func main() {
	adapter := realIO{}
	os.Exit(cli.Run(os.Args[1:], adapter))
}
