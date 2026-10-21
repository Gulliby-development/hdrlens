package negotiate

// BEGIN acceptance rows imports: host-written from the audited spec table. Edit acceptanceActual, not this block.
import (
	acceptancejson "encoding/json"
	acceptancereflect "reflect"
	acceptancetesting "testing"
)

// END acceptance rows imports

import (
	"errors"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

// BEGIN acceptance rows: host-written from the audited spec table. Edit acceptanceActual, not this block.
func TestAcceptanceRowsAccept(t *acceptancetesting.T) {
	var rows []struct {
		ID       string `json:"id"`
		Input    any    `json:"input"`
		Expected any    `json:"expected"`
	}
	if err := acceptancejson.Unmarshal([]byte("[{\"id\":\"AC-002:1\",\"input\":\"Accept: text/html, image\",\"expected\":\"HA001 invalid media range \\\"image\\\" at 1:20\"},{\"id\":\"AC-002:2\",\"input\":\"Accept: ;q=1\",\"expected\":\"HA001 invalid media range \\\"\\\" at 1:9\"},{\"id\":\"AC-002:3\",\"input\":\"Accept: text/html; q = 1\",\"expected\":\"HA002 invalid parameter \\\"q = 1\\\" at 1:20\"},{\"id\":\"AC-002:4\",\"input\":\"Accept: text/html;q=\",\"expected\":\"HA002 invalid parameter \\\"q=\\\" at 1:19\"},{\"id\":\"AC-002:5\",\"input\":\"Accept: text/html;q=1.5\",\"expected\":\"HA003 invalid q value \\\"1.5\\\" at 1:21\"},{\"id\":\"AC-002:6\",\"input\":\"Accept: text/html;q=1;Q=0\",\"expected\":\"HA004 repeated parameter \\\"Q\\\" at 1:23\"},{\"id\":\"AC-002:7\",\"input\":\"Accept:  \",\"expected\":\"HA005 no media range at 1:10\"},{\"id\":\"AC-002:8\",\"input\":\"Accept: text/html\\nAccept: x\",\"expected\":\"HA001 invalid media range \\\"x\\\" at 2:9\"},{\"id\":\"AC-002:9\",\"input\":\"Accept: */html\",\"expected\":\"HA001 invalid media range \\\"*/html\\\" at 1:9\"},{\"id\":\"AC-002:10\",\"input\":\"Accept: text/html;q=1;Q=1.5\",\"expected\":\"HA003 invalid q value \\\"1.5\\\" at 1:25\"},{\"id\":\"AC-002:11\",\"input\":\"Accept: text/html;x=a;X=b\",\"expected\":\"HA004 repeated parameter \\\"X\\\" at 1:23\"},{\"id\":\"AC-003:1\",\"input\":\"Accept: , Text/HTML;x=a+b;Q=0.85, , text/*;q=0., */*;q=1.000,\",\"expected\":[[\"text/html\",850],[\"text/*\",0],[\"*/*\",1000]]},{\"id\":\"AC-003:2\",\"input\":\"aCcEpT: application/a_b+c.d-e\\nAccept: text/plain;q=0.001\",\"expected\":[[\"application/a_b+c.d-e\",1000],[\"text/plain\",1]]},{\"id\":\"AC-004:1\",\"input\":1000,\"expected\":\"1\"},{\"id\":\"AC-004:2\",\"input\":0,\"expected\":\"0\"},{\"id\":\"AC-004:3\",\"input\":850,\"expected\":\"0.85\"},{\"id\":\"AC-004:4\",\"input\":900,\"expected\":\"0.9\"},{\"id\":\"AC-004:5\",\"input\":1,\"expected\":\"0.001\"},{\"id\":\"AC-004:6\",\"input\":10,\"expected\":\"0.01\"},{\"id\":\"AC-005:1\",\"input\":\"Application/A_b+c.d-e\",\"expected\":null},{\"id\":\"AC-005:2\",\"input\":\"text/*\",\"expected\":\"invalid offer \\\"text/*\\\"\"},{\"id\":\"AC-005:3\",\"input\":\"\",\"expected\":\"invalid offer \\\"\\\"\"},{\"id\":\"AC-005:4\",\"input\":\"-\",\"expected\":\"invalid offer \\\"-\\\"\"},{\"id\":\"AC-005:5\",\"input\":\"text/html;q=1\",\"expected\":\"invalid offer \\\"text/html;q=1\\\"\"},{\"id\":\"AC-005:6\",\"input\":\"text /html\",\"expected\":\"invalid offer \\\"text /html\\\"\"},{\"id\":\"AC-006:1\",\"input\":[\"Accept: text/*;q=0.5, text/html;q=0.2\",[\"text/html\",\"text/css\"]],\"expected\":{\"ratings\":[[200,\"text/html\"],[500,\"text/*\"]],\"selected\":\"text/css\"}},{\"id\":\"AC-006:2\",\"input\":[\"Accept: */*;q=1, text/*;q=0.5, text/html;q=0\\nAccept: TEXT/HTML;q=1\",[\"TEXT/HTML\",\"text/css\"]],\"expected\":{\"ratings\":[[0,\"text/html\"],[500,\"text/*\"]],\"selected\":\"text/css\"}},{\"id\":\"AC-006:3\",\"input\":[\"Accept: */*;q=0.1\",[\"image/png\",\"text/plain\"]],\"expected\":{\"ratings\":[[100,\"*/*\"],[100,\"*/*\"]],\"selected\":\"image/png\"}},{\"id\":\"AC-006:4\",\"input\":[\"Accept: text/html;q=0\",[\"text/html\",\"image/png\"]],\"expected\":{\"ratings\":[[0,\"text/html\"],[0,\"-\"]],\"selected\":null}},{\"id\":\"AC-006:5\",\"input\":[\"Host: a\",[\"a/b\",\"c/d\"]],\"expected\":{\"ratings\":[[1000,\"-\"],[1000,\"-\"]],\"selected\":\"a/b\"}}]"), &rows); err != nil {
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
	case float64:
		return QText(int(value)), nil
	case string:
		if !strings.Contains(value, ":") {
			if err := ValidateOffer(value); err != nil {
				return err.Error(), nil
			}
			return nil, nil
		}
		fields, err := message.Parse(value)
		if err != nil {
			return err.Error(), nil
		}
		ranges, err := ParseAccept(fields)
		if err != nil {
			return err.Error(), nil
		}
		projection := make([]any, 0, len(ranges))
		for _, mediaRange := range ranges {
			projection = append(projection, []any{mediaRange.Type + "/" + mediaRange.Subtype, mediaRange.Q})
		}
		return projection, nil
	case []any:
		if len(value) != 2 {
			return nil, errors.New("selection input requires a header and offers")
		}
		header, ok := value[0].(string)
		if !ok {
			return nil, errors.New("selection header must be a string")
		}
		rawOffers, ok := value[1].([]any)
		if !ok {
			return nil, errors.New("selection offers must be an array")
		}
		fields, err := message.Parse(header)
		if err != nil {
			return err.Error(), nil
		}
		ranges, err := ParseAccept(fields)
		if err != nil {
			return err.Error(), nil
		}
		offers := make([]string, 0, len(rawOffers))
		for _, rawOffer := range rawOffers {
			offer, ok := rawOffer.(string)
			if !ok {
				return nil, errors.New("selection offers must be strings")
			}
			offers = append(offers, offer)
		}
		ratings, selected, err := Pick(ranges, offers)
		if err != nil {
			return err.Error(), nil
		}
		projectedRatings := make([]any, 0, len(ratings))
		for _, rating := range ratings {
			projectedRatings = append(projectedRatings, []any{rating.Q, rating.Range})
		}
		var selectedOffer any
		if selected >= 0 {
			selectedOffer = offers[selected]
		}
		return map[string]any{"ratings": projectedRatings, "selected": selectedOffer}, nil
	default:
		return nil, errors.New("unsupported acceptance input")
	}
}

func TestAcceptAdditionalProblemFields(t *acceptancetesting.T) {
	tests := []struct {
		name      string
		header    string
		code      string
		text      string
		line      int
		col       int
		wantError string
	}{
		{name: "comma-only Accept", header: "Accept: , ,", code: "HA005", text: "no media range", line: 1, col: 9, wantError: "HA005 no media range at 1:9"},
		{name: "trailing empty parameter", header: "Accept: text/html;", code: "HA002", text: "invalid parameter \"\"", line: 1, col: 19, wantError: "HA002 invalid parameter \"\" at 1:19"},
		{name: "parameter syntax precedes duplicate detection", header: "Accept: text/html;Q=1;q = 0", code: "HA002", text: "invalid parameter \"q = 0\"", line: 1, col: 23, wantError: "HA002 invalid parameter \"q = 0\" at 1:23"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *acceptancetesting.T) {
			fields, err := message.Parse(test.header)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ParseAccept(fields)
			problem, ok := err.(*message.Problem)
			if !ok {
				t.Fatalf("ParseAccept error has type %T, want *message.Problem", err)
			}
			if problem.Code != test.code || problem.Text != test.text || problem.Line != test.line || problem.Col != test.col {
				t.Fatalf("problem = %#v, want %s %s at %d:%d", problem, test.code, test.text, test.line, test.col)
			}
			if err.Error() != test.wantError {
				t.Fatalf("error = %q, want %q", err.Error(), test.wantError)
			}
		})
	}
}

func TestAcceptQValueBoundaries(t *acceptancetesting.T) {
	tests := []struct {
		name      string
		header    string
		wantQ     int
		wantError string
	}{
		{name: "leading decimal is invalid", header: "Accept: text/html;q=.5", wantError: "HA003 invalid q value \".5\" at 1:21"},
		{name: "leading zero is invalid", header: "Accept: text/html;q=01", wantError: "HA003 invalid q value \"01\" at 1:21"},
		{name: "four fractional digits are invalid", header: "Accept: text/html;q=0.0000", wantError: "HA003 invalid q value \"0.0000\" at 1:21"},
		{name: "bare decimal point after one is valid", header: "Accept: text/html;q=1.", wantQ: 1000},
		{name: "one with nonzero fraction is invalid", header: "Accept: text/html;q=1.001", wantError: "HA003 invalid q value \"1.001\" at 1:21"},
		{name: "bare decimal point after zero is valid", header: "Accept: text/html;q=0.", wantQ: 0},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *acceptancetesting.T) {
			fields, err := message.Parse(test.header)
			if err != nil {
				t.Fatal(err)
			}
			ranges, err := ParseAccept(fields)
			if test.wantError != "" {
				if err == nil || err.Error() != test.wantError {
					t.Fatalf("ParseAccept error = %v, want %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(ranges) != 1 || ranges[0].Q != test.wantQ {
				t.Fatalf("ParseAccept ranges = %#v, want one range with q=%d", ranges, test.wantQ)
			}
		})
	}
}

func TestAcceptanceRunRegressionCases(t *acceptancetesting.T) {
	tests := []struct {
		name     string
		args     []string
		header   string
		wantCode int
		wantErr  string
		wantOut  string
	}{
		{
			name:     "invalid wildcard range",
			args:     []string{"a.txt", "text/html"},
			header:   "Accept: */html",
			wantCode: 0,
			wantErr:  "\"a.txt\": HA001 invalid media range \"*/html\" at 1:9",
		},
		{
			name:     "wildcard offer",
			args:     []string{"a.txt", "text/*"},
			header:   "Accept: text/html",
			wantCode: 0,
			wantErr:  "invalid offer \"text/*\"",
		},
		{
			name:     "first exact range wins across Accept fields",
			args:     []string{"a.txt", "TEXT/HTML", "text/css"},
			header:   "Accept: */*;q=1, text/*;q=0.5, text/html;q=0\nAccept: TEXT/HTML;q=1",
			wantCode: 0,
			wantOut:  "OFFER      Q    RANGE\nTEXT/HTML  0    text/html\ntext/css   0.5  text/*\nselected    text/css\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *acceptancetesting.T) {
			memory := &base.MemIO{Files: map[string][]byte{"a.txt": []byte(test.header)}}
			code, err := Run(test.args, memory)
			if code != test.wantCode {
				t.Fatalf("Run code = %d, want %d", code, test.wantCode)
			}
			if test.wantErr != "" {
				if err == nil || err.Error() != test.wantErr {
					t.Fatalf("Run error = %v, want %q", err, test.wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if memory.Stdout != test.wantOut {
				t.Fatalf("Run output = %q, want %q", memory.Stdout, test.wantOut)
			}
			if memory.Stderr != "" {
				t.Fatalf("Run wrote stderr %q", memory.Stderr)
			}
		})
	}
}
