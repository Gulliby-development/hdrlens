package freshness

// Evaluation is the computed cache lifetime, current age, and resulting verdict.
type Evaluation struct {
	Lifetime int64
	Source   string
	Age      int64
	Verdict  string
}

// Evaluate calculates freshness from validated control metadata and the supplied
// current time. control must be the result of ParseControl and contain a valid
// Date; now is the parsed current time in Unix seconds. The calculation also
// saturates arithmetic and clamps negative age/lifetime values, so malformed
// direct calls cannot overflow int64.
func Evaluate(control Control, now int64, shared bool) Evaluation {
	lifetime, source := lifetimeFor(control, shared)
	storedAge := control.Age
	if storedAge < 0 {
		storedAge = 0
	}
	age := saturatingNonnegativeAdd(storedAge, nonnegativeDifference(now, control.Date))

	verdict := "stale"
	switch {
	case hasDirective(control, "no-store") || (shared && hasDirective(control, "private")):
		verdict = "uncacheable"
	case hasDirective(control, "no-cache"):
		verdict = "revalidate"
	case lifetime > age:
		verdict = "fresh"
	}

	return Evaluation{
		Lifetime: lifetime,
		Source:   source,
		Age:      age,
		Verdict:  verdict,
	}
}

func lifetimeFor(control Control, shared bool) (int64, string) {
	if shared {
		if directive, ok := control.Directives["s-maxage"]; ok {
			return nonnegative(directive.DeltaSeconds), "s-maxage"
		}
	}
	if directive, ok := control.Directives["max-age"]; ok {
		return nonnegative(directive.DeltaSeconds), "max-age"
	}
	if control.HasExpires {
		if control.ExpiresValid {
			return nonnegativeDifference(control.Expires, control.Date), "expires"
		}
		return 0, "expires"
	}
	if control.HasLastModified {
		return nonnegativeDifference(control.Date, control.LastModified) / 10, "heuristic"
	}
	return 0, "none"
}

func hasDirective(control Control, name string) bool {
	_, ok := control.Directives[name]
	return ok
}

func nonnegative(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

// nonnegativeDifference returns max(0, later-earlier) without signed overflow.
func nonnegativeDifference(later, earlier int64) int64 {
	if later <= earlier {
		return 0
	}
	if earlier < 0 && later > maxInt64+earlier {
		return maxInt64
	}
	return later - earlier
}

// saturatingNonnegativeAdd returns a+b, capped at MaxInt64.
func saturatingNonnegativeAdd(a, b int64) int64 {
	a = nonnegative(a)
	b = nonnegative(b)
	if a > maxInt64-b {
		return maxInt64
	}
	return a + b
}

const maxInt64 = int64(1<<63 - 1)
