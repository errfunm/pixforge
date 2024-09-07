package appsvc

import (
	"errors"
	"example.com/imageProc/internal/domain"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

var (
	testEnvironBaseDir, testdataDir string
)

func TestParentImageDir(t *testing.T) {
	testCases := []struct {
		baseUrl    string
		tenantOpts domain.TenantOpts
		name       string
		expected   string
	}{
		{
			baseUrl:    "test",
			tenantOpts: domain.TenantOpts{TenantCode: "tsttnt", OrgCode: "tstorg"},
			name:       "npeirenw",
			expected:   "test/tsttnt-tstorg/npeirenw",
		},
		{
			baseUrl:    "a/b/c",
			tenantOpts: domain.TenantOpts{TenantCode: "tsttnt", OrgCode: "tstorg"},
			name:       "kkwooqieom",
			expected:   "a/b/c/tsttnt-tstorg/kkwooqieom",
		},
	}

	for _, tc := range testCases {
		res := parentImageDir(tc.baseUrl, tc.tenantOpts, tc.name)
		assert.Equal(t, tc.expected, res)
	}
}

func TestChildImageDir(t *testing.T) {
	testCases := []struct {
		parentUrl     string
		format        domain.ImageType
		width, height int
		expected      string
	}{
		{
			parentUrl: "a/b/c/tsttnt-tstorg/testImageName",
			format:    domain.ImageType_AVIF,
			width:     100,
			height:    200,
			expected:  "a/b/c/tsttnt-tstorg/testImageName/avif/100/200",
		},
		{
			parentUrl: "parentUrl",
			format:    domain.ImageType_AVIF,
			width:     300,
			height:    200,
			expected:  "parentUrl/avif/300/200",
		},
	}
	for _, tc := range testCases {
		res := childImageDir(tc.parentUrl, tc.format, tc.width, tc.height)
		assert.Equal(t, tc.expected, res)
	}
}

var initTestEnvironStatus = []struct {
	fileName   string
	tenantOpts domain.TenantOpts
	name       string
	isParent   bool
	width      int
	height     int
	format     domain.ImageType
	storedDir  string
}{
	{
		fileName:   "tstimg1.jpeg",
		tenantOpts: domain.TenantOpts{TenantCode: "tenant1", OrgCode: "org1"},
		name:       "initTestEnviron1",
		isParent:   false,
		width:      500,
		height:     200,
		format:     domain.ImageType_JPEG,
	},
	{
		fileName:   "tstimg2.avif",
		tenantOpts: domain.TenantOpts{TenantCode: "tenant2", OrgCode: "org2"},
		name:       "initTestEnviron2",
		isParent:   true,
	},
	{
		fileName:   "tstimg3.webp",
		tenantOpts: domain.TenantOpts{TenantCode: "tenant3", OrgCode: "org3"},
		name:       "initTestEnviron3",
		isParent:   false,
		width:      275,
		height:     183,
		format:     domain.ImageType_AVIF,
	},
}

func initTestEnvironment() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	testEnvironBaseDir = filepath.Join(wd, "testEnviron")
	testdataDir = filepath.Join(wd, "testdata")

	for i, img := range initTestEnvironStatus {
		image, err := os.ReadFile(filepath.Join(testdataDir, img.fileName))
		if err != nil {
			return err
		}
		var dir string
		parentDir := parentImageDir(testEnvironBaseDir, img.tenantOpts, img.name)
		if !img.isParent {
			dir = childImageDir(parentDir, img.format, img.width, img.height)
		} else {
			dir = parentDir
		}

		if err = os.MkdirAll(dir, 0750); err != nil {
			return err
		}
		fileName := filepath.Join(dir, img.name+"."+img.format.String())
		if err = os.WriteFile(fileName, image, 0666); err != nil {
			return err
		}
		initTestEnvironStatus[i].storedDir = fileName
	}
	return nil
}

func tearDownTestEnvironment() error {
	err := os.RemoveAll(testEnvironBaseDir)
	if err != nil {
		return err
	}
	return nil
}

func TestInitTestEnviron(t *testing.T) {
	err := initTestEnvironment()
	assert.NoError(t, err)
}

func TestStoreParentImage(t *testing.T) {
	t.Run("an image can be stored with no error", func(t *testing.T) {
		if err := initTestEnvironment(); err != nil {
			t.Fatalf("error initializing test environment: %v", err)
		}
		defer func() {
			err := tearDownTestEnvironment()
			if err != nil {
				panic(err)
			}
		}()

		testCases := []struct {
			tenantOpts          domain.TenantOpts
			format              domain.ImageType
			toBeLoadedImageName string
		}{
			{
				tenantOpts:          domain.TenantOpts{TenantCode: "umoitj93", OrgCode: "ownlqz"},
				format:              domain.ImageType_AVIF,
				toBeLoadedImageName: "tstimg1.jpeg",
			},
		}

		liss := NewLocalImageStorageService(testEnvironBaseDir)

		for _, tc := range testCases {
			image, err := os.ReadFile(filepath.Join(testdataDir, tc.toBeLoadedImageName))
			if err != nil {
				t.Fatalf("error reading image from testDataDir: %v", err)
			}

			parentImageName, err := liss.StoreParentImage(image, tc.format, tc.tenantOpts)

			assert.NoError(t, err)

			_, err = os.ReadFile(filepath.Join(
				parentImageDir(testEnvironBaseDir, tc.tenantOpts, parentImageName),
				parentImageName+"."+tc.format.String()))
			if errors.Is(err, os.ErrNotExist) {
				t.Errorf("expected an image to be found but got: %v", err)
			}
		}

	})
}

func TestStoreChildImage(t *testing.T) {
	t.Run("an image can be stored with no error", func(t *testing.T) {
		if err := initTestEnvironment(); err != nil {
			t.Fatalf("error initializing test environment: %v", err)
		}
		defer func() {
			err := tearDownTestEnvironment()
			if err != nil {
				panic(err)
			}
		}()

		testCases := []struct {
			tenantOpts          domain.TenantOpts
			name                string
			width               int
			height              int
			format              domain.ImageType
			toBeLoadedImageName string
		}{
			{
				tenantOpts:          domain.TenantOpts{TenantCode: "umoitj93", OrgCode: "ownlqz"},
				name:                "kjjoidj",
				format:              domain.ImageType_AVIF,
				width:               500,
				height:              300,
				toBeLoadedImageName: "tstimg1.jpeg",
			},
		}

		liss := NewLocalImageStorageService(testEnvironBaseDir)

		for _, tc := range testCases {
			image, err := os.ReadFile(filepath.Join(testdataDir, tc.toBeLoadedImageName))
			if err != nil {
				t.Fatalf("error reading image from testDataDir: %v", err)
			}

			err = liss.StoreChildImage(image,
				tc.name,
				domain.ImageSpec{Width: tc.width, Height: tc.height, Format: tc.format},
				tc.tenantOpts,
			)

			assert.NoError(t, err)

			_, err = os.ReadFile(filepath.Join(
				childImageDir(
					parentImageDir(testEnvironBaseDir, tc.tenantOpts, tc.name),
					tc.format,
					tc.width,
					tc.height,
				),
				tc.name+"."+tc.format.String(),
			),
			)
			if errors.Is(err, os.ErrNotExist) {
				t.Errorf("expected an image to be found but got: %v", err)
			}
		}

	})
}

func TestGetChildImage(t *testing.T) {
	t.Run("an image can be fetched if it exists", func(t *testing.T) {
		if err := initTestEnvironment(); err != nil {
			t.Fatalf("error initializing test environment: %v", err)
		}
		defer func() {
			err := tearDownTestEnvironment()
			if err != nil {
				panic(err)
			}
		}()

		testCases := []struct {
			tenantOpts domain.TenantOpts
			name       string
			format     domain.ImageType
			width      int
			height     int
			storedPath string
		}{
			{
				tenantOpts: initTestEnvironStatus[0].tenantOpts,
				name:       initTestEnvironStatus[0].name,
				format:     initTestEnvironStatus[0].format,
				width:      initTestEnvironStatus[0].width,
				height:     initTestEnvironStatus[0].height,
				storedPath: initTestEnvironStatus[0].storedDir,
			},
		}

		liss := NewLocalImageStorageService(testEnvironBaseDir)

		for _, tc := range testCases {
			expectedImage, err := os.ReadFile(tc.storedPath)
			if err != nil {
				t.Fatalf("error while reading image: %v", err)
			}

			image, err := liss.GetChildImage(tc.name, tc.format, tc.width, tc.height, tc.tenantOpts)

			assert.NoError(t, err)
			assert.Equal(t, expectedImage, image)
		}
	})

	t.Run("an error should be returned if image does not exists", func(t *testing.T) {
		if err := initTestEnvironment(); err != nil {
			t.Fatalf("error initializing test environment: %v", err)
		}
		defer func() {
			err := tearDownTestEnvironment()
			if err != nil {
				panic(err)
			}
		}()

		testCases := []struct {
			tenantOpts domain.TenantOpts
			name       string
			format     domain.ImageType
			width      int
			height     int
		}{
			{
				tenantOpts: initTestEnvironStatus[0].tenantOpts,
				name:       initTestEnvironStatus[0].name,
				format:     initTestEnvironStatus[0].format,
				width:      initTestEnvironStatus[0].width,
				height:     800,
			},
			{
				tenantOpts: domain.TenantOpts{TenantCode: "gihoijg", OrgCode: "gheg93"},
				name:       "gjizoqzgj03",
				format:     domain.ImageType_WEBP,
				width:      333,
				height:     3939,
			},
		}

		liss := NewLocalImageStorageService(testEnvironBaseDir)

		for _, tc := range testCases {
			_, err := liss.GetChildImage(tc.name, tc.format, tc.width, tc.height, tc.tenantOpts)

			assert.ErrorIs(t, err, ErrNoMatchingFile)
		}
	})
}

func TestGetParentImage(t *testing.T) {
	t.Run("an image can be fetched if it exists", func(t *testing.T) {
		if err := initTestEnvironment(); err != nil {
			t.Fatalf("error initializing test environment: %v", err)
		}
		defer func() {
			err := tearDownTestEnvironment()
			if err != nil {
				panic(err)
			}
		}()

		testCases := []struct {
			tenantOpts domain.TenantOpts
			name       string
			storedPath string
		}{
			{
				tenantOpts: initTestEnvironStatus[1].tenantOpts,
				name:       initTestEnvironStatus[1].name,
				storedPath: initTestEnvironStatus[1].storedDir,
			},
		}

		liss := NewLocalImageStorageService(testEnvironBaseDir)

		for _, tc := range testCases {
			expectedImage, err := os.ReadFile(tc.storedPath)
			if err != nil {
				t.Fatalf("error while reading image: %v", err)
			}

			image, err := liss.GetParentImage(tc.name, tc.tenantOpts)

			assert.NoError(t, err)
			assert.Equal(t, expectedImage, image)
		}
	})

	t.Run("an error should be returned if image does not exist", func(t *testing.T) {
		if err := initTestEnvironment(); err != nil {
			t.Fatalf("error initializing test environment: %v", err)
		}
		defer func() {
			err := tearDownTestEnvironment()
			if err != nil {
				panic(err)
			}
		}()

		testCases := []struct {
			tenantOpts domain.TenantOpts
			name       string
		}{
			{
				tenantOpts: domain.TenantOpts{TenantCode: "qzxxo", OrgCode: "owwmc"},
				name:       "eeeieiw",
			},
		}

		liss := NewLocalImageStorageService(testEnvironBaseDir)

		for _, tc := range testCases {
			_, err := liss.GetParentImage(tc.name, tc.tenantOpts)

			assert.ErrorIs(t, err, ErrNoMatchingFile)
		}
	})

	t.Run("an internal error should be returned if image does not exist unexpectedly", func(t *testing.T) {
		if err := initTestEnvironment(); err != nil {
			t.Fatalf("error initializing test environment: %v", err)
		}
		defer func() {
			err := tearDownTestEnvironment()
			if err != nil {
				panic(err)
			}
		}()

		testCases := []struct {
			tenantOpts domain.TenantOpts
			name       string
		}{
			{
				tenantOpts: domain.TenantOpts{TenantCode: "tenant1", OrgCode: "org1"},
				name:       "initTestEnviron1",
			},
		}

		liss := NewLocalImageStorageService(testEnvironBaseDir)

		for _, tc := range testCases {
			_, err := liss.GetParentImage(tc.name, tc.tenantOpts)

			assert.Error(t, err, ErrNoMatchingFile)
		}
	})
}
