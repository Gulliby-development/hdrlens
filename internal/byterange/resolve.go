package byterange

import "sort"

// Span is an inclusive interval of selected byte positions.
type Span struct {
	Start int64
	End   int64
}

// Resolution contains sorted, merged spans and their total byte count.
type Resolution struct {
	Spans []Span
	Total int64
}

// Resolve clips specifications to length, then sorts, merges and totals them.
func Resolve(parsed ParsedRanges, length int64) Resolution {
	result := Resolution{Spans: make([]Span, 0)}
	if !parsed.Present || length <= 0 {
		return result
	}

	lastByte := length - 1
	for _, spec := range parsed.Specs {
		var span Span
		switch spec.Kind {
		case ClosedRange:
			if spec.First < 0 || spec.First >= length || spec.Last < spec.First {
				continue
			}
			span = Span{Start: spec.First, End: spec.Last}
			if span.End > lastByte {
				span.End = lastByte
			}
		case OpenRange:
			if spec.First < 0 || spec.First >= length {
				continue
			}
			span = Span{Start: spec.First, End: lastByte}
		case SuffixRange:
			if spec.Suffix <= 0 {
				continue
			}
			start := int64(0)
			if spec.Suffix < length {
				start = length - spec.Suffix
			}
			span = Span{Start: start, End: lastByte}
		default:
			continue
		}
		result.Spans = append(result.Spans, span)
	}

	sort.Slice(result.Spans, func(i, j int) bool {
		if result.Spans[i].Start == result.Spans[j].Start {
			return result.Spans[i].End < result.Spans[j].End
		}
		return result.Spans[i].Start < result.Spans[j].Start
	})

	merged := result.Spans[:0]
	for _, span := range result.Spans {
		if len(merged) == 0 {
			merged = append(merged, span)
			continue
		}
		last := &merged[len(merged)-1]
		if span.Start <= last.End || (last.End < int64(^uint64(0)>>1) && span.Start == last.End+1) {
			if span.End > last.End {
				last.End = span.End
			}
			continue
		}
		merged = append(merged, span)
	}
	result.Spans = merged

	for _, span := range result.Spans {
		result.Total += span.End - span.Start + 1
	}
	return result
}
