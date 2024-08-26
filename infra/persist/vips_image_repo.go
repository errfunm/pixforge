package persist

import (
	"context"
	"errors"
	"example.com/imageProc/app/util"
	"example.com/imageProc/domain"
	"github.com/davidbyttow/govips/v2/vips"
	"os"
)

type vipsImageRepo struct {
	baseDir string
}

func (i vipsImageRepo) GetImage(ctx context.Context, opts domain.GetImageOpts) ([]byte, error) {
	vips.Startup(nil)
	defer vips.Shutdown()

	var path string
	if opts.IsParent {
		path = util.ResolveStoragePath(i.baseDir, opts.TenantOpts, opts.Name, true, util.ChildPathOpts{})
	} else {
		path = util.ResolveStoragePath(i.baseDir, opts.TenantOpts, opts.Name, false, util.ChildPathOpts{
			ImgType:  opts.Type,
			ImgAR:    opts.AspectRatio,
			ImgWidth: opts.Width,
		})
	}

	fullPath := util.FullImageAddr(path, opts.Name, opts.Type.String())
	imgRef, err := vips.NewImageFromFile(fullPath)
	if err != nil {
		return nil, err
	}
	bytes, err := imgRef.ToBytes()
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (i vipsImageRepo) BuildImageOf(ctx context.Context, image []byte, opts domain.BuildImageOpts) ([]byte, error) {
	vips.Startup(nil)
	defer vips.Shutdown()

	imgRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		return nil, err
	}
	// check if the aspect ratios are the same
	if domain.NewAspectRatioFrom(imgRef.Width(), imgRef.Height()) == opts.AspectRatio {
		scale := float64(opts.Width) / float64(imgRef.Width())
		if err = imgRef.Resize(scale, vips.KernelAuto); err != nil {
			return nil, err
		}
		builtImage, _, err := exportImage(imgRef, opts.ImageType)
		if err != nil {
			return nil, err
		}
		return builtImage, nil
	}
	// if not : do combination of resize and crop
	h := (opts.Width / opts.AspectRatio.Width) * opts.AspectRatio.Height

	if opts.Width >= h {
		s := float64(opts.Width) / float64(imgRef.Width())
		err = imgRef.Resize(s, vips.KernelAuto)
		if err != nil {
			return nil, err
		}

		if h > imgRef.Height() {
			return nil, errors.New("not enough to crop from height")
		}
		remainderHeight := imgRef.Height() - h
		err = imgRef.Crop(0, remainderHeight/2, imgRef.Width(), h)
		if err != nil {
			return nil, err
		}
	} else {
		s := float64(h) / float64(imgRef.Height())
		err = imgRef.Resize(s, vips.KernelAuto)
		if err != nil {
			return nil, err
		}

		if opts.Width > imgRef.Width() {
			return nil, errors.New("not enough to crop from the width")
		}
		remainderWidth := imgRef.Width() - opts.Width
		err = imgRef.Crop(remainderWidth/2, 0, opts.Width, imgRef.Height())
	}

	builtImage, _, err := exportImage(imgRef, opts.ImageType)
	if err != nil {
		return nil, err
	}
	return builtImage, nil
}

func (i vipsImageRepo) CreateImage(ctx context.Context, image []byte, isParent bool, name string, opts domain.TenantOpts) (string, error) {
	var imgName, path string
	vips.Startup(nil)
	defer vips.Shutdown()

	imgRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		return "", err
	}

	imgType, err := domain.ImageTypeFromString(imgRef.Format().FileExt())
	if err != nil {
		return "", err
	}

	if isParent {
		imgName = util.GenerateImageName()
		path = util.ResolveStoragePath(i.baseDir, opts, imgName, true, util.ChildPathOpts{})
	} else {
		imgName = name
		ar := domain.NewAspectRatioFrom(imgRef.Width(), imgRef.Height())
		path = util.ResolveStoragePath(i.baseDir, opts, imgName, false, util.ChildPathOpts{
			ImgType:  imgType,
			ImgAR:    ar,
			ImgWidth: imgRef.Width(),
		})
	}

	if err = os.MkdirAll(path, 0750); err != nil {
		return "", err
	}
	err = os.WriteFile(util.FullImageAddr(path, imgName, imgType.String()), image, 0666)
	if err != nil {
		return "", err
	}
	return imgName, nil
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
