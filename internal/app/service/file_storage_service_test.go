package appsvc

import (
	"example.com/imageProc/internal/domain"
	"github.com/stretchr/testify/assert"
	"testing"
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
