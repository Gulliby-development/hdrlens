package negotiate

import (
	"strconv"
	"strings"
)

// Rating is the quality and deciding Accept range for an offer.
type Rating struct {
	Q     int
	Range string
}

// Pick validates offers, rates them, and returns the selected offer index.
func Pick(ranges []MediaRange, offers []string) ([]Rating, int, error) {
	for _, offer := range offers {
		if err := ValidateOffer(offer); err != nil {
			return nil, -1, err
		}
	}

	ratings := make([]Rating, 0, len(offers))
	selected := -1
	bestQ := 0
	for offerIndex, offer := range offers {
		rating := rateOffer(ranges, offer)
		ratings = append(ratings, rating)
		if rating.Q > bestQ {
			bestQ = rating.Q
			selected = offerIndex
		}
	}
	return ratings, selected, nil
}

func rateOffer(ranges []MediaRange, offer string) Rating {
	if len(ranges) == 0 {
		return Rating{Q: 1000, Range: "-"}
	}
	slash := strings.IndexByte(offer, '/')
	offerType := offer[:slash]
	offerSubtype := offer[slash+1:]

	bestSpecificity := 0
	var best Rating
	for _, mediaRange := range ranges {
		specificity := 0
		switch {
		case mediaRange.Type == "*" && mediaRange.Subtype == "*":
			specificity = 1
		case mediaRange.Subtype == "*" && asciiEqualFold(mediaRange.Type, offerType):
			specificity = 2
		case mediaRange.Subtype != "*" &&
			asciiEqualFold(mediaRange.Type, offerType) &&
			asciiEqualFold(mediaRange.Subtype, offerSubtype):
			specificity = 3
		}
		if specificity > bestSpecificity {
			bestSpecificity = specificity
			best = Rating{Q: mediaRange.Q, Range: asciiLower(mediaRange.Type) + "/" + asciiLower(mediaRange.Subtype)}
		}
	}
	if bestSpecificity == 0 {
		return Rating{Q: 0, Range: "-"}
	}
	return best
}

// QText formats an integer quality in thousandths.
func QText(q int) string {
	if q == 0 {
		return "0"
	}
	whole := q / 1000
	fraction := q % 1000
	if fraction == 0 {
		return strconv.Itoa(whole)
	}
	digits := strconv.Itoa(fraction)
	for len(digits) < 3 {
		digits = "0" + digits
	}
	digits = strings.TrimRight(digits, "0")
	return strconv.Itoa(whole) + "." + digits
}
