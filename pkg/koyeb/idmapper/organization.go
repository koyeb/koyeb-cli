package idmapper

import (
	"context"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type OrganizationMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
	nameMap *IDMap
}

func NewOrganizationMapper(ctx context.Context, client *koyeb.APIClient) *OrganizationMapper {
	return &OrganizationMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
		nameMap: NewIDMap(),
	}
}

func (mapper *OrganizationMapper) ResolveID(val string) (string, error) {
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
		"organization",
		val,
		[]string{"organization full UUID", "organization short ID (8 characters)", "organization name"},
	)
}

func (mapper *OrganizationMapper) getCurrentUserId() (string, error) {
	res, resp, err := mapper.client.ProfileApi.GetCurrentUser(mapper.ctx).Execute()
	if err != nil {
		return "", errors.NewCLIErrorFromAPIError("Your authentication token is not linked to a user", err, resp)
	}
	return *res.GetUser().Id, nil
}

func (mapper *OrganizationMapper) fetch() error {
	radix := NewRadixTree()

	userId, err := mapper.getCurrentUserId()
	if err != nil {
		return err
	}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := mapper.client.OrganizationMembersApi.
			ListOrganizationMembers(mapper.ctx).
			UserId(userId).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error listing organizations to resolve the provided identifier to an object ID",
				err,
				resp,
			)
		}

		members := res.GetMembers()
		for i := range members {
			member := &members[i]
			radix.Insert(getKey(member.Organization.GetId()), member)
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	minLength := radix.MinimalLength(8)
	err = radix.ForEach(func(key Key, value Value) error {
		member := value.(*koyeb.OrganizationMember)
		id := member.Organization.GetId()
		name := member.Organization.GetName()
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
