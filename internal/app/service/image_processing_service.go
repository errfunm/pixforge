package appsvc

import (
	"errors"
	"example.com/imageProc/internal/domain"
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
)

type ImageProcessingServiceInterface interface {
	GetWidth(image []byte) (int, error)
	GetHeight(image []byte) (int, error)
	GetFormat(image []byte) (domain.ImageType, error)
	GetSpec(image []byte) (domain.ImageSpec, error)
	Crop(image []byte, left, top, width, height int) ([]byte, error)
	Resize(image []byte, scale float64) ([]byte, error)
	Export(image []byte, imageType domain.ImageType) ([]byte, error)
}

var ErrUnsupportedImageFormat = errors.New("unsupported image format")

type VipsImageProcessorService struct{}

func (v VipsImageProcessorService) GetWidth(image []byte) (int, error) {
	imageRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		if errors.Is(err, vips.ErrUnsupportedImageFormat) {
			return 0, errors.New("unsupported image format")
		}
		return 0, fmt.Errorf("internal error: %v", err)
	}
	return imageRef.Width(), nil
}

func (v VipsImageProcessorService) GetHeight(image []byte) (int, error) {
	imageRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		if errors.Is(err, vips.ErrUnsupportedImageFormat) {
			return 0, errors.New("unsupported image format")
		}
		return 0, fmt.Errorf("internal error: %v", err)
	}
	return imageRef.Height(), nil
}

func (v VipsImageProcessorService) GetFormat(image []byte) (domain.ImageType, error) {
	imageRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		if errors.Is(err, vips.ErrUnsupportedImageFormat) {
			return 0, ErrUnsupportedImageFormat
		}
		return 0, fmt.Errorf("internal error: %v", err)
	}
	format, err := domain.ImageTypeFromString(imageRef.Format().FileExt())
	if err != nil {
		return -1, err
	}
	return format, nil
}

func (v VipsImageProcessorService) GetSpec(image []byte) (domain.ImageSpec, error) {
	imageRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		if errors.Is(err, vips.ErrUnsupportedImageFormat) {
			return domain.ImageSpec{}, errors.New("unsupported image format")
		}
		return domain.ImageSpec{}, fmt.Errorf("internal error: %v", err)
	}
	format, err := domain.ImageTypeFromString(imageRef.Format().FileExt())
	if err != nil {
		return domain.ImageSpec{}, err
	}
	return domain.ImageSpec{
		Width:  imageRef.Width(),
		Height: imageRef.Height(),
		Format: format,
	}, nil
}

func (v VipsImageProcessorService) Crop(image []byte, left, top, width, height int) ([]byte, error) {
	imageRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		if errors.Is(err, vips.ErrUnsupportedImageFormat) {
			return nil, errors.New("unsupported image format")
		}
		return nil, fmt.Errorf("internal error: %v", err)
	}
	if err = imageRef.Crop(left, top, width, height); err != nil {
		return nil, err
	}
	format, err := domain.ImageTypeFromString(imageRef.Format().FileExt())
	if err != nil {
		return nil, err
	}
	image, err = exportImage(imageRef, format)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (v VipsImageProcessorService) Resize(image []byte, scale float64) ([]byte, error) {
	imageRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		if errors.Is(err, vips.ErrUnsupportedImageFormat) {
			return nil, errors.New("unsupported image format")
		}
		return nil, fmt.Errorf("internal error: %v", err)
	}
	if err = imageRef.Resize(scale, vips.KernelAuto); err != nil {
		return nil, err
	}
	format, err := domain.ImageTypeFromString(imageRef.Format().FileExt())
	if err != nil {
		return nil, err
	}
	image, err = exportImage(imageRef, format)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (v VipsImageProcessorService) Export(image []byte, format domain.ImageType) ([]byte, error) {
	imageRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		if errors.Is(err, vips.ErrUnsupportedImageFormat) {
			return nil, errors.New("unsupported image format")
		}
		return nil, fmt.Errorf("internal error: %v", err)
	}
	image, err = exportImage(imageRef, format)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func NewVipsImageProcessorService() ImageProcessingServiceInterface {
	return VipsImageProcessorService{}
}

func exportImage(imageRef *vips.ImageRef, imgType domain.ImageType) ([]byte, error) {
	var (
		image []byte
		err   error
	)

	switch imgType {
	case domain.ImageType_JPEG:
		image, _, err = imageRef.ExportJpeg(vips.NewJpegExportParams())
	case domain.ImageType_WEBP:
		image, _, err = imageRef.ExportWebp(vips.NewWebpExportParams())
	case domain.ImageType_AVIF:
		image, _, err = imageRef.ExportAvif(vips.NewAvifExportParams())
	case domain.ImageType_PNG:
		image, _, err = imageRef.ExportPng(vips.NewPngExportParams())
	}
	if err != nil {
		return nil, err
	}
	return image, nil
}
