package renderer

import (
	"time"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func FormatTime(t time.Time) string {
	return t.Format(time.RFC822)
}

func FormatAppName(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.App().GetName(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatServiceSlug(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Service().GetSlug(id)
		if err == nil {
			return sid
		}
	}
	return id
}

// FormatID formats the ID to be displayed in the CLI. If full is false, only the first 8 characters are displayed.
func FormatID(fullId string, full bool) string {
	if full {
		return fullId
	}
	return fullId[:8]
}
