package negotiate

import (
	"errors"
	"strings"

	"github.com/Gulliby-development/hdrlens/internal/base"
	"github.com/Gulliby-development/hdrlens/internal/message"
)

const Usage = "hdrlens accept FILE OFFER..."

func Run(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return 0, err
	}
	if len(parsed.Args) == 0 {
		return 0, errors.New("missing FILE")
	}
	if len(parsed.Args) == 1 {
		return 0, errors.New("missing OFFER")
	}

	loaded, err := message.Load(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	ranges, err := ParseAccept(loaded.Fields)
	if err != nil {
		return 0, loaded.Wrap(err)
	}

	offers := parsed.Args[1:]
	ratings, selected, err := Pick(ranges, offers)
	if err != nil {
		return 0, err
	}

	rows := make([][]string, 0, len(offers))
	for index, offer := range offers {
		rows = append(rows, []string{offer, QText(ratings[index].Q), ratings[index].Range})
	}
	lines := base.Table([]string{"OFFER", "Q", "RANGE"}, rows)
	selectedOffer := "none"
	exitCode := 1
	if selected >= 0 {
		selectedOffer = offers[selected]
		exitCode = 0
	}
	lines = append(lines, base.Line("selected", selectedOffer))
	io.Out(strings.Join(lines, "\n") + "\n")
	return exitCode, nil
}
