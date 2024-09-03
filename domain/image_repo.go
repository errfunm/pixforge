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

type RepoGetImageOpts struct {
	TenantOpts  TenantOpts
	IsParent    bool
	Name        string
	Width       *int
	Height      *int
	AspectRatio *AR
	Type        *ImageType
}

type ImageRepoInterface interface {
	GetImage(ctx context.Context, opts RepoGetImageOpts) ([]byte, error)
	BuildImageOf(ctx context.Context, image []byte, opts BuildImageOpts) ([]byte, error)
	CreateImage(ctx context.Context, image []byte, isParent bool, name string, opts TenantOpts) (string, error)
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

func (gio RepoGetImageOpts) SetTenantOpts(tenantOpts TenantOpts) RepoGetImageOpts {
	gio.TenantOpts = tenantOpts
	return gio
}

func (gio RepoGetImageOpts) SetIsParent(isParent bool) RepoGetImageOpts {
	gio.IsParent = isParent
	return gio
}

func (gio RepoGetImageOpts) SetName(name string) RepoGetImageOpts {
	gio.Name = name
	return gio
}

func (gio RepoGetImageOpts) SetWidth(width int) RepoGetImageOpts {
	gio.Width = &width
	return gio
}

func (gio RepoGetImageOpts) SetHeight(height int) RepoGetImageOpts {
	gio.Height = &height
	return gio
}

func (gio RepoGetImageOpts) SetAr(ar AR) RepoGetImageOpts {
	gio.AspectRatio = &ar
	return gio
}

func (gio RepoGetImageOpts) SetFormat(format ImageType) RepoGetImageOpts {
	gio.Type = &format
	return gio
}

func NewRepoGetImageOpts() RepoGetImageOpts {
	return RepoGetImageOpts{}
}
