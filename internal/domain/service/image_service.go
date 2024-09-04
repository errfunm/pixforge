package domainsvc

import (
	"context"
	"errors"
	"example.com/imageProc/internal/domain"
	"example.com/imageProc/internal/infra/persist"
)

type GetImageOpts struct {
	TenantOpts domain.TenantOpts
	Name       string
	Width      *int
	Height     *int
	Ar         *domain.AR
	Type       *domain.ImageType
}

type ImageServiceInterface interface {
	Upload(ctx context.Context, imageByte []byte, tenantOpts domain.TenantOpts) (string, error)
	GetImage(ctx context.Context, opts GetImageOpts) ([]byte, error)
}

type ImageService struct {
	repo domain.ImageRepoInterface
}

var (
	ErrNotFound = errors.New("no primary image found")
)

func (i ImageService) Upload(ctx context.Context, imageByte []byte, tenantOpts domain.TenantOpts) (string, error) {
	imgId, err := i.repo.CreateImage(ctx, imageByte, true, "", tenantOpts)
	if err != nil {
		return "", err
	}
	return imgId, nil
}

func (i ImageService) GetImage(ctx context.Context, opts GetImageOpts) ([]byte, error) {
	img, err := i.repo.GetImage(ctx, domain.RepoGetImageOpts{
		TenantOpts:  opts.TenantOpts,
		Name:        opts.Name,
		Width:       opts.Width,
		Height:      opts.Height,
		AspectRatio: opts.Ar,
		Type:        opts.Type,
	})
	if err == nil {
		return img, nil
	}
	if !errors.Is(err, persist.ErrImageNotFound) {
		return nil, err
	}

	parentImg, err := i.repo.GetImage(ctx, domain.RepoGetImageOpts{
		TenantOpts: opts.TenantOpts,
		IsParent:   true,
		Name:       opts.Name,
	})
	if err != nil {
		if errors.Is(err, persist.ErrImageNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	builtImage, err := i.repo.BuildImageOf(ctx, parentImg, domain.BuildImageOpts{
		Width:       opts.Width,
		Height:      opts.Height,
		AspectRatio: opts.Ar,
		ImageType:   opts.Type,
	})
	if err != nil {
		if errors.Is(err, persist.ErrUnSupportedImageFormat) {
			// TODO: internal error should be returned
		}
		return nil, err
	}
	// cache image before return
	_, err = i.repo.CreateImage(ctx, builtImage, false, opts.Name, opts.TenantOpts)
	if err != nil {
		return nil, err
	}

	return builtImage, nil
}

func NewImageService(repo domain.ImageRepoInterface) ImageServiceInterface {
	return ImageService{
		repo: repo,
	}
}

func (gio GetImageOpts) SetTenantOpts(tenantOpts domain.TenantOpts) GetImageOpts {
	gio.TenantOpts = tenantOpts
	return gio
}

func (gio GetImageOpts) SetName(name string) GetImageOpts {
	gio.Name = name
	return gio
}

func (gio GetImageOpts) SetWidth(width int) GetImageOpts {
	gio.Width = &width
	return gio
}

func (gio GetImageOpts) SetHeight(height int) GetImageOpts {
	gio.Height = &height
	return gio
}

func (gio GetImageOpts) SetAr(ar domain.AR) GetImageOpts {
	gio.Ar = &ar
	return gio
}

func (gio GetImageOpts) SetFormat(format domain.ImageType) GetImageOpts {
	gio.Type = &format
	return gio
}

func NewServiceGetImageOpts() GetImageOpts {
	return GetImageOpts{}
}
