package koyeb

import (
	"github.com/go-openapi/strfmt"
	"github.com/koyeb/koyeb-cli/pkg/gen/kclient/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetTable(t *testing.T) {
	baseItem := models.StorageConnectorListItem{
		ID:        "d7a63c2b-2289-40cb-9e02-b02d0d7ecdc0",
		Name:      "foo",
		Type:      models.StorageConnectorTypeCloudevent,
		URL:       "https://connectors.prod.koyeb.com/cloudevent/foo/bar",
		CreatedAt: strfmt.DateTime(time.Date(2020, 10, 1, 0, 12, 0, 0, time.UTC)),
		UpdatedAt: strfmt.DateTime(time.Date(2020, 11, 1, 0, 12, 0, 0, time.UTC)),
	}

	testCases := map[string]struct {
		in     []models.StorageConnectorListItem
		fields [][]string
	}{
		"empty": {
			in: []models.StorageConnectorListItem{},
		},
		"simple": {
			in: []models.StorageConnectorListItem{
				baseItem,
			},
			fields: [][]string{
				{"d7a63c2b-2289-40cb-9e02-b02d0d7ecdc0", "foo", "cloudevent", "https://connectors.prod.koyeb.com/cloudevent/foo/bar", "2020-10-01T00:12:00.000Z", "2020-11-01T00:12:00.000Z"},
			},
		},
	}

	for n, tt := range testCases {
		t.Run(n, func(t *testing.T) {
			cl := ConnectorList{Elts: tt.in}
			res := cl.GetTable()
			assert.Equal(t, tt.fields, res.fields)
		})
	}
}
