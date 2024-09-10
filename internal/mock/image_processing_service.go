package mock

import (
	"example.com/imageProc/internal/domain"
	"github.com/stretchr/testify/mock"
)

type ImageProcessingService struct {
	mock.Mock
}

func (m *ImageProcessingService) GetWidth(image []byte) (int, error) {
	args := m.Called(image)
	return args.Int(0), args.Error(1)
}

func (m *ImageProcessingService) GetHeight(image []byte) (int, error) {
	args := m.Called(image)
	return args.Int(0), args.Error(1)
}

func (m *ImageProcessingService) GetFormat(image []byte) (domain.ImageType, error) {
	args := m.Called(image)
	return args.Get(0).(domain.ImageType), args.Error(1)
}

func (m *ImageProcessingService) GetSpec(image []byte) (domain.ImageSpec, error) {
	args := m.Called(image)
	return args.Get(0).(domain.ImageSpec), args.Error(1)
}

func (m *ImageProcessingService) Crop(image []byte, left, top, width, height int) ([]byte, error) {
	args := m.Called(image, left, top, width, height)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *ImageProcessingService) Resize(image []byte, scale float64) ([]byte, error) {
	args := m.Called(image, scale)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *ImageProcessingService) Export(image []byte, imageType domain.ImageType) ([]byte, error) {
	args := m.Called(image, imageType)
	return args.Get(0).([]byte), args.Error(1)
}
