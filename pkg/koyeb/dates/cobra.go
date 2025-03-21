package dates

import (
	"time"

	"github.com/araddon/dateparse"
)

// HumanFriendlyDate implements the pflag.Value interface to allow parsing dates in a human-friendly format.
type HumanFriendlyDate struct {
	Time time.Time
}

func (t *HumanFriendlyDate) String() string {
	return t.Time.String()
}

func (t *HumanFriendlyDate) Set(value string) error {
	parsed, err := Parse(value)
	t.Time = parsed
	return err
}

func (t *HumanFriendlyDate) Type() string {
	return "HumanFriendlyDate"
}

func Parse(value string) (time.Time, error) {
	// Try to parse the date using RFC3339 format
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}

	// Fall back to the dateparse library. Use dateparse.ParseStrict instead of
	// dateparse.ParseAny to return an error in case of ambiguity, e.g.
	// "01/02/03" could be interpreted as "January 2, 2003" or "February 1, 2003".
	parsed, err := dateparse.ParseStrict(value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}
