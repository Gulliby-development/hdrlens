package cookies

import (
	"errors"
	"sort"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
)

// URL is the parsed scheme, normalized host, and raw path of a cookie URL.
type URL struct {
	Scheme string
	Host   string
	Path   string
}

// ParseURL parses the manual http/https URL grammar used by the cookie command.
func ParseURL(text string) (URL, error) {
	var parsed URL
	remainder := ""
	switch {
	case strings.HasPrefix(text, "http://"):
		parsed.Scheme = "http"
		remainder = text[len("http://"):]
	case strings.HasPrefix(text, "https://"):
		parsed.Scheme = "https"
		remainder = text[len("https://"):]
	default:
		return URL{}, invalidURL(text)
	}

	authorityEnd := len(remainder)
	for index := 0; index < len(remainder); index++ {
		if remainder[index] == '/' || remainder[index] == '?' || remainder[index] == '#' {
			authorityEnd = index
			break
		}
	}
	authority := remainder[:authorityEnd]
	suffix := remainder[authorityEnd:]
	if authority == "" {
		return URL{}, invalidURL(text)
	}

	host := authority
	if colon := strings.IndexByte(authority, ':'); colon >= 0 {
		if strings.IndexByte(authority[colon+1:], ':') >= 0 {
			return URL{}, invalidURL(text)
		}
		host = authority[:colon]
		port := authority[colon+1:]
		if len(port) < 1 || len(port) > 5 || !allASCIIDigits(port) {
			return URL{}, invalidURL(text)
		}
	}
	if !validHost(host) {
		return URL{}, invalidURL(text)
	}

	parsed.Host = strings.ToLower(host)
	parsed.Path = "/"
	if len(suffix) > 0 && suffix[0] == '/' {
		pathEnd := len(suffix)
		for index := 0; index < len(suffix); index++ {
			if suffix[index] == '?' || suffix[index] == '#' {
				pathEnd = index
				break
			}
		}
		parsed.Path = suffix[:pathEnd]
	}
	return parsed, nil
}

// DefaultPath returns the cookie default path for a parsed request path.
func DefaultPath(path string) string {
	if path == "" {
		return "/"
	}
	lastSlash := strings.LastIndexByte(path, '/')
	if lastSlash <= 0 {
		return "/"
	}
	return path[:lastSlash]
}

// StoredCookie combines the parsed cookie and its storage decision.
type StoredCookie struct {
	Cookie    Cookie
	Status    string
	Reason    string
	Domain    string
	Path      string
	Expires   int64
	HasExpiry bool
}

// StoreCookies applies response-host rejection and expiry rules in file order.
func StoreCookies(cookies []Cookie, response URL, now int64) []StoredCookie {
	stored := make([]StoredCookie, 0, len(cookies))
	for _, cookie := range cookies {
		result := StoredCookie{Cookie: cookie, Status: "stored"}
		if cookie.HasDomain {
			switch {
			case !strings.Contains(cookie.Domain, "."):
				result.Status = "rejected"
				result.Reason = "domain has only one label"
			case cookie.Domain != response.Host && !strings.HasSuffix(response.Host, "."+cookie.Domain):
				result.Status = "rejected"
				result.Reason = "domain does not match the URL host"
			}
		}
		if result.Status != "rejected" && cookie.Secure && response.Scheme == "http" {
			result.Status = "rejected"
			result.Reason = "Secure cookie from an http URL"
		}
		if result.Status == "rejected" {
			stored = append(stored, result)
			continue
		}

		result.Domain = response.Host
		if cookie.HasDomain {
			result.Domain = cookie.Domain
		}
		result.Path = DefaultPath(response.Path)
		if cookie.HasPath {
			result.Path = cookie.Path
		}

		switch {
		case cookie.HasMaxAge && cookie.MaxAge <= 0:
			result.Status = "expired"
		case cookie.HasMaxAge:
			result.Expires = addSaturating(now, cookie.MaxAge)
			result.HasExpiry = true
		case cookie.HasExpires && cookie.Expires <= now:
			result.Status = "expired"
		case cookie.HasExpires:
			result.Expires = cookie.Expires
			result.HasExpiry = true
		}
		stored = append(stored, result)
	}
	return stored
}

// SelectCookies returns matching stored cookies, longest path first and stable
// in file order for cookies whose paths have equal length.
func SelectCookies(cookies []StoredCookie, request URL) []StoredCookie {
	selected := make([]StoredCookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie.Status != "stored" {
			continue
		}
		if cookie.Cookie.HasDomain {
			if request.Host != cookie.Domain && !strings.HasSuffix(request.Host, "."+cookie.Domain) {
				continue
			}
		} else if request.Host != cookie.Domain {
			continue
		}
		if !cookiePathMatches(cookie.Path, request.Path) {
			continue
		}
		if cookie.Cookie.Secure && request.Scheme != "https" {
			continue
		}
		selected = append(selected, cookie)
	}
	sort.SliceStable(selected, func(left, right int) bool {
		return len(selected[left].Path) > len(selected[right].Path)
	})
	return selected
}

// DisplayPath makes a raw path safe for an ASCII-only table cell without
// changing the path used by storage or matching.
func DisplayPath(path string) string {
	shown := base.Show(path)
	return shown[1 : len(shown)-1]
}

// SendValue joins selected cookie name/value pairs for a send row.
func SendValue(cookies []StoredCookie) string {
	if len(cookies) == 0 {
		return "-"
	}
	parts := make([]string, len(cookies))
	for index, cookie := range cookies {
		parts[index] = cookie.Cookie.Name + "=" + cookie.Cookie.Value
	}
	return strings.Join(parts, "; ")
}

func cookiePathMatches(cookiePath, requestPath string) bool {
	if cookiePath == requestPath {
		return true
	}
	if !strings.HasPrefix(requestPath, cookiePath) {
		return false
	}
	if strings.HasSuffix(cookiePath, "/") {
		return true
	}
	return len(requestPath) > len(cookiePath) && requestPath[len(cookiePath)] == '/'
}

func validHost(host string) bool {
	if host == "" {
		return false
	}
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if label == "" || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for index := 0; index < len(label); index++ {
			char := label[index]
			if (char >= 'A' && char <= 'Z') ||
				(char >= 'a' && char <= 'z') ||
				(char >= '0' && char <= '9') ||
				char == '-' {
				continue
			}
			return false
		}
	}
	return true
}

func allASCIIDigits(text string) bool {
	for index := 0; index < len(text); index++ {
		if text[index] < '0' || text[index] > '9' {
			return false
		}
	}
	return true
}

func addSaturating(value, amount int64) int64 {
	maxInt64 := int64(^uint64(0) >> 1)
	if amount > 0 && value > maxInt64-amount {
		return maxInt64
	}
	return value + amount
}

func invalidURL(text string) error {
	return errors.New("invalid URL " + base.Show(text))
}
