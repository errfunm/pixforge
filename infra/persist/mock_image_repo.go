package persist

import (
	"context"
	"example.com/imageProc/domain"
	"github.com/stretchr/testify/mock"
)

type MockImageRepo struct {
	mock.Mock
}

func (m *MockImageRepo) GetImage(ctx context.Context, opts domain.GetImageOpts) ([]byte, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageRepo) BuildImageOf(ctx context.Context, image []byte, opts domain.BuildImageOpts) ([]byte, error) {
	args := m.Called(ctx, image, opts)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageRepo) CreateImage(ctx context.Context, image []byte, opts domain.TenantOpts) (string, error) {
	args := m.Called(ctx, image, opts)
	return args.String(0), args.Error(1)
}

func NewMockImageRepo() domain.ImageRepoInterface {
	return &MockImageRepo{}
}
