package idmapper

import (
	"context"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type RegionalDeploymentMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
}

func NewRegionalDeploymentMapper(ctx context.Context, client *koyeb.APIClient) *RegionalDeploymentMapper {
	return &RegionalDeploymentMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
	}
}

func (mapper *RegionalDeploymentMapper) ResolveID(val string) (string, error) {
	if IsUUIDv4(val) {
		return val, nil
	}

	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	id, ok := mapper.sidMap.GetID(val)
	if ok {
		return id, nil
	}
	return "", errors.NewCLIErrorForMapperResolve(
		"secret",
		val,
		[]string{"regional deployment full UUID", "regional deployment short ID (8 characters)"},
	)
}

func (mapper *RegionalDeploymentMapper) GetShortID(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	sid, ok := mapper.sidMap.GetValue(id)
	if !ok {
		return "", errors.NewCLIErrorForMapperResolve(
			"secret",
			id,
			[]string{"regional deployment full UUID", "regional deployment short ID (8 characters)"},
		)
	}
	return sid, nil
}

func (mapper *RegionalDeploymentMapper) fetch() error {
	radix := NewRadixTree()

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {

		res, resp, err := mapper.client.RegionalDeploymentsApi.ListRegionalDeployments(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error listing the regional deployments to resolve the provided identifier to an object ID",
				err,
				resp,
			)
		}

		deployments := res.GetRegionalDeployments()
		for i := range deployments {
			deployment := &deployments[i]
			radix.Insert(getKey(deployment.GetId()), deployment)
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	minLength := radix.MinimalLength(8)
	err := radix.ForEach(func(key Key, value Value) error {
		deployment := value.(*koyeb.RegionalDeploymentListItem)
		id := deployment.GetId()
		sid := getShortID(id, minLength)

		mapper.sidMap.Set(id, sid)

		return nil
	})
	if err != nil {
		return err
	}

	mapper.fetched = true

	return nil
}
