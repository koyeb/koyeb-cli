package idmapper

import (
	"context"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type SnapshotMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
	nameMap *IDMap
}

func NewSnapshotMapper(ctx context.Context, client *koyeb.APIClient) *SnapshotMapper {
	return &SnapshotMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
		nameMap: NewIDMap(),
	}
}

func (mapper *SnapshotMapper) ResolveID(val string) (string, error) {
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
		"snapshot",
		val,
		[]string{"snapshot full UUID", "snapshot short ID (8 characters)", "snapshot name"},
	)
}

func (mapper *SnapshotMapper) fetch() error {
	radix := NewRadixTree()

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := mapper.client.SnapshotsApi.ListSnapshots(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error listing snapshots to resolve the provided identifier to an object ID",
				err,
				resp,
			)
		}

		snapshots := res.GetSnapshots()

		if len(snapshots) == 0 {
			break
		}

		for i := range snapshots {
			snapshot := &snapshots[i]
			radix.Insert(getKey(snapshot.GetId()), snapshot)
		}

		page++
		offset = page * limit
	}

	minLength := radix.MinimalLength(8)
	err := radix.ForEach(func(key Key, value Value) error {
		snapshot := value.(*koyeb.Snapshot)
		id := snapshot.GetId()
		name := snapshot.GetName()
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
