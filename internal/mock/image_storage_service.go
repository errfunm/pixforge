package mock

import (
	"example.com/imageProc/internal/domain"
	"github.com/stretchr/testify/mock"
)

type ImageStorageService struct {
	mock.Mock
}

func (m *ImageStorageService) StoreParentImage(image []byte, format domain.ImageType, tenantOpts domain.TenantOpts) (string, error) {
	args := m.Called(image, format, tenantOpts)
	return args.String(0), args.Error(1)
}

func (m *ImageStorageService) StoreChildImage(image []byte, name string, spec domain.ImageSpec, tenantOpts domain.TenantOpts) error {
	args := m.Called(image, name, spec, tenantOpts)
	return args.Error(0)
}

func (m *ImageStorageService) GetParentImage(name string, tenantOpts domain.TenantOpts) ([]byte, error) {
	args := m.Called(name, tenantOpts)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *ImageStorageService) GetChildImage(name string, format domain.ImageType, width, height int, tenantOpts domain.TenantOpts) ([]byte, error) {
	args := m.Called(name, format, width, height, tenantOpts)
	return args.Get(0).([]byte), args.Error(1)
}
