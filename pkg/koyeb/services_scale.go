package koyeb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Scale(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := h.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	scaleSpecs, _ := cmd.Flags().GetStringSlice("scale")
	instances, _ := cmd.Flags().GetInt64("instances")
	regions, _ := cmd.Flags().GetStringSlice("regions")

	var scalings []koyeb.ManualServiceScaling

	// Priority 1: Use --scale flags with region:instances format
	if len(scaleSpecs) > 0 {
		scalings, err = parseScaleSpecs(scaleSpecs)
		if err != nil {
			return err
		}
	} else {
		// Priority 2: Use --instances and --regions flags (legacy/simple mode)
		// Create separate scaling entries for each region to make patching easier
		if len(regions) > 0 {
			for _, region := range regions {
				if region != "" {
					var scope string
					if !hasPrefix(region, "region:") {
						scope = fmt.Sprintf("region:%s", region)
					} else {
						scope = region
					}
					manualScaling := koyeb.NewManualServiceScalingWithDefaults()
					manualScaling.SetInstances(instances)
					manualScaling.SetScopes([]string{scope})
					scalings = append(scalings, *manualScaling)
				}
			}
		} else {
			// No regions specified, apply globally
			manualScaling := koyeb.NewManualServiceScalingWithDefaults()
			manualScaling.SetInstances(instances)
			scalings = []koyeb.ManualServiceScaling{*manualScaling}
		}
	}

	// Create the UpdateServiceScalingRequest with scalings array
	updateScalingRequest := koyeb.NewUpdateServiceScalingRequestWithDefaults()
	updateScalingRequest.SetScalings(scalings)

	_, resp, err := ctx.Client.ServicesApi.UpdateServiceScaling(ctx.Context, service).Body(*updateScalingRequest).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while scaling the service `%s`", serviceName),
			err,
			resp,
		)
	}

	log.Infof("Service %s scale configuration updated.", serviceName)
	return nil
}

// parseScaleSpecs parses scale specifications in the format "region:instances" or "instances"
// Examples: "fra:3", "was:2", "3" (applies to all regions)
func parseScaleSpecs(specs []string) ([]koyeb.ManualServiceScaling, error) {
	var scalings []koyeb.ManualServiceScaling

	for _, spec := range specs {
		parts := strings.Split(spec, ":")

		var instances int64
		var scopes []string

		if len(parts) == 1 {
			// Format: "instances" - applies to all regions/no specific scope
			i, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid instances value '%s': %v", parts[0], err)
			}
			instances = i
		} else if len(parts) == 2 {
			// Format: "region:instances"
			region := parts[0]
			i, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid instances value in '%s': %v", spec, err)
			}
			instances = i

			// Add region: prefix if not present
			if !hasPrefix(region, "region:") {
				scopes = []string{fmt.Sprintf("region:%s", region)}
			} else {
				scopes = []string{region}
			}
		} else {
			return nil, fmt.Errorf("invalid scale format '%s': expected 'region:instances' or 'instances'", spec)
		}

		manualScaling := koyeb.NewManualServiceScalingWithDefaults()
		manualScaling.SetInstances(instances)
		if len(scopes) > 0 {
			manualScaling.SetScopes(scopes)
		}
		scalings = append(scalings, *manualScaling)
	}

	return scalings, nil
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
