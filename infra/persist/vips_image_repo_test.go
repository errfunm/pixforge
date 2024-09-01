package persist

import (
	"context"
	"example.com/imageProc/app/util"
	"example.com/imageProc/domain"
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	baseDir     = "/home/errfunm/Projects/Go/imageProc"
	testEnviron = baseDir + "/test"
)

var testImages = []struct {
	tenantOpts domain.TenantOpts
	baseDir    string
	name       string
	isParent   bool
	imgType    domain.ImageType
	width      int
	ar         domain.AR
}{
	{
		tenantOpts: domain.TenantOpts{TenantCode: "tsttnt1", OrgCode: "tstorg1"},
		baseDir:    baseDir,
		name:       "tstimg1",
		isParent:   false,
		imgType:    domain.ImageType_JPEG,
		width:      100,
		ar:         domain.AR{Width: 3, Height: 4},
	},
	{
		tenantOpts: domain.TenantOpts{TenantCode: "tsttnt2", OrgCode: "tstorg2"},
		baseDir:    baseDir,
		name:       "tstimg2",
		isParent:   true,
		imgType:    domain.ImageType_AVIF,
	},
	{
		tenantOpts: domain.TenantOpts{TenantCode: "tsttnt2", OrgCode: "tstorg2"},
		baseDir:    baseDir,
		name:       "tstimg2",
		isParent:   false,
		imgType:    domain.ImageType_AVIF,
		width:      500,
		ar:         domain.AR{Width: 1, Height: 1},
	},
	{
		tenantOpts: domain.TenantOpts{TenantCode: "tsttnt2", OrgCode: "tstorg2"},
		baseDir:    baseDir,
		name:       "tstimg3",
		isParent:   true,
		imgType:    domain.ImageType_WEBP,
	},
}

func initTestEnvironment() error {
	for _, ti := range testImages {
		toBeLoadedImage, err := os.ReadFile(util.FullImageAddr(baseDir+"/testdata", ti.name, ti.imgType.String()))
		if err != nil {
			return err
		}

		storageDir := util.ResolveStoragePath(testEnviron, ti.tenantOpts, ti.name, ti.isParent, util.ChildPathOpts{
			ImgType:  ti.imgType,
			ImgAR:    ti.ar,
			ImgWidth: ti.width,
		})
		if err = os.MkdirAll(storageDir, 0750); err != nil {
			return err
		}

		addrToBeWritten := util.FullImageAddr(storageDir, ti.name, ti.imgType.String())
		if err = os.WriteFile(addrToBeWritten, toBeLoadedImage, 0666); err != nil {
			return err
		}
	}
	return nil
}

func tearDownTestEnvironment() error {
	return os.RemoveAll(baseDir + "/test")
}

func TestCreateImage(t *testing.T) {
	t.Run("an image can be created with no error", func(t *testing.T) {
		err := initTestEnvironment()
		defer func() {
			err = tearDownTestEnvironment()
			if err != nil {

			}
		}()
		if err != nil {
			panic(err)
		}

		repo := NewVipsImageRepo(testEnviron)

		ctx := context.Background()
		testCases := []struct {
			tenantOpts domain.TenantOpts
			imgName    string
			isParent   bool
			imgType    domain.ImageType
		}{
			{
				tenantOpts: domain.TenantOpts{TenantCode: "tsttnt1", OrgCode: "tstorg1"},
				imgName:    "tstimg1",
				isParent:   false,
				imgType:    domain.ImageType_JPEG,
			},
			//{
			//	tenantOpts: domain.TenantOpts{TenantCode: "tsttnt4", OrgCode: "tstorg4"},
			//	imgName:    "tstimg2",
			//	isParent:   true,
			//	imgType:    domain.ImageType_AVIF,
			//},
		}

		for _, tc := range testCases {
			imageRef, err := vips.NewImageFromFile(baseDir + "/testdata" + "/" + tc.imgName + "." + tc.imgType.String())
			if err != nil {
				panic(err)
			}
			img, _, err := exportImage(imageRef, tc.imgType)
			if err != nil {
				panic(err)
			}

			expectedImageStoragePath := util.ResolveStoragePath(testEnviron, tc.tenantOpts, tc.imgName, tc.isParent, util.ChildPathOpts{
				ImgType:  tc.imgType,
				ImgWidth: imageRef.Width(),
				ImgAR:    domain.NewAspectRatioFrom(imageRef.Width(), imageRef.Height()),
			})

			_, err = repo.CreateImage(ctx, img, tc.isParent, tc.imgName, tc.tenantOpts)

			assert.NoError(t, err)
			dirEntry, err := os.ReadDir(expectedImageStoragePath)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			assert.Equal(t, tc.imgName+"."+tc.imgType.String(), dirEntry[0].Name())
		}

	})

	t.Run("an error should be returned if the image is invalid", func(t *testing.T) {

	})
}

func TestGetImage(t *testing.T) {
	t.Run("an image can be fetched if it exists", func(t *testing.T) {
		err := initTestEnvironment()
		defer func() {
			err = tearDownTestEnvironment()
			if err != nil {

			}
		}()
		if err != nil {
			panic(err)
		}

		testCases := []struct {
			imgName          string
			imgType          domain.ImageType
			imgWidth         int
			imgAspectRatio   domain.AR
			isParent         bool
			tenantOpts       domain.TenantOpts
			expectedImageDir string
		}{
			{
				tenantOpts:       domain.TenantOpts{TenantCode: "tsttnt1", OrgCode: "tstorg1"},
				imgName:          "tstimg1",
				isParent:         false,
				imgType:          domain.ImageType_JPEG,
				imgWidth:         100,
				imgAspectRatio:   domain.AR{Width: 3, Height: 4},
				expectedImageDir: fmt.Sprintf("%s/%s-%s/%s/%s/%s/%s", testEnviron, "tsttnt1", "tstorg1", "tstimg1", "jpeg", "3:4", "100"),
			},
			{
				tenantOpts:       domain.TenantOpts{TenantCode: "tsttnt2", OrgCode: "tstorg2"},
				imgName:          "tstimg2",
				isParent:         true,
				imgType:          domain.ImageType_AVIF,
				expectedImageDir: fmt.Sprintf("%s/%s-%s/%s", testEnviron, "tsttnt2", "tstorg2", "tstimg2"),
			},
			{
				tenantOpts:       domain.TenantOpts{TenantCode: "tsttnt2", OrgCode: "tstorg2"},
				imgName:          "tstimg2",
				isParent:         false,
				imgType:          domain.ImageType_AVIF,
				imgWidth:         500,
				imgAspectRatio:   domain.AR{Width: 1, Height: 1},
				expectedImageDir: fmt.Sprintf("%s/%s-%s/%s/%s/%s/%s", testEnviron, "tsttnt2", "tstorg2", "tstimg2", "avif", "1:1", "500"),
			},
		}

		repo := NewVipsImageRepo(testEnviron)

		for _, tc := range testCases {
			image, err := repo.GetImage(context.Background(), domain.RepoGetImageOpts{
				TenantOpts:  tc.tenantOpts,
				Name:        tc.imgName,
				IsParent:    tc.isParent,
				Type:        tc.imgType,
				Width:       tc.imgWidth,
				AspectRatio: tc.imgAspectRatio,
			})
			if err != nil {
				return
			}
			assert.NoError(t, err)
			//expectedImage, _ := os.ReadFile(tc.expectedImageDir + "/" + tc.imgName + "." + tc.imgType.String())
			imageRef, err := vips.NewImageFromFile(tc.expectedImageDir + "/" + tc.imgName + "." + tc.imgType.String())
			if err != nil {
				panic(err)
			}
			expectedImage, _, err := exportImage(imageRef, tc.imgType)
			if err != nil {
				panic(err)
			}
			assert.Equal(t, expectedImage, image)
		}
	})

	t.Run("an error should be returned if it does not exist", func(t *testing.T) {
		err := initTestEnvironment()
		defer func() {
			err = tearDownTestEnvironment()
			if err != nil {

			}
		}()
		if err != nil {
			panic(err)
		}
		testCases := []struct {
			imgName        string
			imgType        domain.ImageType
			imgWidth       int
			imgAspectRatio domain.AR
			isParent       bool
			tenantOpts     domain.TenantOpts
			expectedError  error
		}{
			{
				tenantOpts:     domain.TenantOpts{TenantCode: "tsttnt1", OrgCode: "tstorg1"},
				imgName:        "tstimg1",
				isParent:       false,
				imgType:        domain.ImageType_JPEG,
				imgWidth:       200, // actual image width 100
				imgAspectRatio: domain.AR{Width: 3, Height: 4},
				expectedError:  ErrImageNotFound,
			},
			{
				tenantOpts:    domain.TenantOpts{TenantCode: "tsttnt2", OrgCode: "tstorg2"},
				imgName:       "tstimg", // actual image's type is tstimg
				isParent:      true,
				expectedError: ErrImageNotFound,
			},
			{
				tenantOpts:     domain.TenantOpts{TenantCode: "tsttnt3", OrgCode: "tstorg3"}, // actual image's = tsttnt2
				imgName:        "tstimg2",
				isParent:       false,
				imgType:        domain.ImageType_AVIF,
				imgWidth:       500,
				imgAspectRatio: domain.AR{Width: 1, Height: 1},
				expectedError:  ErrImageNotFound,
			},
		}

		repo := NewVipsImageRepo(testEnviron)
		for _, tc := range testCases {
			_, err = repo.GetImage(context.Background(), domain.RepoGetImageOpts{
				TenantOpts:  tc.tenantOpts,
				Name:        tc.imgName,
				IsParent:    tc.isParent,
				Type:        tc.imgType,
				Width:       tc.imgWidth,
				AspectRatio: tc.imgAspectRatio,
			})
			assert.ErrorIs(t, err, tc.expectedError)
		}
	})
}

func TestBuildImageOf(t *testing.T) {
	t.Run("a new image can be built if the input is a valid image", func(t *testing.T) {
		err := initTestEnvironment()
		defer func() {
			err := tearDownTestEnvironment()
			if err != nil {

			}
		}()
		if err != nil {
			panic(err)
		}

		testCases := []struct {
			imgWidth       int
			imgAspectRatio domain.AR
			imgType        domain.ImageType
			path           string
		}{
			{
				imgWidth:       275,
				imgAspectRatio: domain.AR{Width: 275, Height: 183},
				imgType:        domain.ImageType_WEBP,
				path: util.FullImageAddr(
					util.ResolveStoragePath(
						testEnviron,
						testImages[0].tenantOpts,
						testImages[0].name,
						testImages[0].isParent,
						util.ChildPathOpts{
							ImgType:  testImages[0].imgType,
							ImgWidth: testImages[0].width,
							ImgAR:    testImages[0].ar,
						}),
					testImages[0].name,
					testImages[0].imgType.String(),
				),
			},
			{
				imgWidth:       4623,
				imgAspectRatio: domain.AR{Width: 1541, Height: 1024},
				imgType:        domain.ImageType_AVIF,
				path: util.FullImageAddr(
					util.ResolveStoragePath(
						testEnviron,
						testImages[1].tenantOpts,
						testImages[1].name,
						testImages[1].isParent,
						util.ChildPathOpts{
							ImgType:  testImages[1].imgType,
							ImgWidth: testImages[1].width,
							ImgAR:    testImages[1].ar,
						}),
					testImages[1].name,
					testImages[1].imgType.String(),
				),
			},
		}

		repo := NewVipsImageRepo(testEnviron)

		for _, tc := range testCases {
			// read an image from testenviron
			image, err := os.ReadFile(tc.path)
			if err != nil {
				panic(err)
			}
			// call repo.BuildImageOf()
			builtImage, err := repo.BuildImageOf(context.Background(), image, domain.BuildImageOpts{
				ImageType:   tc.imgType,
				Width:       tc.imgWidth,
				AspectRatio: tc.imgAspectRatio,
			})
			// assert NoError
			assert.NoError(t, err)
			// read with vips
			builtImageRef, err := vips.NewImageFromBuffer(builtImage)
			if err != nil {
				panic(err)
			}
			// assert type/width/ar
			assert.Equal(t, tc.imgWidth, builtImageRef.Width())
			assert.Equal(t, tc.imgAspectRatio, domain.NewAspectRatioFrom(builtImageRef.Width(), builtImageRef.Height()))
			builtImageType, err := domain.ImageTypeFromString(builtImageRef.Format().FileExt())
			if err != nil {
				panic(err)
			}
			assert.Equal(t, tc.imgType, builtImageType)
		}
	})

	t.Run("an image with the given width and aspect ratio has to be possible", func(t *testing.T) {
		err := initTestEnvironment()
		defer func() {
			err := tearDownTestEnvironment()
			if err != nil {

			}
		}()
		if err != nil {
			panic(err)
		}

		path := util.FullImageAddr(
			util.ResolveStoragePath(
				testEnviron,
				testImages[3].tenantOpts,
				testImages[3].name,
				testImages[3].isParent,
				util.ChildPathOpts{
					ImgType:  testImages[3].imgType,
					ImgWidth: testImages[3].width,
					ImgAR:    testImages[3].ar,
				}),
			testImages[3].name,
			testImages[3].imgType.String(),
		)
		image, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}

		testCases := []struct {
			width int
			ar    domain.AR
		}{
			{width: 200, ar: domain.AR{Width: 275, Height: 183}},
			{width: 100, ar: domain.AR{Width: 3, Height: 4}},
		}

		repo := NewVipsImageRepo(testEnviron)

		for _, tc := range testCases {
			_, err = repo.BuildImageOf(context.Background(), image, domain.BuildImageOpts{Width: tc.width, AspectRatio: tc.ar})
			assert.ErrorIs(t, err, ErrInvalidDimensions)
		}
	})

	t.Run("an error should be returned if the given image is not valid", func(t *testing.T) {
		testCases := []struct {
			invalidImage []byte
		}{
			{invalidImage: []byte("fksjkfjogheohorhigdjfoigdfjgjdgjoisr")},
			{invalidImage: []byte("invalid image byte")},
		}

		repo := NewVipsImageRepo(testEnviron)

		for _, tc := range testCases {
			_, err := repo.BuildImageOf(context.Background(), tc.invalidImage, domain.BuildImageOpts{
				ImageType:   domain.ImageType_AVIF,
				Width:       100,
				AspectRatio: domain.AR{Width: 3, Height: 4},
			})
			assert.ErrorIs(t, err, ErrUnSupportedImageFormat)
		}
	})
}
