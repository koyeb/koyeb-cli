package renderer

import (
	"time"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func FormatTime(t time.Time) string {
	return t.Format(time.RFC822)
}

func FormatAppID(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.App().GetShortID(id)
		if err == nil {
			return sid
		}
	}
	return id
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

func FormatDomainID(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Domain().GetShortID(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatServiceID(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Service().GetShortID(id)
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

func FormatDeploymentID(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Deployment().GetShortID(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatRegionalDeploymentID(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.RegionalDeployment().GetShortID(id)
		if err == nil {
			return sid
		}
		panic(err)
	}
	return id
}

func FormatInstanceID(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Instance().GetShortID(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatSecretID(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Secret().GetShortID(id)
		if err == nil {
			return sid
		}
	}
	return id
}
