package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

// List lists the service pools of the current scope.
func (h *PoolHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.ServicePool{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := ctx.Client.ServicePoolsApi.ListServicePools(ctx.Context).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError("Error while listing pools", err, resp)
		}
		pools := res.GetServicePools()
		if len(pools) == 0 {
			break
		}
		list = append(list, pools...)

		page++
		offset = page * limit
	}

	full := GetBoolFlags(cmd, "full")
	listPoolsReply := NewListPoolsReply(ctx.Mapper, &koyeb.ListServicePoolsReply{ServicePools: list}, full)
	ctx.Renderer.Render(listPoolsReply)
	return nil
}

type ListPoolsReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListServicePoolsReply
	full   bool
}

func NewListPoolsReply(mapper *idmapper.Mapper, value *koyeb.ListServicePoolsReply, full bool) *ListPoolsReply {
	return &ListPoolsReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListPoolsReply) Title() string {
	return "Pools"
}

func (r *ListPoolsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListPoolsReply) Headers() []string {
	return []string{"id", "name", "status", "size", "ready_count", "generation", "created_at"}
}

func (r *ListPoolsReply) Fields() []map[string]string {
	pools := r.value.GetServicePools()
	resp := make([]map[string]string, 0, len(pools))

	for i := range pools {
		pool := &pools[i]
		fields := map[string]string{
			"id":          renderer.FormatID(pool.GetId(), r.full),
			"name":        pool.GetName(),
			"status":      string(pool.GetStatus()),
			"size":        strconv.FormatInt(pool.GetSize(), 10),
			"ready_count": strconv.FormatInt(pool.GetReadyCount(), 10),
			"generation":  pool.GetGeneration(),
			"created_at":  renderer.FormatTime(pool.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}
