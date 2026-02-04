package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) UpdateScale(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := h.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	// First, get the current scaling configuration
	currentRes, resp, err := ctx.Client.ServicesApi.GetServiceScaling(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving current scale configuration for service `%s`", serviceName),
			err,
			resp,
		)
	}

	// Get existing scalings
	existingScalings := make(map[string]*koyeb.ManualServiceScaling)
	for _, scaling := range currentRes.GetScalings() {
		key := buildScalingKey(scaling.GetScopes())
		existingScalings[key] = &scaling
	}

	scaleSpecs, _ := cmd.Flags().GetStringSlice("scale")
	instances, _ := cmd.Flags().GetInt64("instances")
	regions, _ := cmd.Flags().GetStringSlice("regions")

	var newScalings []koyeb.ManualServiceScaling
	var removals []string

	// Priority 1: Use --scale flags with region:instances format or !region for removal
	if len(scaleSpecs) > 0 {
		for _, spec := range scaleSpecs {
			// Check if this is a removal spec (starts with !)
			if len(spec) > 0 && spec[0] == '!' {
				region := spec[1:]
				if region == "" {
					// Remove global scaling (empty scopes)
					removals = append(removals, "__global__")
				} else if !hasPrefix(region, "region:") {
					removals = append(removals, fmt.Sprintf("region:%s", region))
				} else {
					removals = append(removals, region)
				}
			} else {
				// Parse as regular scaling spec
				scalings, err := parseScaleSpecs([]string{spec})
				if err != nil {
					return err
				}
				newScalings = append(newScalings, scalings...)
			}
		}
	} else {
		// Priority 2: Use --instances and --regions flags (legacy/simple mode)
		// Create separate scaling entries for each region to make patching easier
		if len(regions) > 0 {
			for _, region := range regions {
				if region != "" {
					// Check if this is a removal spec (starts with !)
					if region[0] == '!' {
						removeRegion := region[1:]
						var scope string
						if !hasPrefix(removeRegion, "region:") {
							scope = fmt.Sprintf("region:%s", removeRegion)
						} else {
							scope = removeRegion
						}
						removals = append(removals, scope)
					} else {
						// Add new scaling for this region
						var scope string
						if !hasPrefix(region, "region:") {
							scope = fmt.Sprintf("region:%s", region)
						} else {
							scope = region
						}
						manualScaling := koyeb.NewManualServiceScalingWithDefaults()
						manualScaling.SetInstances(instances)
						manualScaling.SetScopes([]string{scope})
						newScalings = append(newScalings, *manualScaling)
					}
				}
			}
		} else {
			// No regions specified, apply globally
			manualScaling := koyeb.NewManualServiceScalingWithDefaults()
			manualScaling.SetInstances(instances)
			newScalings = []koyeb.ManualServiceScaling{*manualScaling}
		}
	}

	// Remove specified regions
	for _, removal := range removals {
		key := buildScalingKey([]string{removal})
		delete(existingScalings, key)
	}

	// Merge new scalings with existing ones (PATCH behavior)
	for _, newScaling := range newScalings {
		key := buildScalingKey(newScaling.GetScopes())
		existingScalings[key] = &newScaling
	}

	// Build final scalings array from the merged map
	// Sort so that entries with empty scopes (global) come last
	var finalScalings []koyeb.ManualServiceScaling
	var globalScalings []koyeb.ManualServiceScaling

	for _, scaling := range existingScalings {
		if len(scaling.GetScopes()) == 0 {
			globalScalings = append(globalScalings, *scaling)
		} else {
			finalScalings = append(finalScalings, *scaling)
		}
	}

	// Append global scalings at the end
	finalScalings = append(finalScalings, globalScalings...)

	// Create the UpdateServiceScalingRequest with merged scalings array
	updateScalingRequest := koyeb.NewUpdateServiceScalingRequestWithDefaults()
	updateScalingRequest.SetScalings(finalScalings)

	_, resp, err = ctx.Client.ServicesApi.UpdateServiceScaling(ctx.Context, service).Body(*updateScalingRequest).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the service `%s` scale configuration", serviceName),
			err,
			resp,
		)
	}

	log.Infof("Service %s scale configuration updated.", serviceName)
	return nil
}

// buildScalingKey creates a unique key for a scaling configuration based on its scopes
// Empty scopes means global scaling
func buildScalingKey(scopes []string) string {
	if len(scopes) == 0 {
		return "__global__"
	}
	// Sort and join scopes to create a consistent key
	// For simplicity, we'll use the first scope as the key since typically there's one scope per scaling
	return scopes[0]
}
