package domainsvc

import (
	"context"
	"errors"
	"example.com/imageProc/domain"
)

type GetImageOpts struct {
	tenantOpts domain.TenantOpts
	Name       string
	Width      int
	Ar         domain.AR
	Type       domain.ImageType
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
	imgId, err := i.repo.CreateImage(ctx, imageByte, tenantOpts)
	if err != nil {
		return "", err
	}
	return imgId, nil
}

func (i ImageService) GetImage(ctx context.Context, opts GetImageOpts) ([]byte, error) {
	img, err := i.repo.GetImage(ctx, domain.GetImageOpts{
		TenantOpts:  opts.tenantOpts,
		Name:        opts.Name,
		Width:       opts.Width,
		AspectRatio: opts.Ar,
		Type:        opts.Type,
	})
	if err == nil {
		return img, nil
	}
	if !errors.Is(err, domain.ErrImageNotFound) {
		return nil, err
	}

	primaryImg, err := i.repo.GetImage(ctx, domain.GetImageOpts{
		TenantOpts:  opts.tenantOpts,
		IsPrimary:   true,
		Name:        opts.Name,
		Width:       opts.Width,
		AspectRatio: opts.Ar,
		Type:        opts.Type,
	})
	if err != nil {
		if errors.Is(err, domain.ErrImageNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	builtImage, err := i.repo.BuildImageOf(ctx, primaryImg, domain.BuildImageOpts{
		Width:       opts.Width,
		AspectRatio: opts.Ar,
		ImageType:   opts.Type,
	})
	if err != nil {
		return nil, err
	}

	// cache image before return
	_, err = i.repo.CreateImage(ctx, builtImage, opts.tenantOpts)
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
