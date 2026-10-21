package negotiate

import (
	"errors"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

// MediaRange is one normalized Accept media range and its quality in thousandths.
type MediaRange struct {
	Type, Subtype string
	Q             int
}

// ParseAccept validates and normalizes all Accept fields in file order.
func ParseAccept(fields message.Fields) ([]MediaRange, error) {
	ranges := make([]MediaRange, 0)
	for _, field := range fields.Named("Accept") {
		elements := message.Split(field.Value, field.Col, ',', false)
		hasElement := false
		for _, element := range elements {
			if element.Text != "" {
				hasElement = true
				break
			}
		}
		if !hasElement {
			return nil, acceptProblem(field, "HA005", "no media range", field.Col)
		}

		for _, element := range elements {
			if element.Text == "" {
				continue
			}
			parts := message.Split(element.Text, element.Col, ';', false)
			rangeText := parts[0].Text
			mediaType, mediaSubtype, ok := parseMediaRange(rangeText)
			if !ok {
				return nil, acceptProblem(field, "HA001", "invalid media range "+base.Show(rangeText), parts[0].Col)
			}

			q := 1000
			seen := make(map[string]bool, len(parts)-1)
			for _, parameter := range parts[1:] {
				name, value, valueOffset, valid := parseParameter(parameter.Text)
				if !valid {
					return nil, acceptProblem(field, "HA002", "invalid parameter "+base.Show(parameter.Text), parameter.Col)
				}
				lowerName := asciiLower(name)
				if lowerName == "q" {
					parsedQ, validQ := parseQ(value)
					if !validQ {
						return nil, acceptProblem(field, "HA003", "invalid q value "+base.Show(value), parameter.Col+valueOffset)
					}
					q = parsedQ
				}
				if seen[lowerName] {
					return nil, acceptProblem(field, "HA004", "repeated parameter "+base.Show(name), parameter.Col)
				}
				seen[lowerName] = true
			}
			ranges = append(ranges, MediaRange{Type: mediaType, Subtype: mediaSubtype, Q: q})
		}
	}
	return ranges, nil
}

func acceptProblem(field message.Field, code, text string, col int) error {
	return &message.Problem{Code: code, Text: text, Line: field.Line, Col: col}
}

func parseMediaRange(value string) (string, string, bool) {
	slash := strings.IndexByte(value, '/')
	if slash <= 0 || slash == len(value)-1 || strings.IndexByte(value[slash+1:], '/') >= 0 {
		return "", "", false
	}
	mediaType := value[:slash]
	subtype := value[slash+1:]
	if mediaType == "*" && subtype == "*" {
		return "*", "*", true
	}
	if mediaType == "*" || !validMediaToken(mediaType) {
		return "", "", false
	}
	if subtype == "*" {
		return asciiLower(mediaType), "*", true
	}
	if !validMediaToken(subtype) {
		return "", "", false
	}
	return asciiLower(mediaType), asciiLower(subtype), true
}

func validMediaToken(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		char := value[index]
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '_' || char == '+' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func parseParameter(value string) (string, string, int, bool) {
	equal := strings.IndexByte(value, '=')
	if equal <= 0 || equal == len(value)-1 {
		return "", "", 0, false
	}
	name := value[:equal]
	parameterValue := value[equal+1:]
	if !validParameterName(name) || !validParameterValue(parameterValue) {
		return "", "", 0, false
	}
	return name, parameterValue, equal + 1, true
}

func validParameterName(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		char := value[index]
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '_' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func validParameterValue(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		char := value[index]
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '_' || char == '+' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func parseQ(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	if value[0] != '0' && value[0] != '1' {
		return 0, false
	}
	quality := 0
	if value[0] == '1' {
		quality = 1000
	}
	if len(value) == 1 {
		return quality, true
	}
	if value[1] != '.' || len(value) > 5 {
		return 0, false
	}
	fractionDigits := 0
	fraction := 0
	for index := 2; index < len(value); index++ {
		char := value[index]
		if char < '0' || char > '9' {
			return 0, false
		}
		if value[0] == '1' && char != '0' {
			return 0, false
		}
		fraction = fraction*10 + int(char-'0')
		fractionDigits++
	}
	for fractionDigits < 3 {
		fraction *= 10
		fractionDigits++
	}
	if value[0] == '0' {
		return fraction, true
	}
	return quality, true
}

// ValidateOffer checks that value is a concrete TYPE/SUB offer.
func ValidateOffer(value string) error {
	mediaType, subtype, ok := parseMediaRange(value)
	if !ok || mediaType == "*" || subtype == "*" {
		return errors.New("invalid offer " + base.Show(value))
	}
	return nil
}

func asciiLower(value string) string {
	bytes := []byte(value)
	for index, char := range bytes {
		if char >= 'A' && char <= 'Z' {
			bytes[index] = char + ('a' - 'A')
		}
	}
	return string(bytes)
}

func asciiEqualFold(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := 0; index < len(left); index++ {
		leftChar, rightChar := left[index], right[index]
		if leftChar >= 'A' && leftChar <= 'Z' {
			leftChar += 'a' - 'A'
		}
		if rightChar >= 'A' && rightChar <= 'Z' {
			rightChar += 'a' - 'A'
		}
		if leftChar != rightChar {
			return false
		}
	}
	return true
}
