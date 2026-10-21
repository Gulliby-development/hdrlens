package cookies

import (
	"strings"
	"unicode/utf8"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

// Cookie is one parsed Set-Cookie field with the recognized attributes that
// were present and valid.
type Cookie struct {
	Line       int
	Name       string
	Value      string
	Expires    int64
	HasExpires bool
	MaxAge     int64
	HasMaxAge  bool
	Domain     string
	HasDomain  bool
	Path       string
	HasPath    bool
	Secure     bool
}

// ParseCookies parses every Set-Cookie field in file order.
func ParseCookies(fields message.Fields) ([]Cookie, error) {
	setCookies := fields.Named("Set-Cookie")
	cookies := make([]Cookie, 0, len(setCookies))
	for _, field := range setCookies {
		cookie, err := ParseCookie(field)
		if err != nil {
			return nil, err
		}
		cookies = append(cookies, cookie)
	}
	return cookies, nil
}

// ParseCookie parses a single Set-Cookie field.
func ParseCookie(field message.Field) (Cookie, error) {
	pieces := message.Split(field.Value, field.Col, ';', false)
	pair := pieces[0]
	equals := strings.IndexByte(pair.Text, '=')
	if equals < 0 {
		return Cookie{}, cookieProblem("HK001", "missing \"=\"", field.Line, pair.Col)
	}

	name := pair.Text[:equals]
	if name == "" || !validCookieName(name) {
		return Cookie{}, cookieProblem("HK002", "invalid cookie name "+base.Show(name), field.Line, pair.Col)
	}
	value := pair.Text[equals+1:]
	if !validCookieValue(value) {
		valueCol := pair.Col + utf16Columns(pair.Text[:equals+1])
		return Cookie{}, cookieProblem("HK003", "invalid cookie value "+base.Show(value), field.Line, valueCol)
	}

	cookie := Cookie{
		Line:  field.Line,
		Name:  name,
		Value: value,
	}
	for _, piece := range pieces[1:] {
		if piece.Text == "" {
			continue
		}
		attribute := parseAttribute(piece.Text)
		switch asciiLower(attribute.name) {
		case "expires":
			if expires, ok := message.ParseDate(attribute.value); ok {
				cookie.Expires = expires
				cookie.HasExpires = true
			}
		case "max-age":
			if age, ok := parseMaxAge(attribute.value); ok {
				cookie.MaxAge = age
				cookie.HasMaxAge = true
			}
		case "domain":
			domain := attribute.value
			if strings.HasPrefix(domain, ".") {
				domain = domain[1:]
			}
			if domain != "" {
				cookie.Domain = strings.ToLower(domain)
				cookie.HasDomain = true
			}
		case "path":
			if strings.HasPrefix(attribute.value, "/") {
				cookie.Path = attribute.value
				cookie.HasPath = true
			}
		case "secure":
			cookie.Secure = true
		}
	}
	return cookie, nil
}

type cookieAttribute struct {
	name  string
	value string
}

func parseAttribute(text string) cookieAttribute {
	equals := strings.IndexByte(text, '=')
	if equals < 0 {
		return cookieAttribute{name: trimCookieWhitespace(text)}
	}
	return cookieAttribute{
		name:  trimCookieWhitespace(text[:equals]),
		value: trimCookieWhitespace(text[equals+1:]),
	}
}

func trimCookieWhitespace(text string) string {
	start := 0
	end := len(text)
	for start < end && (text[start] == ' ' || text[start] == '\t') {
		start++
	}
	for end > start && (text[end-1] == ' ' || text[end-1] == '\t') {
		end--
	}
	return text[start:end]
}

func validCookieName(name string) bool {
	for index := 0; index < len(name); index++ {
		char := name[index]
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			strings.ContainsRune("!#$%&'*+-.^_`|~-", rune(char)) {
			continue
		}
		return false
	}
	return true
}

func validCookieValue(value string) bool {
	for len(value) > 0 {
		char, size := utf8.DecodeRuneInString(value)
		if size == 1 && char == utf8.RuneError {
			return false
		}
		if !(char == 0x21 ||
			(char >= 0x23 && char <= 0x2B) ||
			(char >= 0x2D && char <= 0x3A) ||
			(char >= 0x3C && char <= 0x5B) ||
			(char >= 0x5D && char <= 0x7E)) {
			return false
		}
		value = value[size:]
	}
	return true
}

func parseMaxAge(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}
	digits := value
	if digits[0] == '-' {
		digits = digits[1:]
	}
	if len(digits) < 1 || len(digits) > 10 {
		return 0, false
	}
	var amount int64
	for index := 0; index < len(digits); index++ {
		if digits[index] < '0' || digits[index] > '9' {
			return 0, false
		}
		amount = amount*10 + int64(digits[index]-'0')
	}
	if value[0] == '-' {
		amount = -amount
	}
	return amount, true
}

func cookieProblem(code, text string, line, column int) error {
	return &message.Problem{Code: code, Text: text, Line: line, Col: column}
}

func utf16Columns(text string) int {
	columns := 0
	for len(text) > 0 {
		char, size := utf8.DecodeRuneInString(text)
		if size < 1 {
			size = 1
		}
		if char > 0xFFFF {
			columns += 2
		} else {
			columns++
		}
		text = text[size:]
	}
	return columns
}

func asciiLower(text string) string {
	var result strings.Builder
	result.Grow(len(text))
	for index := 0; index < len(text); index++ {
		char := text[index]
		if char >= 'A' && char <= 'Z' {
			char += 'a' - 'A'
		}
		result.WriteByte(char)
	}
	return result.String()
}
