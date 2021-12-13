package renderer

import (
	"time"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper2"
)

func FormatTime(t time.Time) string {
	return t.Format(time.RFC822)
}

func FormatID(id string, full bool) string {
	if full {
		return id
	}
	return id[:4]
}

func FormatAppID(mapper *idmapper2.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.App().GetShortID(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatAppName(mapper *idmapper2.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.App().GetName(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatServiceID(mapper *idmapper2.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Service().GetShortID(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatServiceSlug(mapper *idmapper2.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Service().GetSlug(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatDeploymentID(mapper *idmapper2.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Deployment().GetShortID(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatSecretID(mapper *idmapper2.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Secret().GetShortID(id)
		if err == nil {
			return sid
		}
	}
	return id
}
