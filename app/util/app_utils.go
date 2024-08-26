package util

import (
	"example.com/imageProc/domain"
	"fmt"
	"time"
)

type ChildPathOpts struct {
	ImgType  domain.ImageType
	ImgAR    domain.AR
	ImgWidth int
}

func ResolveStoragePath(baseUrl string, tenantOpts domain.TenantOpts, name string, isParent bool,
	childImageOpts ChildPathOpts) string {

	parentDir := parentImageDir(baseUrl, tenantOpts, name)
	if isParent {
		return parentDir
	}
	return fmt.Sprintf("%s/%s/%s/%d", parentDir, childImageOpts.ImgType.String(), childImageOpts.ImgAR.String(), childImageOpts.ImgWidth)
}

// parentImageDir || imageParentDir
func parentImageDir(baseUrl string, tenantOpts domain.TenantOpts, name string) string {
	return fmt.Sprintf("%s/%s-%s/%s", baseUrl, tenantOpts.TenantCode, tenantOpts.OrgCode, name)
}

func FullImageAddr(pathToImg string, name string, fileExt string) string {
	return pathToImg + "/" + name + "." + fileExt
}

func GenerateImageName() string {
	return string(rune(time.Now().Nanosecond()))
}
