package idmapper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type AppMapper struct {
	ctx           context.Context
	client        *koyeb.APIClient
	project       string
	fetched       bool
	sidMap        *IDMap
	nameMap       *IDMap
	autoDomainMap *IDMap
}

func NewAppMapper(ctx context.Context, client *koyeb.APIClient, project string) *AppMapper {
	return &AppMapper{
		ctx:           ctx,
		client:        client,
		project:       project,
		fetched:       false,
		sidMap:        NewIDMap(),
		nameMap:       NewIDMap(),
		autoDomainMap: NewIDMap(),
	}
}

func (mapper *AppMapper) ResolveID(val string) (string, error) {
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
		"application",
		val,
		[]string{"application full UUID", "application short ID (8 characters)", "application name"},
	)
}

func (mapper *AppMapper) GetName(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	name, ok := mapper.nameMap.GetValue(id)
	if !ok {
		res, resp, err := mapper.client.AppsApi.GetApp(mapper.ctx, id).Execute()
		if err != nil {
			return "", errors.NewCLIErrorFromAPIError(
				fmt.Sprintf("Error retrieving the application %q", id),
				err,
				resp,
			)
		}
		app := res.GetApp()
		return app.GetName(), nil
	}

	return name, nil
}

func (mapper *AppMapper) GetAutoDomain(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	name, ok := mapper.autoDomainMap.GetValue(id)
	if !ok {
		res, resp, err := mapper.client.AppsApi.GetApp(mapper.ctx, id).Execute()
		if err != nil {
			return "", errors.NewCLIErrorFromAPIError(
				fmt.Sprintf("Error retrieving the application %q", id),
				err,
				resp,
			)
		}
		app := res.GetApp()
		for _, domain := range app.GetDomains() {
			if domain.GetType() != koyeb.DOMAINTYPE_AUTOASSIGNED {
				continue
			}

			if !domain.HasCloudflare() {
				continue
			}

			return domain.GetId(), nil
		}

		return "", fmt.Errorf("app automatic domain not found for %q", id)
	}

	return name, nil
}

func (mapper *AppMapper) fetch() error {
	radix := NewRadixTree()

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {

		req := mapper.client.AppsApi.ListApps(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10))
		if mapper.project != "" {
			req = req.ProjectId(mapper.project)
		}
		res, resp, err := req.Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error listing applications to resolve the provided identifier to an object ID",
				err,
				resp,
			)
		}

		apps := res.GetApps()
		for i := range apps {
			app := &apps[i]
			radix.Insert(getKey(app.GetId()), app)
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	minLength := radix.MinimalLength(8)
	err := radix.ForEach(func(key Key, value Value) error {
		app := value.(*koyeb.AppListItem)
		id := app.GetId()
		name := app.GetName()
		sid := getShortID(id, minLength)
		domains := app.GetDomains()

		mapper.sidMap.Set(id, sid)
		mapper.nameMap.Set(id, name)
		for _, domain := range domains {
			if domain.GetType() != koyeb.DOMAINTYPE_AUTOASSIGNED {
				continue
			}

			// We want the original autoassigned domain for the app, not other ones that
			// could have been provisioned with Koyeb Load Balancer, for example
			if !domain.HasCloudflare() {
				continue
			}

			mapper.autoDomainMap.Set(id, domain.GetId())
			break
		}

		return nil
	})
	if err != nil {
		return err
	}

	mapper.fetched = true

	return nil
}
