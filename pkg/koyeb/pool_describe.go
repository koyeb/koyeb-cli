package koyeb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

// Describe retrieves a service pool with its definition details.
func (h *PoolHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	poolID, err := ResolvePoolArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.ServicePoolsApi.GetServicePool(ctx.Context, poolID).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the pool `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	describePoolReply := NewDescribePoolReply(ctx.Mapper, res.GetServicePool(), full)
	ctx.Renderer.Render(describePoolReply)
	return nil
}

type DescribePoolReply struct {
	mapper *idmapper.Mapper
	value  koyeb.ServicePool
	full   bool
}

func NewDescribePoolReply(mapper *idmapper.Mapper, value koyeb.ServicePool, full bool) *DescribePoolReply {
	return &DescribePoolReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (DescribePoolReply) Title() string {
	return "Pool"
}

func (r *DescribePoolReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *DescribePoolReply) Headers() []string {
	return []string{
		"id", "name", "status", "size", "ready_count", "generation",
		"image", "instance_types", "regions", "created_at", "updated_at",
	}
}

func (r *DescribePoolReply) Fields() []map[string]string {
	item := r.value
	definition := item.GetDefinition()
	docker := definition.GetDocker()

	fields := map[string]string{
		"id":             renderer.FormatID(item.GetId(), r.full),
		"name":           item.GetName(),
		"status":         string(item.GetStatus()),
		"size":           strconv.FormatInt(item.GetSize(), 10),
		"ready_count":    strconv.FormatInt(item.GetReadyCount(), 10),
		"generation":     item.GetGeneration(),
		"image":          docker.GetImage(),
		"instance_types": formatPoolInstanceTypes(definition.GetInstanceTypes()),
		"regions":        strings.Join(definition.GetRegions(), ","),
		"created_at":     renderer.FormatTime(item.GetCreatedAt()),
		"updated_at":     renderer.FormatTime(item.GetUpdatedAt()),
	}

	return []map[string]string{fields}
}

func formatPoolInstanceTypes(instanceTypes []koyeb.DeploymentInstanceType) string {
	names := make([]string, 0, len(instanceTypes))
	for i := range instanceTypes {
		if name := instanceTypes[i].GetType(); name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, ",")
}
