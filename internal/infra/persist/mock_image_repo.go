package persist

import (
	"context"
	domain2 "example.com/imageProc/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockImageRepo struct {
	mock.Mock
}

func (m *MockImageRepo) GetImage(ctx context.Context, opts domain2.RepoGetImageOpts) ([]byte, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageRepo) BuildImageOf(ctx context.Context, image []byte, opts domain2.BuildImageOpts) ([]byte, error) {
	args := m.Called(ctx, image, opts)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockImageRepo) CreateImage(ctx context.Context, image []byte, isParent bool, name string, opts domain2.TenantOpts) (string, error) {
	args := m.Called(ctx, image, isParent, name, opts)
	return args.String(0), args.Error(1)
}

func NewMockImageRepo() domain2.ImageRepoInterface {
	return &MockImageRepo{}
}
