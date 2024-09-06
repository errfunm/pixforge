package domain

import (
	"context"
)

type BuildImageOpts struct {
	Width       *int
	Height      *int
	AspectRatio *AR
	ImageType   *ImageType
}

type ImageRepoInterface interface {
	BuildImageOf(ctx context.Context, image []byte, opts BuildImageOpts) ([]byte, error)
}

func (bo BuildImageOpts) SetWidth(width int) BuildImageOpts {
	bo.Width = &width
	return bo
}

func (bo BuildImageOpts) SetHeight(height int) BuildImageOpts {
	bo.Height = &height
	return bo
}

func (bo BuildImageOpts) SetAr(ar AR) BuildImageOpts {
	bo.AspectRatio = &ar
	return bo
}

func (bo BuildImageOpts) SetFormat(format ImageType) BuildImageOpts {
	bo.ImageType = &format
	return bo
}

func NewBuildImageOpts() BuildImageOpts {
	return BuildImageOpts{}
}
