package domainsvc

import (
	"context"
	"errors"
	appsvc "example.com/imageProc/internal/app/service"
	"example.com/imageProc/internal/domain"
	"fmt"
	"math"
)

type GetImageOpts struct {
	TenantOpts domain.TenantOpts
	Name       string
	Width      *int
	Height     *int
	Ar         *domain.AR
	Type       domain.ImageType
}

type ImageServiceInterface interface {
	Upload(ctx context.Context, imageByte []byte, tenantOpts domain.TenantOpts) (string, error)
	GetImage(ctx context.Context, opts GetImageOpts) ([]byte, error)
}

type ImageService struct {
	processorService appsvc.ImageProcessingServiceInterface
	storageService   appsvc.ImageStorageServiceInterface
}

var (
	ErrNotFound               = errors.New("no primary image found")
	ErrUnsupportedImageFormat = errors.New("unsupported image format")
)

func (i ImageService) Upload(ctx context.Context, imageByte []byte, tenantOpts domain.TenantOpts) (string, error) {
	format, err := i.processorService.GetFormat(imageByte)
	if err != nil {
		if errors.Is(err, appsvc.ErrUnsupportedImageFormat) {
			return "", ErrUnsupportedImageFormat
		}
		return "", fmt.Errorf("internal error: %v", err)
	}

	imgId, err := i.storageService.StoreParentImage(imageByte, format, tenantOpts)
	if err != nil {
		return "", err
	}
	return imgId, nil
}

func (i ImageService) GetImage(ctx context.Context, opts GetImageOpts) ([]byte, error) {
	var parentImage []byte
	var parentImageSpec domain.ImageSpec
	var targetWidth, targetHeight int
	var targetImageFormat domain.ImageType
	// check whether parentImage needs to be fetched at first or not
	if parentImageNeedsToBeFetched(opts) {
		var err error
		parentImage, err = i.storageService.GetParentImage(opts.Name, opts.TenantOpts)
		if err != nil {
			if errors.Is(err, appsvc.ErrNoMatchingFile) {
				return nil, ErrNotFound
			}
			return nil, errors.New("internal error")
		}
		parentImageSpec, err = i.processorService.GetSpec(parentImage)
		if err != nil {
			return nil, errors.New("internal error")
		}
	}
	// determineDimensions
	targetWidth, targetHeight = determineDimensions(opts, parentImageSpec.Width, parentImageSpec.Height)
	// determineImageFormat
	if opts.Type == domain.ImageType_AUTO {
		targetImageFormat = parentImageSpec.Format
	} else {
		targetImageFormat = opts.Type
	}
	// fetch childImage
	childImage, err := i.storageService.GetChildImage(opts.Name, targetImageFormat, targetWidth, targetHeight, opts.TenantOpts)
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
		parentImageSpec, err = i.processorService.GetSpec(parentImage)
		if err != nil {
			return nil, errors.New("internal error")
		}
	}
	// buildImage then return
	scale := calculateScale(parentImageSpec.Width, parentImageSpec.Height, &targetWidth, &targetHeight)
	resizedImage, err := i.processorService.Resize(parentImage, scale)
	if err != nil {
		return nil, err
	}
	resizedImageSpec, err := i.processorService.GetSpec(resizedImage)
	if err != nil {
		return nil, err
	}
	var centeredImage []byte
	switch {
	case resizedImageSpec.Width == targetWidth && resizedImageSpec.Height == targetHeight:
		centeredImage = resizedImage
	case resizedImageSpec.Width == targetWidth:
		remainder := resizedImageSpec.Height - targetHeight
		// TODO: check if remainder < 0
		centeredImage, err = i.processorService.Crop(resizedImage, 0, remainder/2, resizedImageSpec.Width, targetHeight)
	case resizedImageSpec.Height == targetHeight:
		remainder := resizedImageSpec.Width - targetWidth
		// TODO: check if remainder < 0
		centeredImage, err = i.processorService.Crop(resizedImage, remainder/2, 0, targetWidth, resizedImageSpec.Height)
	}
	if err != nil {
		return nil, err
	}

	targetImage, err := i.processorService.Export(centeredImage, targetImageFormat)
	if err != nil {
		return nil, err
	}
	// cache image before return
	err = i.storageService.StoreChildImage(targetImage,
		opts.Name,
		domain.ImageSpec{
			Width:  targetWidth,
			Height: targetHeight,
			Format: targetImageFormat,
		},
		opts.TenantOpts,
	)
	if err != nil {
		return nil, err
	}

	return targetImage, nil
}

func NewImageService(storageSvc appsvc.ImageStorageServiceInterface,
	processorSvc appsvc.ImageProcessingServiceInterface) ImageServiceInterface {
	return ImageService{
		storageService:   storageSvc,
		processorService: processorSvc,
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
	gio.Type = format
	return gio
}

func NewServiceGetImageOpts() GetImageOpts {
	return GetImageOpts{}
}

func determineDimensions(opts GetImageOpts, originalDimensions ...int) (int, int) {
	var originalAr domain.AR
	var originalWidth, originalHeight int
	if originalDimensionsNeeded(opts) {
		originalWidth = originalDimensions[0]
		originalHeight = originalDimensions[1]
		originalAr = domain.NewAspectRatioFrom(originalWidth, originalHeight)
	}

	if opts.Ar != nil {
		if opts.Width != nil && opts.Height != nil {
			return *opts.Width, *opts.Height
		}
		if opts.Width != nil {
			return *opts.Width, heightFromAspectRatio(*opts.Width, *opts.Ar)
		}
		if opts.Height != nil {
			return widthFromAspectRatio(*opts.Height, *opts.Ar), *opts.Height
		}
		if originalAr.Float64() > opts.Ar.Float64() {
			return widthFromAspectRatio(originalHeight, *opts.Ar), originalHeight
		}
		return originalWidth, heightFromAspectRatio(originalWidth, *opts.Ar)
	}
	if opts.Width != nil && opts.Height != nil {
		return *opts.Width, *opts.Height
	}
	if opts.Width != nil {
		return *opts.Width, heightFromAspectRatio(*opts.Width, originalAr)
	}
	if opts.Height != nil {
		return widthFromAspectRatio(*opts.Height, originalAr), *opts.Height
	}
	return originalWidth, originalHeight
}

func heightFromAspectRatio(width int, ar domain.AR) int {
	return int(math.Round(float64(width) / ar.Float64()))
}

func widthFromAspectRatio(height int, ar domain.AR) int {
	return int(math.Round(float64(height) * ar.Float64()))
}

func calculateScale(originalWidth, originalHeight int, targetWidth, targetHeight *int) float64 {
	switch {
	case targetWidth != nil && targetHeight != nil:
		scaleWidth := float64(*targetWidth) / float64(originalWidth)
		scaleHeight := float64(*targetHeight) / float64(originalHeight)
		return math.Max(scaleWidth, scaleHeight)
	case targetWidth != nil:
		return float64(*targetWidth) / float64(originalWidth)
	case targetHeight != nil:
		return float64(*targetHeight) / float64(originalHeight)
	default:
		return 1.0
	}
}

func parentImageNeedsToBeFetched(opts GetImageOpts) bool {
	return opts.Type == domain.ImageType_AUTO || originalDimensionsNeeded(opts)
}

func originalDimensionsNeeded(opts GetImageOpts) bool {
	return (opts.Ar == nil && (opts.Width == nil || opts.Height == nil)) ||
		(opts.Ar != nil && (opts.Width == nil && opts.Height == nil))
}
