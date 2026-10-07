package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

// List lists the service pools of the current scope, optionally filtered
// by --name. Unset --limit/--offset keep the fetch-all walk bounded by the
// reported total; set flags bound the listing window.
func (h *PoolHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")

	offset, offsetSet := int64(0), cmd.Flags().Changed("offset")
	if offsetSet {
		offset, _ = cmd.Flags().GetInt64("offset")
	}
	limit, limitSet := int64(0), cmd.Flags().Changed("limit")
	if limitSet {
		limit, _ = cmd.Flags().GetInt64("limit")
	}

	list, err := idmapper.FetchAllPages(func(pageOffset, pageLimit int64) ([]koyeb.ServicePool, int64, error) {
		// The walk pages in window coordinates: pageOffset counts the items
		// already served since the window start. Translate to the listing
		// coordinates the API expects, and cap the reported total so the walk
		// still bounds itself on it when flags are set. Unset flags keep the
		// fetch-all behavior untouched.
		reqLimit := pageLimit
		if limitSet {
			reqLimit = min(pageLimit, limit-pageOffset)
		}
		res, resp, err := ctx.API.ListServicePools(
			ctx.Context,
			name,
			strconv.FormatInt(offset+pageOffset, 10),
			strconv.FormatInt(reqLimit, 10),
		)
		if err != nil {
			return nil, 0, errors.NewCLIErrorFromAPIError("Error while listing pools", err, resp)
		}
		count := res.GetCount() - offset
		if limitSet {
			count = min(count, limit)
		}
		if count < 0 {
			count = 0
		}
		return res.GetServicePools(), count, nil
	}, 100)
	if err != nil {
		return err
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
