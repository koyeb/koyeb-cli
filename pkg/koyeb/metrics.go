package koyeb

import (
	"time"

	"github.com/araddon/dateparse"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func NewMetricsCmd() *cobra.Command {
	h := NewMetricsHandler()

	metricsCmd := &cobra.Command{
		Use:     "metrics ACTION",
		Aliases: []string{"metric"},
		Short:   "Metrics",
	}

	getMetricsCmd := &cobra.Command{
		Use:   "get",
		Short: "Get metrics for a service or instance",
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			service := GetStringFlags(cmd, "service")
			instance := GetStringFlags(cmd, "instance")

			if service == "" && instance == "" {
				return &errors.CLIError{
					What:       "Error while fetching metrics",
					Why:        "you must specify --service or --instance",
					Additional: []string{},
					Orig:       nil,
					Solution:   "Add the missing flag and try again",
				}
			} else if service != "" && instance != "" {
				return &errors.CLIError{
					What:       "Error while fetching metrics",
					Why:        "you must specify --service or --instance, not both",
					Additional: []string{},
					Orig:       nil,
					Solution:   "Remove the extra flag and try again",
				}
			}

			var start *time.Time
			var end *time.Time

			if value := GetStringFlags(cmd, "start"); value != "" {
				parsed, err := dateparse.ParseStrict(value)
				if err != nil {
					return &errors.CLIError{
						What:       "Error while fetching metrics",
						Why:        "invalid date format for --start",
						Additional: []string{},
						Orig:       err,
						Solution:   "Fix the date format and try again",
					}
				}
				start = &parsed
			}
			if value := GetStringFlags(cmd, "end"); value != "" {
				parsed, err := dateparse.ParseStrict(value)
				if err != nil {
					return &errors.CLIError{
						What:       "Error while fetching metrics",
						Why:        "invalid date format for --end",
						Additional: []string{},
						Orig:       err,
						Solution:   "Fix the date format and try again",
					}
				}
				end = &parsed
			}

			if service != "" {
				return h.GetForService(ctx, cmd, service, start, end)
			}
			return h.GetForInstance(ctx, cmd, instance, start, end)
		}),
	}
	getMetricsCmd.Flags().String("service", "", "Service name or ID")
	getMetricsCmd.Flags().String("instance", "", "Instance name or ID")
	getMetricsCmd.Flags().String("start", "", "Start date for the metrics")
	getMetricsCmd.Flags().String("end", "", "End date for the metrics")

	metricsCmd.AddCommand(getMetricsCmd)

	return metricsCmd
}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

type MetricsHandler struct {
}

func (h *MetricsHandler) ResolveServiceArgs(ctx *CLIContext, val string) (string, error) {
	serviceMapper := ctx.Mapper.Service()
	id, err := serviceMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *MetricsHandler) ResolveInstanceArgs(ctx *CLIContext, val string) (string, error) {
	instanceMapper := ctx.Mapper.Instance()
	id, err := instanceMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}
