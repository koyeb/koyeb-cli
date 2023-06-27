package idmapper

import (
	"context"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type DeploymentMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
}

func NewDeploymentMapper(ctx context.Context, client *koyeb.APIClient) *DeploymentMapper {
	return &DeploymentMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
	}
}

func (mapper *DeploymentMapper) ResolveID(val string) (string, error) {
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
		"deployments",
		val,
		[]string{"deployment full UUID", "deployment short ID (8 characters)"},
	)
}

func (mapper *DeploymentMapper) GetShortID(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	sid, ok := mapper.sidMap.GetValue(id)
	if !ok {
		return "", errors.NewCLIErrorForMapperResolve(
			"deployments",
			id,
			[]string{"deployment full UUID", "deployment short ID (8 characters)"},
		)
	}

	return sid, nil
}

func (mapper *DeploymentMapper) fetch() error {
	radix := NewRadixTree()

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {

		res, resp, err := mapper.client.DeploymentsApi.ListDeployments(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error listing deployments to resolve the provided identifier to an object ID",
				err,
				resp,
			)
		}

		deployments := res.GetDeployments()
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
		deployment := value.(*koyeb.DeploymentListItem)
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
