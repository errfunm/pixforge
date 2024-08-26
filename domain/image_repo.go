package domain

import (
	"context"
)

type BuildImageOpts struct {
	Width       int
	AspectRatio AR
	ImageType   ImageType
}

type GetImageOpts struct {
	TenantOpts  TenantOpts
	IsParent    bool
	Name        string
	Width       int
	AspectRatio AR
	Type        ImageType
}

type ImageRepoInterface interface {
	GetImage(ctx context.Context, opts GetImageOpts) ([]byte, error)
	BuildImageOf(ctx context.Context, image []byte, opts BuildImageOpts) ([]byte, error)
	CreateImage(ctx context.Context, image []byte, isParent bool, name string, opts TenantOpts) (string, error)
}
