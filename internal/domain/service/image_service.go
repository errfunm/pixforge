package domainsvc

import (
	"context"
	"errors"
	appsvc "example.com/imageProc/internal/app/service"
	"example.com/imageProc/internal/domain"
	"example.com/imageProc/internal/infra/persist"
	"github.com/davidbyttow/govips/v2/vips"
	"math"
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
	repo           domain.ImageRepoInterface
	storageService appsvc.ImageStorageServiceInterface
}

var (
	ErrNotFound = errors.New("no primary image found")
)

func (i ImageService) Upload(ctx context.Context, imageByte []byte, tenantOpts domain.TenantOpts) (string, error) {
	imgId, err := i.storageService.StoreParentImage(imageByte, tenantOpts)
	if err != nil {
		return "", err
	}
	return imgId, nil
}

func (i ImageService) GetImage(ctx context.Context, opts GetImageOpts) ([]byte, error) {
	var parentImage []byte
	var parentImageWidth, parentImageHeight, newWidth, newHeight int
	var parentImageFormat, newImageFormat domain.ImageType
	// check whether parentImage needs to be fetched at first or not
	parentHasToBeFetched := false
	if opts.Type == nil {
		parentHasToBeFetched = true
	} else {
		if opts.Width == nil {
			if opts.Height == nil || opts.Ar == nil {
				parentHasToBeFetched = true
			}
		}
		if opts.Height == nil {
			if opts.Width == nil || opts.Ar == nil {
				parentHasToBeFetched = true
			}
		}
	}
	// if true then fetch parentImage
	if parentHasToBeFetched {
		parentImage, err := i.storageService.GetParentImage(opts.Name, opts.TenantOpts)
		if err != nil {
			if errors.Is(err, appsvc.ErrNoMatchingFile) {
				return nil, ErrNotFound
			}
			return nil, errors.New("internal error")
		}
		imgRef, err := vips.NewImageFromBuffer(parentImage)
		if err != nil {
			return nil, err
		}
		parentImageWidth = imgRef.Width()
		parentImageHeight = imgRef.Height()
		parentImageFormat, err = domain.ImageTypeFromString(imgRef.Format().FileExt())
		if err != nil {
			return nil, errors.New("internal error")
		}
	}
	// determineDimensions
	newWidth, newHeight = determineDimensions(opts.Width, opts.Height, opts.Ar, parentImageWidth, parentImageHeight)
	// determineImageFormat
	if opts.Type == nil {
		newImageFormat = parentImageFormat
	}
	newImageFormat = *opts.Type
	// fetch childImage
	childImage, err := i.storageService.GetChildImage(opts.Name, newImageFormat, newWidth, newHeight, opts.TenantOpts)
	if err == nil {
		return childImage, nil
	}
	if !errors.Is(err, appsvc.ErrNoMatchingFile) {
		return nil, errors.New("internal error")
	}
	// fetch parentImage to buildImageFrom
	if parentImage == nil {
		parentImage, err = i.storageService.GetParentImage(opts.Name, opts.TenantOpts)
		if err != nil {
			if errors.Is(err, appsvc.ErrNoMatchingFile) {
				return nil, ErrNotFound
			}
			return nil, errors.New("internal error")
		}
	}
	// buildImage then return
	builtImage, err := i.repo.BuildImageOf(ctx, parentImage, domain.BuildImageOpts{
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
	err = i.storageService.StoreChildImage(builtImage, opts.Name, opts.TenantOpts)
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

func determineDimensions(targetWidth, targetHeight *int, targetAr *domain.AR, originalDimensions ...int) (int, int) {
	originalWidth := originalDimensions[0]
	originalHeight := originalDimensions[1]

	originalAr := domain.NewAspectRatioFrom(originalWidth, originalHeight)
	var newWidth, newHeight int
	switch {
	case targetWidth != nil && targetHeight != nil && targetAr != nil:
		if math.Abs(float64(*targetWidth)/float64(*targetHeight)-targetAr.Float64()) < 1e-6 {
			newWidth = *targetWidth
			newHeight = *targetHeight
		} else {
			newWidth = *targetWidth
			newHeight = int(math.Round(float64(newWidth) / targetAr.Float64()))
		}
	case targetWidth != nil && targetHeight == nil && targetAr == nil:
		newWidth = *targetWidth
		newHeight = int(math.Round(float64(newWidth) / originalAr.Float64()))
	case targetWidth == nil && targetHeight != nil && targetAr == nil:
		newHeight = *targetHeight
		newWidth = int(math.Round(float64(newHeight) * originalAr.Float64()))
	case targetWidth == nil && targetHeight == nil && targetAr != nil:
		if originalAr.Float64() > targetAr.Float64() {
			newHeight = originalHeight
			newWidth = int(math.Round(float64(newHeight) * targetAr.Float64()))
		} else {
			newWidth = originalWidth
			newHeight = int(math.Round(float64(newWidth) / targetAr.Float64()))
		}
	case targetWidth != nil && targetHeight == nil && targetAr != nil:
		newWidth = *targetWidth
		newHeight = int(math.Round(float64(newWidth) / targetAr.Float64()))
	case targetWidth == nil && targetHeight != nil && targetAr != nil:
		newHeight = *targetHeight
		newWidth = int(math.Round(float64(newHeight) * targetAr.Float64()))
	case targetWidth != nil && targetHeight != nil && targetAr == nil:
		newWidth = *targetWidth
		newHeight = *targetHeight
	case targetWidth == nil && targetHeight == nil && targetAr == nil:
		newWidth = originalWidth
		newHeight = originalHeight
	}
	return newWidth, newHeight
}
