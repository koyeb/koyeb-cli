package idmapper2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
)

type SecretMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
	nameMap *IDMap
}

func NewSecretMapper(ctx context.Context, client *koyeb.APIClient) *SecretMapper {
	return &SecretMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
		nameMap: NewIDMap(),
	}
}

func (mapper *SecretMapper) ResolveID(val string) (string, error) {
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

	return "", fmt.Errorf("id not found %q", val)
}

func (mapper *SecretMapper) GetShortID(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	sid, ok := mapper.sidMap.GetValue(id)
	if !ok {
		return "", fmt.Errorf("secret short id not found for %q", id)
	}

	return sid, nil
}

func (mapper *SecretMapper) fetch() error {
	radix := NewRadixTree()

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {

		resp, _, err := mapper.client.SecretsApi.ListSecrets(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.Wrap(err, "cannot list apps from API")
		}

		secrets := resp.GetSecrets()
		for i := range secrets {
			secret := &secrets[i]
			id := secret.GetId()
			radix.Insert(Key(strings.ReplaceAll(id, "-", "")), secret)
		}

		page++
		offset = page * limit
		if offset >= resp.GetCount() {
			break
		}
	}

	minLength := radix.MinimalLength(8)
	radix.ForEach(func(key Key, value Value) {
		secret := value.(*koyeb.Secret)
		id := secret.GetId()
		name := secret.GetName()
		sid := strings.ReplaceAll(id, "-", "")[:minLength]

		mapper.sidMap.Set(id, sid)
		mapper.nameMap.Set(id, name)
	})

	mapper.fetched = true

	return nil
}
