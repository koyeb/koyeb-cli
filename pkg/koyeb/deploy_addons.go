package koyeb

import (
	"fmt"
	"path/filepath"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
)

type addonsHandler struct {
	addons []Addon
}

func NewAddonsHandler(addons []string) (*addonsHandler, error) {
	handler := &addonsHandler{}

	for _, addon := range addons {
		log.Infof("Registering addon: %s", addon)
		err := handler.RegisterAddon(addon)
		if err != nil {
			return nil, err
		}
	}

	return handler, nil
}

type Addon interface {
	Setup(ctx *CLIContext, dir string) error
	PreDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error
	PostDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error
	Cleanup(ctx *CLIContext) error
}

func (h *addonsHandler) RegisterAddon(name string) error {
	switch name {
	case "cog":
		h.addons = append(h.addons, &cogAddon{})
	default:
		return fmt.Errorf("unknown addon: %s", name)
	}
	return nil
}

func (h *addonsHandler) PreDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error {
	for _, addon := range h.addons {
		log.Debugf("Running addon pre-deploy")
		if err := addon.PreDeploy(ctx, definition); err != nil {
			return err
		}
	}
	return nil
}

func (h *addonsHandler) PostDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error {
	for _, addon := range h.addons {
		log.Debugf("Running addon post-deploy")
		if err := addon.PostDeploy(ctx, definition); err != nil {
			return err
		}
	}
	return nil
}

func (h *addonsHandler) Setup(ctx *CLIContext, dir string) error {
	basePath, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	for _, addon := range h.addons {
		log.Debugf("Running addon setup in dir %s", dir)
		if err := addon.Setup(ctx, basePath); err != nil {
			return err
		}
	}
	return nil
}

func (h *addonsHandler) Cleanup(ctx *CLIContext) error {
	for _, addon := range h.addons {
		log.Debugf("Running addon cleanup")
		if err := addon.Cleanup(ctx); err != nil {
			return err
		}
	}
	return nil
}
