package persist

import "errors"

var (
	ErrImageNotFound          = errors.New("image not found")
	ErrUnSupportedImageFormat = errors.New("image format is unsupported")
	ErrInvalidDimensions      = errors.New("image with requested dimensions is invalid")
)
