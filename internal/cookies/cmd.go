package cookies

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

// Usage is the cookie command's usage line.
const Usage = "hdrlens cookie FILE URL --now DATE [--send URL]"

// Run validates and executes the cookie storage and send command.
func Run(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, []base.Spec{
		{Name: "--now", Takes: true},
		{Name: "--send", Takes: true},
	})
	if err != nil {
		return 0, err
	}
	nowText, hasNow := parsed.Values["--now"]
	if !hasNow {
		return 0, errors.New("missing --now")
	}
	if len(parsed.Args) == 0 {
		return 0, errors.New("missing FILE")
	}
	if len(parsed.Args) == 1 {
		return 0, errors.New("missing URL")
	}
	if len(parsed.Args) > 2 {
		return 0, errors.New("too many arguments")
	}

	loaded, err := message.Load(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	cookies, err := ParseCookies(loaded.Fields)
	if err != nil {
		return 0, loaded.Wrap(err)
	}
	if len(cookies) == 0 {
		return 0, loaded.Wrap(errors.New("no Set-Cookie field"))
	}

	response, err := ParseURL(parsed.Args[1])
	if err != nil {
		return 0, err
	}
	now, ok := message.ParseDate(nowText)
	if !ok {
		return 0, errors.New("invalid date " + base.Show(nowText))
	}

	var request URL
	hasSend := false
	if sendText, exists := parsed.Values["--send"]; exists {
		hasSend = true
		request, err = ParseURL(sendText)
		if err != nil {
			return 0, err
		}
	}

	stored := StoreCookies(cookies, response, now)
	rows := make([][]string, 0, len(stored))
	rejected := false
	for _, cookie := range stored {
		status := cookie.Status
		domain := cookie.Domain
		path := cookie.Path
		expires := "session"
		switch cookie.Status {
		case "rejected":
			status = "rejected (" + cookie.Reason + ")"
			domain = "-"
			path = "-"
			expires = "-"
			rejected = true
		case "expired":
			expires = "-"
		default:
			if cookie.HasExpiry {
				expires = message.FormatDate(cookie.Expires)
			}
		}
		rows = append(rows, []string{
			strconv.Itoa(cookie.Cookie.Line),
			cookie.Cookie.Name,
			status,
			DisplayPath(domain),
			DisplayPath(path),
			expires,
		})
	}

	lines := base.Table(
		[]string{"LINE", "NAME", "STATUS", "DOMAIN", "PATH", "EXPIRES"},
		rows,
	)
	if hasSend {
		lines = append(lines, base.Line("send", SendValue(SelectCookies(stored, request))))
	}
	io.Out(strings.Join(lines, "\n") + "\n")
	if rejected {
		return 1, nil
	}
	return 0, nil
}
