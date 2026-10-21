package base

// BEGIN acceptance rows imports: host-written from the audited spec table. Edit acceptanceActual, not this block.
import (
	acceptancejson "encoding/json"
	acceptancereflect "reflect"
	acceptancetesting "testing"
)

// END acceptance rows imports

import "errors"

// BEGIN acceptance rows: host-written from the audited spec table. Edit acceptanceActual, not this block.
func TestAcceptanceRowsBase(t *acceptancetesting.T) {
	var rows []struct {
		ID       string `json:"id"`
		Input    any    `json:"input"`
		Expected any    `json:"expected"`
	}
	if err := acceptancejson.Unmarshal([]byte("[{\"id\":\"AC-002:1\",\"input\":{\"args\":[\"\"],\"function\":\"Show\"},\"expected\":\"\\\"\\\"\"},{\"id\":\"AC-002:2\",\"input\":{\"args\":[\"abc\"],\"function\":\"Show\"},\"expected\":\"\\\"abc\\\"\"},{\"id\":\"AC-002:3\",\"input\":{\"args\":[\"café😀\\n\"],\"function\":\"Show\"},\"expected\":\"\\\"caf???\\\"\"},{\"id\":\"AC-002:4\",\"input\":{\"args\":[\"\\\"\\\\\"],\"function\":\"Show\"},\"expected\":\"\\\"\\\"\\\\\\\"\"},{\"id\":\"AC-002:5\",\"input\":{\"args\":[\"key\",\"v\"],\"function\":\"Line\"},\"expected\":\"key         v\"},{\"id\":\"AC-002:6\",\"input\":{\"args\":[\"abcdefghijk\",\"v\"],\"function\":\"Line\"},\"expected\":\"abcdefghijk  v\"},{\"id\":\"AC-002:7\",\"input\":{\"args\":[1,\"field\"],\"function\":\"Plural\"},\"expected\":\"1 field\"},{\"id\":\"AC-002:8\",\"input\":{\"args\":[0,\"field\"],\"function\":\"Plural\"},\"expected\":\"0 fields\"},{\"id\":\"AC-002:9\",\"input\":{\"args\":[-2,\"field\"],\"function\":\"Plural\"},\"expected\":\"-2 fields\"},{\"id\":\"AC-002:10\",\"input\":{\"args\":[[\"A\",\"BB\"],[[\"xxx\",\"c\"],[\"y\",\"\"]]],\"function\":\"Table\"},\"expected\":[\"A    BB\",\"xxx  c\",\"y\"]},{\"id\":\"AC-003:1\",\"input\":[\"file\",\"-\",\"-x\",\"--by-name\"],\"expected\":{\"Args\":[\"file\",\"-\",\"-x\"],\"Flags\":{\"--by-name\":true},\"Values\":{}}},{\"id\":\"AC-003:2\",\"input\":[\"--now=A=B\"],\"expected\":{\"Args\":[],\"Flags\":{},\"Values\":{\"--now\":\"A=B\"}}},{\"id\":\"AC-003:3\",\"input\":[\"--now=\"],\"expected\":{\"Args\":[],\"Flags\":{},\"Values\":{\"--now\":\"\"}}},{\"id\":\"AC-003:4\",\"input\":[\"--now\",\"--by-name\"],\"expected\":{\"Args\":[],\"Flags\":{},\"Values\":{\"--now\":\"--by-name\"}}},{\"id\":\"AC-003:5\",\"input\":[\"--\"],\"expected\":\"unknown option \\\"--\\\"\"},{\"id\":\"AC-003:6\",\"input\":[\"--help\"],\"expected\":\"unknown option \\\"--help\\\"\"},{\"id\":\"AC-003:7\",\"input\":[\"--write\"],\"expected\":\"unknown option \\\"--write\\\"\"},{\"id\":\"AC-003:8\",\"input\":[\"--by-name=1\"],\"expected\":\"--by-name takes no value\"},{\"id\":\"AC-003:9\",\"input\":[\"--now\"],\"expected\":\"--now needs a value\"},{\"id\":\"AC-003:10\",\"input\":[\"--now\",\"A\",\"--now\",\"B\"],\"expected\":\"repeated option --now\"},{\"id\":\"AC-003:11\",\"input\":[\"--by-name\",\"--by-name\"],\"expected\":\"repeated option --by-name\"},{\"id\":\"AC-003:12\",\"input\":[\"--write\",\"--now\"],\"expected\":\"unknown option \\\"--write\\\"\"}]"), &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		row := row
		t.Run(row.ID, func(t *acceptancetesting.T) {
			got, err := acceptanceActual(row.Input)
			if err != nil {
				t.Fatal(err)
			}
			text, err := acceptancejson.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var value any
			if err := acceptancejson.Unmarshal(text, &value); err != nil {
				t.Fatal(err)
			}
			if !acceptancereflect.DeepEqual(value, row.Expected) {
				want, _ := acceptancejson.Marshal(row.Expected)
				t.Fatalf("got %s, want %s", text, want)
			}
		})
	}
}

// END acceptance rows

var _ IO = (*MemIO)(nil)

func acceptanceActual(input any) (any, error) {
	if argvInput, ok := input.([]any); ok {
		argv := make([]string, len(argvInput))
		for i, arg := range argvInput {
			value, ok := arg.(string)
			if !ok {
				return nil, errors.New("Parse arguments must be strings")
			}
			argv[i] = value
		}
		parsed, err := Parse(argv, []Spec{{Name: "--by-name"}, {Name: "--now", Takes: true}})
		if err != nil {
			return err.Error(), nil
		}
		if parsed.Args == nil {
			parsed.Args = []string{}
		}
		return parsed, nil
	}

	object, ok := input.(map[string]any)
	if !ok {
		return nil, errors.New("acceptance input must be an object")
	}
	function, ok := object["function"].(string)
	if !ok {
		return nil, errors.New("acceptance function must be a string")
	}
	args, ok := object["args"].([]any)
	if !ok {
		return nil, errors.New("acceptance args must be an array")
	}

	switch function {
	case "Show":
		if len(args) != 1 {
			return nil, errors.New("Show requires one argument")
		}
		value, ok := args[0].(string)
		if !ok {
			return nil, errors.New("Show argument must be a string")
		}
		return Show(value), nil
	case "Line":
		if len(args) != 2 {
			return nil, errors.New("Line requires two arguments")
		}
		key, keyOK := args[0].(string)
		value, valueOK := args[1].(string)
		if !keyOK || !valueOK {
			return nil, errors.New("Line arguments must be strings")
		}
		return Line(key, value), nil
	case "Plural":
		if len(args) != 2 {
			return nil, errors.New("Plural requires two arguments")
		}
		number, numberOK := args[0].(float64)
		word, wordOK := args[1].(string)
		if !numberOK || !wordOK || number != float64(int(number)) {
			return nil, errors.New("Plural arguments have invalid types")
		}
		return Plural(int(number), word), nil
	case "Table":
		if len(args) != 2 {
			return nil, errors.New("Table requires two arguments")
		}
		headerInput, headerOK := args[0].([]any)
		rowsInput, rowsOK := args[1].([]any)
		if !headerOK || !rowsOK {
			return nil, errors.New("Table arguments must be arrays")
		}
		header := make([]string, len(headerInput))
		for i, cell := range headerInput {
			value, ok := cell.(string)
			if !ok {
				return nil, errors.New("Table header cells must be strings")
			}
			header[i] = value
		}
		rows := make([][]string, len(rowsInput))
		for i, rawRow := range rowsInput {
			rowInput, ok := rawRow.([]any)
			if !ok {
				return nil, errors.New("Table rows must be arrays")
			}
			row := make([]string, len(rowInput))
			for j, cell := range rowInput {
				value, ok := cell.(string)
				if !ok {
					return nil, errors.New("Table cells must be strings")
				}
				row[j] = value
			}
			rows[i] = row
		}
		return Table(header, rows), nil
	case "Parse":
		return nil, errors.New("Parse input must be an array")
	default:
		return nil, errors.New("unknown acceptance function")
	}
}

func TestReadSourceValidLiteralCases(t *acceptancetesting.T) {
	tests := []struct {
		name    string
		stdin   []byte
		files   map[string][]byte
		arg     string
		display string
		text    string
	}{
		{name: "stdin text", stdin: []byte("abc"), arg: "-", display: "<stdin>", text: "abc"},
		{name: "file with leading BOM", files: map[string][]byte{"a.txt": []byte("\uFEFFhello")}, arg: "a.txt", display: "\"a.txt\"", text: "hello"},
		{name: "one BOM removed from two", stdin: []byte("\uFEFF\uFEFFhello"), arg: "-", display: "<stdin>", text: "\uFEFFhello"},
		{name: "embedded BOM retained", stdin: []byte("a\uFEFFb"), arg: "-", display: "<stdin>", text: "a\uFEFFb"},
		{name: "empty stdin", stdin: []byte{}, arg: "-", display: "<stdin>", text: ""},
		{name: "unicode text", stdin: []byte("café😀"), arg: "-", display: "<stdin>", text: "café😀"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *acceptancetesting.T) {
			source, err := ReadSource(&MemIO{Stdin: test.stdin, Files: test.files}, test.arg)
			if err != nil {
				t.Fatal(err)
			}
			if source.Display != test.display || source.Text != test.text {
				t.Fatalf("got Source{%q, %q}, want Source{%q, %q}", source.Display, source.Text, test.display, test.text)
			}
		})
	}
}

func TestReadSourceInvalidUTF8LiteralCases(t *acceptancetesting.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "invalid byte", data: []byte{65, 255, 10}},
		{name: "overlong sequence", data: []byte{192, 175}},
		{name: "truncated sequence", data: []byte{226, 130}},
		{name: "encoded surrogate", data: []byte{237, 160, 128}},
		{name: "above Unicode maximum", data: []byte{244, 144, 128, 128}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *acceptancetesting.T) {
			_, err := ReadSource(&MemIO{Stdin: test.data}, "-")
			if err == nil || err.Error() != "<stdin>: input is not valid UTF-8" {
				t.Fatalf("got %v, want <stdin>: input is not valid UTF-8", err)
			}
		})
	}
}

func TestReadSourceMissingFile(t *acceptancetesting.T) {
	_, err := ReadSource(&MemIO{Files: map[string][]byte{}}, "nowhere.txt")
	if err == nil || err.Error() != "\"nowhere.txt\": cannot read the file" {
		t.Fatalf("got %v, want \"nowhere.txt\": cannot read the file", err)
	}
}

func TestMemIOLiteralReadsAndAppends(t *acceptancetesting.T) {
	memory := &MemIO{
		Stdin: []byte("in"),
		Files: map[string][]byte{"a": []byte("file")},
	}
	stdin, err := memory.ReadStdin()
	if err != nil || string(stdin) != "in" {
		t.Fatalf("got stdin %q, %v; want in, nil", stdin, err)
	}
	file, err := memory.ReadFile("a")
	if err != nil || string(file) != "file" {
		t.Fatalf("got file %q, %v; want file, nil", file, err)
	}
	if _, err := memory.ReadFile("absent"); err == nil {
		t.Fatal("ReadFile(absent) succeeded")
	}
	memory.Out("a")
	memory.Out("b")
	memory.Err("x")
	memory.Err("y")
	if memory.Stdout != "ab" || memory.Stderr != "xy" {
		t.Fatalf("got output %q and error %q, want ab and xy", memory.Stdout, memory.Stderr)
	}
}

func TestParseRepeatedOptionPrecedenceLiteralCases(t *acceptancetesting.T) {
	tests := []struct {
		name string
		argv []string
		want string
	}{
		{name: "flag duplicate with value", argv: []string{"--by-name", "--by-name=1"}, want: "repeated option --by-name"},
		{name: "value duplicate without value", argv: []string{"--now=A", "--now"}, want: "repeated option --now"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *acceptancetesting.T) {
			_, err := Parse(test.argv, []Spec{{Name: "--by-name"}, {Name: "--now", Takes: true}})
			if err == nil || err.Error() != test.want {
				t.Fatalf("got %v, want %q", err, test.want)
			}
		})
	}
}
