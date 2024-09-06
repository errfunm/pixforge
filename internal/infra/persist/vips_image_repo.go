package persist

import (
	"context"
	"errors"
	"example.com/imageProc/internal/domain"
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"math"
)

type vipsImageRepo struct {
	baseDir string
}

func (i vipsImageRepo) BuildImageOf(ctx context.Context, image []byte, opts domain.BuildImageOpts) ([]byte, error) {
	imgRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		if errors.Is(err, vips.ErrUnsupportedImageFormat) {
			return nil, ErrUnSupportedImageFormat
		}
		return nil, err
	}
	originalWidth := imgRef.Width()
	originalHeight := imgRef.Height()
	var newWidth, newHeight int

	newWidth, newHeight = determineDimensions(opts.Width, opts.Height, opts.AspectRatio, originalWidth, originalHeight)

	scale := calculateScale(originalWidth, originalHeight, &newWidth, &newHeight)
	err = imgRef.Resize(scale, vips.KernelAuto)
	if err != nil {
		return nil, err
	}

	centeredImage, err := cropCenter(imgRef, newWidth, newHeight)
	if err != nil {
		return nil, err
	}

	var newImageFormat domain.ImageType
	if opts.ImageType == nil {
		newImageFormat, err = domain.ImageTypeFromString(imgRef.Format().FileExt())
		if err != nil {
			return nil, ErrInternal
		}
	} else {
		newImageFormat = *opts.ImageType
	}

	newImage, _, err := exportImage(centeredImage, newImageFormat)
	if err != nil {
		return nil, err
	}
	return newImage, nil
}

func NewVipsImageRepo(baseDir string) domain.ImageRepoInterface {
	return vipsImageRepo{
		baseDir: baseDir,
	}
}

func exportImage(imageRef *vips.ImageRef, imgType domain.ImageType) ([]byte, *vips.ImageMetadata, error) {
	var (
		image         []byte
		imageMetaData *vips.ImageMetadata
		err           error
	)

	switch imgType {
	case domain.ImageType_JPEG:
		image, imageMetaData, err = imageRef.ExportJpeg(vips.NewJpegExportParams())
	case domain.ImageType_WEBP:
		image, imageMetaData, err = imageRef.ExportWebp(vips.NewWebpExportParams())
	case domain.ImageType_AVIF:
		image, imageMetaData, err = imageRef.ExportAvif(vips.NewAvifExportParams())
	case domain.ImageType_PNG:
		image, imageMetaData, err = imageRef.ExportPng(vips.NewPngExportParams())
	}

	if err != nil {
		return nil, nil, err
	}
	return image, imageMetaData, nil
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

func cropCenter(image *vips.ImageRef, targetWidth, targetHeight int) (*vips.ImageRef, error) {
	var err error
	switch {
	case image.Width() == targetWidth && image.Height() == targetHeight:
		return image, nil
	case image.Width() == targetWidth:
		remainder := image.Height() - targetHeight
		// TODO: check if remainder < 0
		err = image.Crop(0, remainder/2, image.Width(), targetHeight)
	case image.Height() == targetHeight:
		remainder := image.Width() - targetWidth
		// TODO: check if remainder < 0
		err = image.Crop(remainder/2, 0, targetWidth, image.Height())
	default:
		err = fmt.Errorf("error while center cropping original: %d*%d, target:%d*%d", image.Width(),
			image.Height(), targetWidth, targetHeight)
	}
	if err != nil {
		return nil, err
	}
	return image, nil
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
