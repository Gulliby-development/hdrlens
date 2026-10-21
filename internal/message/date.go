package message

import (
	"strconv"
	"strings"
	"time"
)

var weekdayNames = [...]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
var fullWeekdayNames = [...]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
var monthNames = [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

const secondsPerDay int64 = 24 * 60 * 60

// ParseDate parses the fixed, RFC 850, or asctime English HTTP-date form.
func ParseDate(text string) (int64, bool) {
	if sec, ok := parseFixedDate(text); ok {
		return sec, true
	}
	if sec, ok := parseRFC850Date(text); ok {
		return sec, true
	}
	return parseAsctimeDate(text)
}

func parseFixedDate(text string) (int64, bool) {
	if len(text) != 29 ||
		text[3] != ',' || text[4] != ' ' || text[7] != ' ' ||
		text[11] != ' ' || text[16] != ' ' ||
		text[19] != ':' || text[22] != ':' ||
		text[25] != ' ' || text[26:] != "GMT" {
		return 0, false
	}

	weekday := nameIndex(text[0:3], weekdayNames[:])
	day, dayOK := parseTwoDigits(text[5:7])
	month := nameIndex(text[8:11], monthNames[:])
	year, yearOK := parseFourDigits(text[12:16])
	hour, hourOK := parseTwoDigits(text[17:19])
	minute, minuteOK := parseTwoDigits(text[20:22])
	second, secondOK := parseTwoDigits(text[23:25])
	if !dayOK || !yearOK || !hourOK || !minuteOK || !secondOK {
		return 0, false
	}
	return checkedDate(weekday, day, month, year, hour, minute, second)
}

func parseRFC850Date(text string) (int64, bool) {
	comma := strings.IndexByte(text, ',')
	if comma < 0 || comma+2 > len(text) {
		return 0, false
	}
	weekday := nameIndex(text[:comma], fullWeekdayNames[:])
	if weekday < 0 {
		return 0, false
	}
	start := comma + 2
	if len(text) != start+22 ||
		text[comma+1] != ' ' ||
		text[start+2] != '-' || text[start+6] != '-' ||
		text[start+9] != ' ' || text[start+18] != ' ' ||
		text[start+19:] != "GMT" ||
		text[start+12] != ':' || text[start+15] != ':' {
		return 0, false
	}

	day, dayOK := parseTwoDigits(text[start : start+2])
	month := nameIndex(text[start+3:start+6], monthNames[:])
	shortYear, yearOK := parseTwoDigits(text[start+7 : start+9])
	hour, hourOK := parseTwoDigits(text[start+10 : start+12])
	minute, minuteOK := parseTwoDigits(text[start+13 : start+15])
	second, secondOK := parseTwoDigits(text[start+16 : start+18])
	if !dayOK || !yearOK || !hourOK || !minuteOK || !secondOK {
		return 0, false
	}
	year := 2000 + shortYear
	if shortYear >= 70 {
		year = 1900 + shortYear
	}
	return checkedDate(weekday, day, month, year, hour, minute, second)
}

func parseAsctimeDate(text string) (int64, bool) {
	if len(text) != 24 ||
		text[3] != ' ' || text[7] != ' ' ||
		text[10] != ' ' || text[13] != ':' || text[16] != ':' ||
		text[19] != ' ' {
		return 0, false
	}

	weekday := nameIndex(text[0:3], weekdayNames[:])
	month := nameIndex(text[4:7], monthNames[:])
	day := 0
	if text[8] == ' ' {
		if !asciiDigit(text[9]) || text[9] == '0' {
			return 0, false
		}
		day = int(text[9] - '0')
	} else {
		var dayOK bool
		day, dayOK = parseTwoDigits(text[8:10])
		if !dayOK || day < 10 {
			return 0, false
		}
	}
	hour, hourOK := parseTwoDigits(text[11:13])
	minute, minuteOK := parseTwoDigits(text[14:16])
	second, secondOK := parseTwoDigits(text[17:19])
	year, yearOK := parseFourDigits(text[20:24])
	if !hourOK || !minuteOK || !secondOK || !yearOK {
		return 0, false
	}
	return checkedDate(weekday, day, month, year, hour, minute, second)
}

func checkedDate(weekday, day, month, year, hour, minute, second int) (int64, bool) {
	if weekday < 0 || month < 0 ||
		year < 1970 || year > 2999 || day < 1 ||
		hour > 23 || minute > 59 || second > 59 {
		return 0, false
	}

	date := time.Date(year, time.Month(month+1), day, hour, minute, second, 0, time.UTC)
	if date.Year() != year || int(date.Month()) != month+1 || date.Day() != day ||
		int(date.Weekday()+6)%7 != weekday {
		return 0, false
	}
	return date.Unix(), true
}

// FormatDate formats Unix seconds as an English GMT date. It also supports
// dates outside ParseDate's accepted year range.
func FormatDate(sec int64) string {
	days := sec / secondsPerDay
	secondsOfDay := sec % secondsPerDay
	if secondsOfDay < 0 {
		days--
		secondsOfDay += secondsPerDay
	}

	year, month, day := civilDate(days)
	weekday := (days + 3) % 7 // 1970-01-01 was Thursday.
	if weekday < 0 {
		weekday += 7
	}
	hour := secondsOfDay / 3600
	minute := secondsOfDay / 60 % 60
	second := secondsOfDay % 60

	var output strings.Builder
	output.Grow(29)
	output.WriteString(weekdayNames[weekday])
	output.WriteString(", ")
	output.WriteString(twoDigits(day))
	output.WriteByte(' ')
	output.WriteString(monthNames[month-1])
	output.WriteByte(' ')
	output.WriteString(formatYear(year))
	output.WriteByte(' ')
	output.WriteString(twoDigits(hour))
	output.WriteByte(':')
	output.WriteString(twoDigits(minute))
	output.WriteByte(':')
	output.WriteString(twoDigits(second))
	output.WriteString(" GMT")
	return output.String()
}

func nameIndex(text string, names []string) int {
	for i, name := range names {
		if text == name {
			return i
		}
	}
	return -1
}

func parseTwoDigits(text string) (int, bool) {
	if len(text) != 2 || !asciiDigit(text[0]) || !asciiDigit(text[1]) {
		return 0, false
	}
	return int(text[0]-'0')*10 + int(text[1]-'0'), true
}

func parseFourDigits(text string) (int, bool) {
	if len(text) != 4 {
		return 0, false
	}
	value := 0
	for i := 0; i < len(text); i++ {
		if !asciiDigit(text[i]) {
			return 0, false
		}
		value = value*10 + int(text[i]-'0')
	}
	return value, true
}

func asciiDigit(value byte) bool {
	return value >= '0' && value <= '9'
}

func twoDigits(value int64) string {
	return string([]byte{'0' + byte(value/10), '0' + byte(value%10)})
}

func formatYear(year int64) string {
	text := strconv.FormatInt(year, 10)
	if year < 0 {
		zeros := 4 - (len(text) - 1)
		if zeros > 0 {
			return "-" + strings.Repeat("0", zeros) + text[1:]
		}
		return text
	}
	if len(text) < 4 {
		return strings.Repeat("0", 4-len(text)) + text
	}
	return text
}

// civilDate converts days since 1970-01-01 to a proleptic Gregorian date.
func civilDate(days int64) (year, month, day int64) {
	z := days + 719468
	era := z / 146097
	if z < 0 && z%146097 != 0 {
		era--
	}
	dayOfEra := z - era*146097
	yearOfEra := (dayOfEra - dayOfEra/1460 + dayOfEra/36524 - dayOfEra/146096) / 365
	year = yearOfEra + era*400
	dayOfYear := dayOfEra - (365*yearOfEra + yearOfEra/4 - yearOfEra/100)
	monthPrime := (5*dayOfYear + 2) / 153
	day = dayOfYear - (153*monthPrime+2)/5 + 1
	if monthPrime < 10 {
		month = monthPrime + 3
	} else {
		month = monthPrime - 9
	}
	if month <= 2 {
		year++
	}
	return year, month, day
}
