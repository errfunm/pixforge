package domainsvc

import (
	"context"
	"example.com/imageProc/internal/domain"
	"example.com/imageProc/internal/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpload(t *testing.T) {
	t.Run("an image can be uploaded if it is valid", func(t *testing.T) {
		ctx := context.Background()
		img := []byte("valid image")
		imgName := "test_image_name"
		imgFormat := domain.ImageType_WEBP
		tenantOpts := domain.TenantOpts{}

		mockStorageSvc := new(mock.ImageStorageService)
		mockImageProcessingSvc := new(mock.ImageProcessingService)

		mockImageProcessingSvc.On("GetFormat", img).Return(imgFormat, nil)
		mockStorageSvc.On("StoreParentImage", img, imgFormat, tenantOpts).Return(imgName, nil)

		svc := NewImageService(mockStorageSvc, mockImageProcessingSvc)

		imgId, err := svc.Upload(ctx, img, tenantOpts)

		assert.NoError(t, err)
		assert.Equal(t, imgName, imgId)
		mockStorageSvc.AssertExpectations(t)
		mockImageProcessingSvc.AssertExpectations(t)
	})
}

func TestGetImage(t *testing.T) {
	t.Run("an image can be fetched if it exists", func(t *testing.T) {
		testCases := []struct {
			opts                     GetImageOpts
			isParentNeedsToBeFetched bool
			image                    []byte
		}{
			{
				opts: NewServiceGetImageOpts().
					SetName("testimagename1").
					SetFormat(domain.ImageType_JPEG).
					SetWidth(200).
					SetHeight(300),
				isParentNeedsToBeFetched: false,
				image:                    []byte("this is an image"),
			},
			{
				opts: NewServiceGetImageOpts().
					SetName("testimagename1").
					SetFormat(domain.ImageType_JPEG).
					SetWidth(200).
					SetAr(domain.AR{Width: 1, Height: 1}),
				isParentNeedsToBeFetched: false,
				image:                    []byte("this is an image"),
			},
			{
				opts: NewServiceGetImageOpts().
					SetName("testimagename1").
					SetFormat(domain.ImageType_JPEG).
					SetHeight(200).
					SetAr(domain.AR{Width: 1, Height: 1}),
				isParentNeedsToBeFetched: false,
				image:                    []byte("this is an image"),
			},
			{
				opts: NewServiceGetImageOpts().
					SetName("testimagename1").
					SetFormat(domain.ImageType_JPEG).
					SetWidth(200).
					SetHeight(300).
					SetAr(domain.AR{Width: 1, Height: 1}),
				isParentNeedsToBeFetched: false,
				image:                    []byte("this is an image"),
			},
			{
				opts: NewServiceGetImageOpts().
					SetName("testimagename1").
					SetWidth(200).
					SetHeight(300).
					SetAr(domain.AR{Width: 1, Height: 1}),
				isParentNeedsToBeFetched: true,
				image:                    []byte("this is an image"),
			},
			{
				opts: NewServiceGetImageOpts().
					SetName("testimagename1").
					SetFormat(domain.ImageType_JPEG),
				isParentNeedsToBeFetched: true,
				image:                    []byte("this is an image"),
			},
			{
				opts: NewServiceGetImageOpts().
					SetName("testimagename1").
					SetWidth(300).
					SetFormat(domain.ImageType_JPEG),
				isParentNeedsToBeFetched: true,
				image:                    []byte("this is an image"),
			},
			{
				opts: NewServiceGetImageOpts().
					SetName("testimagename1").
					SetHeight(300).
					SetFormat(domain.ImageType_JPEG),
				isParentNeedsToBeFetched: true,
				image:                    []byte("this is an image"),
			},
			{
				opts: NewServiceGetImageOpts().
					SetName("testimagename1").
					SetAr(domain.AR{Width: 1, Height: 1}).
					SetFormat(domain.ImageType_JPEG),
				isParentNeedsToBeFetched: true,
				image:                    []byte("this is an image"),
			},
		}
		mockStorageSvc := new(mock.ImageStorageService)
		mockImageProcessingSvc := new(mock.ImageProcessingService)

		for _, tc := range testCases {
			parentImage := []byte("this is the parent image")
			parentImageSpec := domain.ImageSpec{
				Width:  500,
				Height: 500,
				Format: domain.ImageType_AVIF,
			}
			var childImageFormat domain.ImageType
			if tc.opts.Type == nil {
				childImageFormat = parentImageSpec.Format
			} else {
				childImageFormat = *tc.opts.Type
			}

			normalizedWidth, normalizedHeight := determineDimensions(tc.opts, parentImageSpec.Width, parentImageSpec.Height)

			mockStorageSvc.On("GetParentImage", tc.opts.Name, tc.opts.TenantOpts).
				Return(parentImage, nil)

			mockImageProcessingSvc.On("GetSpec", parentImage).Return(parentImageSpec, nil)

			mockStorageSvc.On("GetChildImage", tc.opts.Name, childImageFormat, normalizedWidth, normalizedHeight,
				tc.opts.TenantOpts).Return(tc.image, nil)

			svc := NewImageService(mockStorageSvc, mockImageProcessingSvc)

			fetchedImage, err := svc.GetImage(context.Background(), tc.opts)

			assert.NoError(t, err)
			assert.Equal(t, tc.image, fetchedImage)

			if tc.isParentNeedsToBeFetched {
				mockStorageSvc.AssertExpectations(t)
				mockImageProcessingSvc.AssertExpectations(t)
			} else {
				mockStorageSvc.AssertNotCalled(t, "GetParentImage", tc.opts.Name, tc.opts.TenantOpts)
				mockImageProcessingSvc.AssertNotCalled(t, "GetSpec", parentImage)
			}
		}
	})
}

func TestDetermineDimensions(t *testing.T) {
	testCases := []struct {
		opts           GetImageOpts
		originalWidth  int
		originalHeight int
		expectedWidth  int
		expectedHeight int
	}{
		{
			opts:           NewServiceGetImageOpts().SetWidth(300).SetHeight(500).SetAr(domain.AR{Width: 1, Height: 1}),
			expectedWidth:  300,
			expectedHeight: 500,
		},
		{
			opts:           NewServiceGetImageOpts().SetWidth(300).SetAr(domain.AR{Width: 1, Height: 1}),
			expectedWidth:  300,
			expectedHeight: 300,
		},
		{
			opts:           NewServiceGetImageOpts().SetHeight(500).SetAr(domain.AR{Width: 1, Height: 1}),
			expectedWidth:  500,
			expectedHeight: 500,
		},
		{
			opts:           NewServiceGetImageOpts().SetWidth(300).SetHeight(500),
			expectedWidth:  300,
			expectedHeight: 500,
		},
		{
			opts:           NewServiceGetImageOpts().SetAr(domain.AR{Width: 3, Height: 4}),
			originalWidth:  500,
			originalHeight: 500,
			expectedWidth:  375,
			expectedHeight: 500,
		},
		{
			opts:           NewServiceGetImageOpts().SetWidth(340),
			originalWidth:  500,
			originalHeight: 500,
			expectedWidth:  340,
			expectedHeight: 340,
		},
		{
			opts:           NewServiceGetImageOpts().SetHeight(270),
			originalWidth:  900,
			originalHeight: 800,
			expectedWidth:  304,
			expectedHeight: 270,
		},
		{
			opts:           NewServiceGetImageOpts(),
			originalWidth:  1080,
			originalHeight: 2375,
			expectedWidth:  1080,
			expectedHeight: 2375,
		},
	}
	for _, tc := range testCases {
		w, h := determineDimensions(tc.opts, tc.originalWidth, tc.originalHeight)
		assert.Equal(t, tc.expectedWidth, w, tc)
		assert.Equal(t, tc.expectedHeight, h, tc)
	}
}

func TestParentImageNeedsToBeFetched(t *testing.T) {
	testCases := []struct {
		opts     GetImageOpts
		expected bool
	}{
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetHeight(0).
				SetAr(domain.AR{}),
			expected: false,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetWidth(0).
				SetAr(domain.AR{}),
			expected: false,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetWidth(0).
				SetHeight(0),
			expected: false,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetWidth(0).
				SetHeight(0).
				SetAr(domain.AR{}),
			expected: false,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetAr(domain.AR{}),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetHeight(0),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetWidth(0),
			expected: true,
		},

		{
			opts:     NewServiceGetImageOpts(),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetAr(domain.AR{}),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetHeight(0),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetWidth(0),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetHeight(0).
				SetAr(domain.AR{}),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetWidth(0).
				SetAr(domain.AR{}),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetWidth(0).
				SetHeight(0),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetWidth(0).
				SetHeight(0).
				SetAr(domain.AR{}),
			expected: true,
		},
	}
	for _, tc := range testCases {
		assert.Equal(t, tc.expected, parentImageNeedsToBeFetched(tc.opts), tc.opts)
	}
}

func TestParentDimensionsNeeded(t *testing.T) {
	testCases := []struct {
		opts     GetImageOpts
		expected bool
	}{
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetHeight(0).
				SetAr(domain.AR{}),
			expected: false,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetWidth(0).
				SetAr(domain.AR{}),
			expected: false,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetWidth(0).
				SetHeight(0),
			expected: false,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetWidth(0).
				SetHeight(0).
				SetAr(domain.AR{}),
			expected: false,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetAr(domain.AR{}),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetHeight(0),
			expected: true,
		},
		{
			opts: NewServiceGetImageOpts().
				SetFormat(domain.ImageType_JPEG).
				SetWidth(0),
			expected: true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, originalDimensionsNeeded(tc.opts), tc.opts)
	}
}
