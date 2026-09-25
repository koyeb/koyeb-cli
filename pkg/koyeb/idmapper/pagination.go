package idmapper

// FetchAllPages pages through a list endpoint until an empty page arrives
// or the reported total is covered. The total bound keeps a server that
// ignores the offset from looping the CLI forever.
func FetchAllPages[T any](fetch func(offset, limit int64) ([]T, int64, error), limit int64) ([]T, error) {
	var all []T
	page := int64(0)
	offset := int64(0)
	for {
		items, count, err := fetch(offset, limit)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return all, nil
		}
		all = append(all, items...)

		page++
		offset = page * limit
		if offset >= count {
			return all, nil
		}
	}
}
