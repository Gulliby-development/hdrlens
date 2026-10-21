package message

// BEGIN acceptance rows imports: host-written from the audited spec table. Edit acceptanceActual, not this block.
import (
	acceptancejson "encoding/json"
	acceptancereflect "reflect"
	acceptancetesting "testing"
)

// END acceptance rows imports

import (
	"errors"

	"github.com/Gulliby-development/hdrlens/internal/base"
)

// BEGIN acceptance rows: host-written from the audited spec table. Edit acceptanceActual, not this block.
func TestAcceptanceRowsMessage(t *acceptancetesting.T) {
	var rows []struct {
		ID       string `json:"id"`
		Input    any    `json:"input"`
		Expected any    `json:"expected"`
	}
	if err := acceptancejson.Unmarshal([]byte("[{\"id\":\"AC-001:1\",\"input\":{\"fields\":[{\"Col\":7,\"Line\":1,\"Name\":\"Host\",\"Value\":\"a\"},{\"Col\":7,\"Line\":3,\"Name\":\"HOST\",\"Value\":\"b\"},{\"Col\":4,\"Line\":4,\"Name\":\"X\",\"Value\":\"c\"}],\"name\":\"host\"},\"expected\":[{\"Col\":7,\"Line\":1,\"Name\":\"Host\",\"Value\":\"a\"},{\"Col\":7,\"Line\":3,\"Name\":\"HOST\",\"Value\":\"b\"}]},{\"id\":\"AC-001:2\",\"input\":{\"fields\":[{\"Col\":4,\"Line\":1,\"Name\":\"Ä\",\"Value\":\"a\"}],\"name\":\"ä\"},\"expected\":[]},{\"id\":\"AC-001:3\",\"input\":{\"fields\":[],\"name\":\"Host\"},\"expected\":[]},{\"id\":\"AC-002:1\",\"input\":{\"Code\":\"HL003\",\"Col\":7,\"Line\":1,\"Text\":\"missing colon\"},\"expected\":\"HL003 missing colon at 1:7\"},{\"id\":\"AC-002:2\",\"input\":{\"Code\":\"HL002\",\"Col\":1,\"Line\":4,\"Text\":\"continuation line\"},\"expected\":\"HL002 continuation line at 4:1\"},{\"id\":\"AC-003:1\",\"input\":[\"a, b ;c\",10,\";\",false],\"expected\":[{\"Col\":10,\"Text\":\"a, b\"},{\"Col\":16,\"Text\":\"c\"}]},{\"id\":\"AC-003:2\",\"input\":[\" ,x\",5,\",\",false],\"expected\":[{\"Col\":6,\"Text\":\"\"},{\"Col\":7,\"Text\":\"x\"}]},{\"id\":\"AC-003:3\",\"input\":[\"\",1,\",\",false],\"expected\":[{\"Col\":1,\"Text\":\"\"}]},{\"id\":\"AC-003:4\",\"input\":[\",,\",1,\",\",false],\"expected\":[{\"Col\":1,\"Text\":\"\"},{\"Col\":2,\"Text\":\"\"},{\"Col\":3,\"Text\":\"\"}]},{\"id\":\"AC-003:5\",\"input\":[\"a,\\\"b,c\\\",d\",1,\",\",true],\"expected\":[{\"Col\":1,\"Text\":\"a\"},{\"Col\":3,\"Text\":\"\\\"b,c\\\"\"},{\"Col\":9,\"Text\":\"d\"}]},{\"id\":\"AC-003:6\",\"input\":[\"a,\\\"b,c\\\",d\",1,\",\",false],\"expected\":[{\"Col\":1,\"Text\":\"a\"},{\"Col\":3,\"Text\":\"\\\"b\"},{\"Col\":6,\"Text\":\"c\\\"\"},{\"Col\":9,\"Text\":\"d\"}]},{\"id\":\"AC-003:7\",\"input\":[\"a,\\\"b,c\",1,\",\",true],\"expected\":[{\"Col\":1,\"Text\":\"a\"},{\"Col\":3,\"Text\":\"\\\"b,c\"}]},{\"id\":\"AC-003:8\",\"input\":[\"\\t a \\t, \\t\",4,\",\",false],\"expected\":[{\"Col\":6,\"Text\":\"a\"},{\"Col\":12,\"Text\":\"\"}]},{\"id\":\"AC-006:1\",\"input\":\"Thu, 01 Jan 1970 00:00:00 GMT\",\"expected\":0},{\"id\":\"AC-006:2\",\"input\":\"Sun, 06 Nov 1994 08:49:37 GMT\",\"expected\":784111777},{\"id\":\"AC-006:3\",\"input\":\"Tue, 29 Feb 2000 00:00:00 GMT\",\"expected\":951782400},{\"id\":\"AC-006:4\",\"input\":\"Mon, 06 Nov 1994 08:49:37 GMT\",\"expected\":null},{\"id\":\"AC-006:5\",\"input\":\"Sun, 06 nov 1994 08:49:37 GMT\",\"expected\":null},{\"id\":\"AC-006:6\",\"input\":\"Sun, 6 Nov 1994 08:49:37 GMT\",\"expected\":null},{\"id\":\"AC-006:7\",\"input\":\"Sun, 06 Nov 1994 24:00:00 GMT\",\"expected\":null},{\"id\":\"AC-006:8\",\"input\":\"Sun, 06 Nov 1994 08:60:00 GMT\",\"expected\":null},{\"id\":\"AC-006:9\",\"input\":\"Sun, 06 Nov 1994 08:49:60 GMT\",\"expected\":null},{\"id\":\"AC-006:10\",\"input\":\"Wed, 29 Feb 2100 00:00:00 GMT\",\"expected\":null},{\"id\":\"AC-006:11\",\"input\":\"Wed, 31 Dec 1969 23:59:59 GMT\",\"expected\":null},{\"id\":\"AC-006:12\",\"input\":\"Wed, 01 Jan 3000 00:00:00 GMT\",\"expected\":null},{\"id\":\"AC-006:13\",\"input\":\"Sun, 06 Nov 1994 08:49:37 UTC\",\"expected\":null},{\"id\":\"AC-006:14\",\"input\":\"Sun, 06 Nov 1994 08:49:37 GMT \",\"expected\":null},{\"id\":\"AC-007:1\",\"input\":0,\"expected\":\"Thu, 01 Jan 1970 00:00:00 GMT\"},{\"id\":\"AC-007:2\",\"input\":784111777,\"expected\":\"Sun, 06 Nov 1994 08:49:37 GMT\"},{\"id\":\"AC-007:3\",\"input\":951782400,\"expected\":\"Tue, 29 Feb 2000 00:00:00 GMT\"},{\"id\":\"AC-007:4\",\"input\":32503680000,\"expected\":\"Wed, 01 Jan 3000 00:00:00 GMT\"}]"), &rows); err != nil {
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

func acceptanceActual(input any) (any, error) {
	switch value := input.(type) {
	case []any:
		if len(value) != 4 {
			return nil, errors.New("Split requires four arguments")
		}
		text, textOK := value[0].(string)
		column, columnOK := value[1].(float64)
		separator, separatorOK := value[2].(string)
		quoted, quotedOK := value[3].(bool)
		if !textOK || !columnOK || !separatorOK || len(separator) != 1 || !quotedOK {
			return nil, errors.New("Split arguments have invalid types")
		}
		return Split(text, int(column), separator[0], quoted), nil
	case string:
		seconds, ok := ParseDate(value)
		if !ok {
			return nil, nil
		}
		return seconds, nil
	case float64:
		return FormatDate(int64(value)), nil
	case map[string]any:
		if rawFields, ok := value["fields"]; ok {
			fieldsInput, ok := rawFields.([]any)
			if !ok {
				return nil, errors.New("fields must be an array")
			}
			fields := make(Fields, 0, len(fieldsInput))
			for _, rawField := range fieldsInput {
				fieldInput, ok := rawField.(map[string]any)
				if !ok {
					return nil, errors.New("field must be an object")
				}
				name, nameOK := fieldInput["Name"].(string)
				fieldValue, valueOK := fieldInput["Value"].(string)
				line, lineOK := fieldInput["Line"].(float64)
				column, columnOK := fieldInput["Col"].(float64)
				if !nameOK || !valueOK || !lineOK || !columnOK {
					return nil, errors.New("field has invalid properties")
				}
				fields = append(fields, Field{
					Name: name, Value: fieldValue, Line: int(line), Col: int(column),
				})
			}
			name, ok := value["name"].(string)
			if !ok {
				return nil, errors.New("name must be a string")
			}
			return fields.Named(name), nil
		}
		code, codeOK := value["Code"].(string)
		text, textOK := value["Text"].(string)
		line, lineOK := value["Line"].(float64)
		column, columnOK := value["Col"].(float64)
		if !codeOK || !textOK || !lineOK || !columnOK {
			return nil, errors.New("Problem has invalid properties")
		}
		return (&Problem{
			Code: code, Text: text, Line: int(line), Col: int(column),
		}).Error(), nil
	default:
		return nil, errors.New("unsupported acceptance input")
	}
}

func TestLoadLiteralCases(t *acceptancetesting.T) {
	tests := []struct {
		name    string
		arg     string
		stdin   []byte
		files   map[string][]byte
		want    Loaded
		wantErr string
	}{
		{
			name: "named file",
			arg:  "a.txt",
			files: map[string][]byte{
				"a.txt": []byte("Host: a\n"),
			},
			want: Loaded{
				Display: "\"a.txt\"",
				Fields:  Fields{{Name: "Host", Value: "a", Line: 1, Col: 7}},
			},
		},
		{
			name:  "stdin",
			arg:   "-",
			stdin: []byte("A:b"),
			want: Loaded{
				Display: "<stdin>",
				Fields:  Fields{{Name: "A", Value: "b", Line: 1, Col: 3}},
			},
		},
		{
			name:    "missing file",
			arg:     "missing.txt",
			wantErr: "\"missing.txt\": cannot read the file",
		},
		{
			name:    "malformed UTF-8 stdin bytes",
			arg:     "-",
			stdin:   []byte{0x41, 0xFF, 0x0A},
			wantErr: "<stdin>: input is not valid UTF-8",
		},
		{
			name: "empty header",
			arg:  "empty.txt",
			files: map[string][]byte{
				"empty.txt": []byte("# nothing"),
			},
			wantErr: "\"empty.txt\": no header fields",
		},
		{
			name: "parser problem has one display prefix",
			arg:  "bad.txt",
			files: map[string][]byte{
				"bad.txt": []byte(" x:y"),
			},
			wantErr: "\"bad.txt\": HL002 continuation line at 1:1",
		},
		{
			name: "single BOM is removed by ReadSource",
			arg:  "bom.txt",
			files: map[string][]byte{
				"bom.txt": []byte{0xEF, 0xBB, 0xBF, 'A', ':', 'b'},
			},
			want: Loaded{
				Display: "\"bom.txt\"",
				Fields:  Fields{{Name: "A", Value: "b", Line: 1, Col: 3}},
			},
		},
		{
			name: "only one of two BOMs is removed",
			arg:  "double-bom.txt",
			files: map[string][]byte{
				"double-bom.txt": []byte{
					0xEF, 0xBB, 0xBF, 0xEF, 0xBB, 0xBF, 'A', ':', 'b',
				},
			},
			wantErr: "\"double-bom.txt\": HL001 unexpected character \"?\" at 1:1",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *acceptancetesting.T) {
			memory := &base.MemIO{Stdin: test.stdin, Files: test.files}
			got, err := Load(memory, test.arg)
			if test.wantErr != "" {
				if err == nil || err.Error() != test.wantErr {
					t.Fatalf("got %v, want %q", err, test.wantErr)
				}
				if got.Display != "" || got.Fields != nil {
					t.Fatalf("got partial Loaded value %#v on error", got)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if got.Display != test.want.Display || !acceptancereflect.DeepEqual(got.Fields, test.want.Fields) {
					t.Fatalf("got %#v, want %#v", got, test.want)
				}
			}
			if memory.Stdout != "" || memory.Stderr != "" {
				t.Fatalf("Load wrote output: stdout %q, stderr %q", memory.Stdout, memory.Stderr)
			}
		})
	}
}

func TestLoadedWrapLiteralCases(t *acceptancetesting.T) {
	loaded := Loaded{Display: "<stdin>"}
	if err := loaded.Wrap(nil); err != nil {
		t.Fatalf("Wrap(nil) = %v, want nil", err)
	}
	err := loaded.Wrap(errors.New("bad value"))
	if err == nil || err.Error() != "<stdin>: bad value" {
		t.Fatalf("Wrap(error) = %v, want <stdin>: bad value", err)
	}
}

func TestSplitUTF16AndNoEscapeQuotes(t *acceptancetesting.T) {
	tests := []struct {
		name   string
		text   string
		col    int
		sep    byte
		quoted bool
		want   []Piece
	}{
		{
			name:   "supplementary character counts as two columns",
			text:   "a😀,x",
			col:    1,
			sep:    ',',
			quoted: false,
			want:   []Piece{{Text: "a😀", Col: 1}, {Text: "x", Col: 5}},
		},
		{
			name:   "backslash does not escape quote",
			text:   `x,"a\",b`,
			col:    1,
			sep:    ',',
			quoted: true,
			want:   []Piece{{Text: "x", Col: 1}, {Text: "\"a\\\"", Col: 3}, {Text: "b", Col: 8}},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *acceptancetesting.T) {
			got := Split(test.text, test.col, test.sep, test.quoted)
			if !acceptancereflect.DeepEqual(got, test.want) {
				t.Fatalf("Split(%q) = %#v, want %#v", test.text, got, test.want)
			}
		})
	}
}

func TestParseDateYearBoundaries(t *acceptancetesting.T) {
	tests := []struct {
		text string
		ok   bool
	}{
		{text: "Thu, 01 Jan 1970 00:00:00 GMT", ok: true},
		{text: "Tue, 31 Dec 2999 23:59:59 GMT", ok: true},
		{text: "Wed, 01 Jan 3000 00:00:00 GMT", ok: false},
		{text: "Wed, 31 Dec 1969 23:59:59 GMT", ok: false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.text, func(t *acceptancetesting.T) {
			_, ok := ParseDate(test.text)
			if ok != test.ok {
				t.Fatalf("ParseDate(%q) valid = %v, want %v", test.text, ok, test.ok)
			}
		})
	}
}
