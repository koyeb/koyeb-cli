package idmapper

import (
	"context"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type ProjectMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
	nameMap *IDMap
}

func NewProjectMapper(ctx context.Context, client *koyeb.APIClient) *ProjectMapper {
	return &ProjectMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
		nameMap: NewIDMap(),
	}
}

func (mapper *ProjectMapper) ResolveID(val string) (string, error) {
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

	id, ok = mapper.nameMap.GetID(val)
	if ok {
		return id, nil
	}
	return "", errors.NewCLIErrorForMapperResolve(
		"project",
		val,
		[]string{"project full UUID", "project short ID (8 characters)", "project name"},
	)
}

func (mapper *ProjectMapper) fetch() error {
	radix := NewRadixTree()

	page := int64(0)
	limit := int64(100)
	for {
		offset := page * limit
		res, resp, err := mapper.client.ProjectsApi.ListProjects(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error listing projects to resolve the provided identifier to an object ID",
				err,
				resp,
			)
		}

		projects := res.GetProjects()
		for i := range projects {
			project := &projects[i]
			radix.Insert(getKey(project.GetId()), project)
		}

		if int64(len(projects)) < limit {
			break
		}

		page++
	}

	minLength := radix.MinimalLength(8)
	err := radix.ForEach(func(key Key, value Value) error {
		project := value.(*koyeb.Project)
		id := project.GetId()
		name := project.GetName()
		sid := getShortID(id, minLength)

		mapper.sidMap.Set(id, sid)
		mapper.nameMap.Set(id, name)

		return nil
	})
	if err != nil {
		return err
	}

	mapper.fetched = true
	return nil
}
