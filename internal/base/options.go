package base

import (
	"errors"
	"strings"
)

// Spec describes an option and whether it consumes a value.
type Spec struct {
	Name  string
	Takes bool
}

// Parsed contains the options and positional arguments parsed from argv.
type Parsed struct {
	Flags  map[string]bool
	Values map[string]string
	Args   []string
}

// Parse reads options from left to right and returns the first error.
func Parse(argv []string, specs []Spec) (Parsed, error) {
	parsed := Parsed{
		Flags:  make(map[string]bool),
		Values: make(map[string]string),
		Args:   make([]string, 0, len(argv)),
	}
	seen := make(map[string]bool)

	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if !strings.HasPrefix(arg, "--") {
			parsed.Args = append(parsed.Args, arg)
			continue
		}

		name := arg
		value := ""
		hasValue := false
		if equals := strings.IndexByte(arg, '='); equals >= 0 {
			name = arg[:equals]
			value = arg[equals+1:]
			hasValue = true
		}

		var option *Spec
		for j := range specs {
			if specs[j].Name == name {
				option = &specs[j]
				break
			}
		}
		if option == nil {
			return Parsed{}, errors.New("unknown option " + Show(arg))
		}
		if seen[name] {
			return Parsed{}, errors.New("repeated option " + name)
		}
		seen[name] = true

		if !option.Takes {
			if hasValue {
				return Parsed{}, errors.New(name + " takes no value")
			}
			parsed.Flags[name] = true
			continue
		}

		if !hasValue {
			if i+1 >= len(argv) {
				return Parsed{}, errors.New(name + " needs a value")
			}
			i++
			value = argv[i]
		}
		parsed.Values[name] = value
	}

	return parsed, nil
}
