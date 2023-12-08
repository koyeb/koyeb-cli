package koyeb

import (
	"strings"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) ExtractFileSpec(ctx *CLIContext, target string) (*FileSpec, error) {
	switch i := strings.Index(target, ":"); i {
	case 0:
		return nil, &errors.CLIError{
			What:       "Error while copying",
			Why:        "Filespec must match the canonical format: [instance:]file/path",
			Additional: nil,
			Orig:       nil,
			Solution:   "If the problem persists, try to update the CLI to the latest version.",
		}
	case -1:
		return &FileSpec{
			FilePath: target,
		}, nil
	default:
		id, filePath := target[:i], target[i+1:]

		instanceID, err := h.ResolveInstanceArgs(ctx, id)
		if err != nil {
			return nil, err
		}

		return &FileSpec{
			InstanceID: instanceID,
			FilePath:   filePath,
		}, nil
	}
}

func (h *InstanceHandler) Cp(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	srcSpec, err := h.ExtractFileSpec(ctx, src)
	if err != nil {
		return err
	}
	dstSpec, err := h.ExtractFileSpec(ctx, dst)
	if err != nil {
		return err
	}

	manager, err := NewCopyManager(srcSpec, dstSpec)
	if err != nil {
		return err
	}

	log.Infof("Copying from %s to %s ...", src, dst)
	return manager.Copy(ctx)
}
