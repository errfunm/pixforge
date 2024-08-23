package domain

import (
	"context"
	"example.com/imageProc/domain"
)

type BuildImageOpts struct {
	TenantOpts  domain.TenantOpts
	Width       int
	AspectRatio domain.AR
	ImageType   domain.ImageType
}

type GetImageOpts struct {
	TenantOpts  domain.TenantOpts
	Name        string
	Width       int
	AspectRatio domain.AR
	Type        domain.ImageType
}

type ImageRepoInterface interface {
	GetImage(ctx context.Context, opts GetImageOpts) ([]byte, error)
	BuildImageOf(ctx context.Context, image []byte, opts BuildImageOpts) ([]byte, error)
	CreateImage(ctx context.Context, image []byte, opts domain.TenantOpts) (string, error)
}
