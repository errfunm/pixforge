package domain

import (
	"errors"
	"strconv"
	"strings"
)

type AR struct {
	Width  int
	Height int
}

func ParseAspectRatio(str string) (AR, error) {
	ar := AR{}
	arString := strings.Split(str, ":")
	if len(arString) != 2 {
		return AR{}, errors.New("not a valid aspect ratio")
	}
	for i := range arString {
		_d, err := strconv.Atoi(arString[0])
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
	ImageType_AVIF ImageType = iota
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
		return "jpg"
	case ImageType_PNG:
		return "png"
	default:
		return "unknown"
	}
}

func ImageTypeFromString(imgTypeStr string) (ImageType, error) {
	switch imgTypeStr {
	case "avif":
		return ImageType_AVIF, nil
	case "webp":
		return ImageType_WEBP, nil
	case "jpg":
		return ImageType_JPEG, nil
	case "png":
		return ImageType_PNG, nil
	default:
		return -1, errors.New("unknown image type")
	}
}

type TenantOpts struct {
	TenantCode string
	OrgCode    string
}
