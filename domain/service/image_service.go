package domainsvc

import (
	"context"
	"example.com/imageProc/domain"
)

type GetImageOpts struct {
	tenantOpts domain.TenantOpts
	Width      int
	Ar         domain.AR
	Type       domain.ImageType
}

type ImageServiceInterface interface {
	Upload(ctx context.Context, imageByte []byte, tenantOpts domain.TenantOpts) (string, error)
	GetImage(ctx context.Context, imageId string, opts GetImageOpts) ([]byte, error)
}
