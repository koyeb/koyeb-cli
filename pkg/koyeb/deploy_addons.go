package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
)

type addonsHandler struct {
	addons []Addon
}

func NewAddonsHandler(addons []string) (*addonsHandler, error) {
	handler := &addonsHandler{}

	for _, addon := range addons {
		err := handler.RegisterAddon(addon)
		if err != nil {
			return nil, err
		}
	}

	return handler, nil
}

type Addon interface {
	Setup(ctx *CLIContext) error
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
		log.Debugf("Running pre-deploy for addon %s", addon)
		if err := addon.PreDeploy(ctx, definition); err != nil {
			return err
		}
	}
	return nil
}

func (h *addonsHandler) PostDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error {
	for _, addon := range h.addons {
		log.Debugf("Running post-deploy for addon %s", addon)
		if err := addon.PostDeploy(ctx, definition); err != nil {
			return err
		}
	}
	return nil
}

func (h *addonsHandler) Setup(ctx *CLIContext) error {
	for _, addon := range h.addons {
		log.Debugf("Running setup for addon %s", addon)
		if err := addon.Setup(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (h *addonsHandler) Cleanup(ctx *CLIContext) error {
	for _, addon := range h.addons {
		log.Debugf("Running cleanup for addon %s", addon)
		if err := addon.Cleanup(ctx); err != nil {
			return err
		}
	}
	return nil
}
