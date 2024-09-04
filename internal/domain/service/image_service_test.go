package domainsvc

import (
	"context"
	domain2 "example.com/imageProc/internal/domain"
	persist2 "example.com/imageProc/internal/infra/persist"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpload(t *testing.T) {
	t.Run("an image can be uploaded if it is valid", func(t *testing.T) {
		ctx := context.Background()
		img := []byte("valid image")
		imgName := "test_image_name"
		opts := domain2.TenantOpts{}

		mockRepo := new(persist2.MockImageRepo)
		mockRepo.On("CreateImage", ctx, img, true, "", opts).Return(imgName, nil)

		svc := NewImageService(mockRepo)

		imgId, err := svc.Upload(ctx, img, opts)

		assert.NoError(t, err)
		assert.Equal(t, imgName, imgId)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetImage(t *testing.T) {
	t.Run("an image can be fetched if it exists", func(t *testing.T) {
		ctx := context.Background()
		img := []byte("a test image")
		imgName := "test_image_name"
		tenantOpts := domain2.TenantOpts{}

		repoGetImageOpts := domain2.RepoGetImageOpts{
			TenantOpts: tenantOpts,
			Name:       imgName,
		}

		mockRepo := new(persist2.MockImageRepo)
		mockRepo.On("GetImage", ctx, repoGetImageOpts).Return(img, nil)

		svc := NewImageService(mockRepo)

		fetchedImage, err := svc.GetImage(ctx, GetImageOpts{TenantOpts: tenantOpts, Name: imgName})

		assert.NoError(t, err)
		assert.Equal(t, img, fetchedImage)
		mockRepo.AssertExpectations(t)
	})

	t.Run("an error should be returned if no primary version of the requested image exist", func(t *testing.T) {
		ctx := context.Background()
		imgName := "test_image_name"
		tenantOpts := domain2.TenantOpts{}

		mockRepo := new(persist2.MockImageRepo)
		mockRepo.On("GetImage", ctx, domain2.RepoGetImageOpts{TenantOpts: tenantOpts, Name: imgName}).
			Return([]byte{}, persist2.ErrImageNotFound)
		mockRepo.On("GetImage", ctx, domain2.RepoGetImageOpts{TenantOpts: tenantOpts, Name: imgName, IsParent: true}).
			Return([]byte{}, persist2.ErrImageNotFound)

		svc := NewImageService(mockRepo)

		_, err := svc.GetImage(ctx, GetImageOpts{TenantOpts: tenantOpts, Name: imgName})

		assert.ErrorIs(t, err, ErrNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("an image can be built from its primary version if it does not exist", func(t *testing.T) {
		ctx := context.Background()
		imgName := "test_image_name"
		tenantOpts := domain2.TenantOpts{}
		primaryImg := []byte("this is the primary version of the image requested")
		builtImg := []byte("this is a brand new image built from primary version")

		mockRepo := new(persist2.MockImageRepo)
		mockRepo.On("GetImage", ctx, domain2.RepoGetImageOpts{TenantOpts: tenantOpts, Name: imgName}).
			Return([]byte{}, persist2.ErrImageNotFound)
		mockRepo.On("GetImage", ctx, domain2.RepoGetImageOpts{TenantOpts: tenantOpts, Name: imgName, IsParent: true}).
			Return(primaryImg, nil)
		mockRepo.On("BuildImageOf", ctx, primaryImg, domain2.BuildImageOpts{}).
			Return(builtImg, nil)
		mockRepo.On("CreateImage", ctx, builtImg, false, imgName, tenantOpts).
			Return(imgName, nil)

		svc := NewImageService(mockRepo)

		fetchedImg, err := svc.GetImage(ctx, GetImageOpts{TenantOpts: tenantOpts, Name: imgName})

		assert.NoError(t, err)
		assert.Equal(t, builtImg, fetchedImg)
		mockRepo.AssertExpectations(t)

	})
}