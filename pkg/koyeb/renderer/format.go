package renderer

import "time"

func FormatTime(t time.Time) string {
	return t.Format(time.RFC822)
}

func FormatID(id string, full bool) string {
	if full {
		return id
	}
	return id[:8]
}
