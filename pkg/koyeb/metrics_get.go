package koyeb

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *MetricsHandler) GetForInstance(ctx *CLIContext, cmd *cobra.Command, instance string, start *time.Time, end *time.Time) error {
	instanceId, err := h.ResolveInstanceArgs(ctx, instance)
	if err != nil {
		return err
	}
	return h.get(ctx, cmd, [][]string{
		{"CPU usage", "CPU_TOTAL_PERCENT"},
		{"Memory usage", "MEM_RSS"},
	}, "", instanceId, start, end)
}

func (h *MetricsHandler) GetForService(ctx *CLIContext, cmd *cobra.Command, service string, start *time.Time, end *time.Time) error {
	serviceId, err := h.ResolveServiceArgs(ctx, service)
	if err != nil {
		return err
	}
	return h.get(ctx, cmd, [][]string{
		{"CPU usage", "CPU_TOTAL_PERCENT"},
		{"Memory usage", "MEM_RSS"},
		{"Requests throughput", "HTTP_THROUGHPUT"},
		{"Response time (50%)", "HTTP_RESPONSE_TIME_50P"},
		{"Response time (90%)", "HTTP_RESPONSE_TIME_90P"},
		{"Response time (99%)", "HTTP_RESPONSE_TIME_99P"},
		{"Response time (max)", "HTTP_RESPONSE_TIME_MAX"},
		{"Public data transfer input", "PUBLIC_DATA_TRANSFER_IN"},
		{"Public data transfer output", "PUBLIC_DATA_TRANSFER_OUT"},
	}, serviceId, "", start, end)
}

// Implementation of the `get` command. Do not call this function directly, use `GetForInstance` or `GetForService` instead.
func (h *MetricsHandler) get(ctx *CLIContext, cmd *cobra.Command, metrics [][]string, serviceId string, instanceId string, start *time.Time, end *time.Time) error {
	full := GetBoolFlags(cmd, "full")
	renderer := renderer.NewChainRenderer(ctx.Renderer)

	for _, e := range metrics {
		metricHumanName, metricName := e[0], e[1]

		req := ctx.Client.MetricsApi.GetMetrics(ctx.Context)
		if serviceId != "" {
			req = req.ServiceId(serviceId)
		}
		if instanceId != "" {
			req = req.InstanceId(instanceId)
		}
		if start != nil {
			req = req.Start(*start)
		}
		if end != nil {
			req = req.End(*end)
		}
		res, resp, err := req.Name(metricName).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				fmt.Sprintf("Error while retrieving the metrics %s for %s/%s", metricHumanName, serviceId, instanceId),
				err,
				resp,
			)
		}

		getMetricsReply := NewGetMetricsReply(ctx.Mapper, metricName, metricHumanName, res, full)
		renderer = renderer.Render(getMetricsReply)
	}
	return nil
}

type GetMetricsReply struct {
	mapper                  *idmapper.Mapper
	metricName              string
	metricHumanReadableName string
	res                     *koyeb.GetMetricsReply
	full                    bool
}

func NewGetMetricsReply(mapper *idmapper.Mapper, metricName string, metricHumanName string, res *koyeb.GetMetricsReply, full bool) *GetMetricsReply {
	return &GetMetricsReply{
		mapper,
		metricName,
		metricHumanName,
		res,
		full,
	}
}

func (r *GetMetricsReply) Title() string {
	return r.metricHumanReadableName
}

// MetricsJSON is the JSON representation of the metrics, used to display when -o json is used.
type MetricsJSON struct {
	Name string            `json:"metric_name"`
	Data []MetricsJSONData `json:"data"`
}

type MetricsJSONData struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

func (r *GetMetricsReply) MarshalBinary() ([]byte, error) {
	v := MetricsJSON{
		Name: r.metricName,
	}

	for _, metric := range r.res.GetMetrics() {
		for _, sample := range metric.GetSamples() {
			v.Data = append(v.Data, MetricsJSONData{
				Timestamp: sample.GetTimestamp(),
				Value:     sample.GetValue(),
			})
		}
	}
	return json.Marshal(v)
}

func (r *GetMetricsReply) Headers() []string {
	return []string{"timestamp", "value"}
}

func (r *GetMetricsReply) Fields() []map[string]string {
	resp := []map[string]string{}

	for _, metric := range r.res.GetMetrics() {
		for _, sample := range metric.GetSamples() {
			resp = append(resp, map[string]string{
				"timestamp": sample.GetTimestamp(),
				"value":     fmt.Sprintf("%f", sample.GetValue()),
			})
		}
	}
	return resp
}
