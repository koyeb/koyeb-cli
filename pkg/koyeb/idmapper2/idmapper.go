package idmapper2

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

type Mapper struct {
	app *AppMapper
}

func NewMapper(ctx context.Context, client *koyeb.APIClient) *Mapper {
	return &Mapper{
		app: NewAppMapper(ctx, client),
	}
}

func (mapper *Mapper) App() *AppMapper {
	return mapper.app
}
