package persist

import (
	"context"
	"example.com/imageProc/app/util"
	"example.com/imageProc/domain"
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
		//imgType: domain.ImageType_JPEG,
		//width:   100,
		//ar:      domain.AR{Width: 3, Height: 4},
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
	t.Run("an image can be loaded with no error", func(t *testing.T) {
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
		img, err := os.ReadFile(baseDir + "/testdata/tstimg1.jpg")
		if err != nil {
			panic(err)
		}
		isParent := false
		tenantOpts := domain.TenantOpts{TenantCode: "tsttnt1", OrgCode: "tstorg1"}
		// call repo.CreateImage()
		imgName, err := repo.CreateImage(ctx, img, isParent, "newimg1", tenantOpts)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, "newimg1", imgName)
	})
}
