package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type AR struct {
	Width  int
	Height int
}

func (ar AR) String() string {
	return fmt.Sprintf("%d:%d", ar.Width, ar.Height)
}

func (ar AR) Float64() float64 {
	return float64(ar.Width) / float64(ar.Height)
}

func NewAspectRatioFrom(width int, height int) AR {
	if width == 0 || height == 0 {
		return AR{Width: 0, Height: 0}
	}

	gcf := GreatCommonFactor(width, height)

	return AR{
		Width:  width / gcf,
		Height: height / gcf,
	}
}

func ParseAspectRatio(str string) (AR, error) {
	ar := AR{}
	arString := strings.Split(str, ":")
	if len(arString) != 2 {
		return AR{}, errors.New("not a valid aspect ratio")
	}
	for i := range arString {
		_d, err := strconv.Atoi(arString[i])
		if err != nil {
			return AR{}, errors.New("not a valid aspect ratio")
		}
		if i == 0 {
			ar.Width = _d
		}
		ar.Height = _d
	}
	return ar, nil
}

type ImageType int

const (
	ImageType_AUTO ImageType = iota
	ImageType_AVIF
	ImageType_WEBP
	ImageType_JPEG
	ImageType_PNG
)

func (imgT ImageType) String() string {
	switch imgT {
	case ImageType_AVIF:
		return "avif"
	case ImageType_WEBP:
		return "webp"
	case ImageType_JPEG:
		return "jpeg"
	case ImageType_PNG:
		return "png"
	case ImageType_AUTO:
		return "auto"
	default:
		return "unknown"
	}
}

func ImageTypeFromString(imgTypeStr string) (ImageType, error) {
	switch imgTypeStr {
	case "avif":
		return ImageType_AVIF, nil
	case ".avif":
		return ImageType_AVIF, nil
	case "webp":
		return ImageType_WEBP, nil
	case ".webp":
		return ImageType_WEBP, nil
	case "jpg":
		return ImageType_JPEG, nil
	case ".jpg":
		return ImageType_JPEG, nil
	case "jpeg":
		return ImageType_JPEG, nil
	case ".jpeg":
		return ImageType_JPEG, nil
	case "png":
		return ImageType_PNG, nil
	case ".png":
		return ImageType_PNG, nil
	case "auto":
		return ImageType_AUTO, nil
	default:
		return -1, fmt.Errorf("unsupported image format: %v", imgTypeStr)
	}
}

type TenantOpts struct {
	TenantCode string
	OrgCode    string
}

func GreatCommonFactor(a int, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

type ImageSpec struct {
	Width  int
	Height int
	Format ImageType
}
