package util

import (
	"example.com/imageProc/domain"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCalculateImageFullPath(t *testing.T) {
	type imageAttr struct {
		tenantOpts domain.TenantOpts
		baseURL    string
		imgName    string
		isParent   bool
		imgType    domain.ImageType
		imgAR      domain.AR
		imgWidth   int
	}
	testCases := []struct {
		in       imageAttr
		expected string
	}{
		{imageAttr{
			tenantOpts: domain.TenantOpts{TenantCode: "tsttnt", OrgCode: "tstorg"},
			baseURL:    "test",
			imgName:    "kfdkllak",
			isParent:   false,
			imgType:    domain.ImageType_JPEG,
			imgAR:      domain.AR{Width: 3, Height: 4},
			imgWidth:   300},
			"test/tsttnt-tstorg/kfdkllak/jpg/3:4/300"},

		{imageAttr{
			tenantOpts: domain.TenantOpts{TenantCode: "tsttnt", OrgCode: "tstorg"},
			baseURL:    "test",
			imgName:    "kjskldjd",
			isParent:   true,
			imgType:    domain.ImageType_AVIF,
			imgAR:      domain.AR{Width: 16, Height: 9},
			imgWidth:   400},
			"test/tsttnt-tstorg/kjskldjd"},
	}

	for _, tc := range testCases {
		res := ResolveStoragePath(tc.in.baseURL, tc.in.tenantOpts, tc.in.imgName, tc.in.isParent,
			ChildPathOpts{ImgType: tc.in.imgType, ImgAR: tc.in.imgAR, ImgWidth: tc.in.imgWidth})
		assert.Equal(t, tc.expected, res)
	}
}

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
	}

	for _, tc := range testCases {
		res := parentImageDir(tc.baseUrl, tc.tenantOpts, tc.name)
		assert.Equal(t, tc.expected, res)
	}
}

func TestFullImageAddr(t *testing.T) {
	testCases := []struct {
		path     string
		name     string
		fileExt  string
		expected string
	}{
		{
			path:     "/home/testusr/child-dir",
			name:     "tstimg",
			fileExt:  "jpeg",
			expected: "/home/testusr/child-dir/tstimg.jpeg",
		},
	}
	for _, tc := range testCases {
		res := FullImageAddr(tc.path, tc.name, tc.fileExt)
		assert.Equal(t, tc.expected, res)
	}
}

func TestGenerateImageName(t *testing.T) {
	nowNano := time.Now().Nanosecond()
	res := GenerateImageName()
	assert.Equal(t, string(rune(nowNano)), res)
}
