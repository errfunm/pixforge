package domain

import (
	"context"
	"errors"
)

type BuildImageOpts struct {
	Width       int
	AspectRatio AR
	ImageType   ImageType
}

type GetImageOpts struct {
	TenantOpts  TenantOpts
	IsPrimary   bool
	Name        string
	Width       int
	AspectRatio AR
	Type        ImageType
}

var (
	ErrImageNotFound = errors.New("image not found")
)

type ImageRepoInterface interface {
	GetImage(ctx context.Context, opts GetImageOpts) ([]byte, error)
	BuildImageOf(ctx context.Context, image []byte, opts BuildImageOpts) ([]byte, error)
	CreateImage(ctx context.Context, image []byte, opts TenantOpts) (string, error)
}
