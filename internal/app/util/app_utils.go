package util

import (
	"errors"
	"example.com/imageProc/internal/domain"
	"fmt"
	"os"
	"strings"
	"time"
)

type ChildPathOpts struct {
	ImgType  domain.ImageType
	ImgAR    domain.AR
	ImgWidth int
}

var (
	ErrNoMatchingFile   = errors.New("no file found with in the directory with the given pattern")
	ErrPathDoesNotExist = errors.New("path does not exist")
)

func ResolveStoragePath(baseUrl string, tenantOpts domain.TenantOpts, name string, isParent bool,
	childImageOpts ChildPathOpts) string {

	parentDir := parentImageDir(baseUrl, tenantOpts, name)
	if isParent {
		return parentDir
	}
	return fmt.Sprintf("%s/%s/%s/%d", parentDir, childImageOpts.ImgType.String(), childImageOpts.ImgAR.String(), childImageOpts.ImgWidth)
}

func parentImageDir(baseUrl string, tenantOpts domain.TenantOpts, name string) string {
	return fmt.Sprintf("%s/%s-%s/%s", baseUrl, tenantOpts.TenantCode, tenantOpts.OrgCode, name)
}

func FullImageAddr(pathToImg string, name string, fileExt string) string {
	return pathToImg + "/" + name + "." + fileExt
}

func FindImage(dir string, pattern string) (string, error) {
	dirEntry, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrPathDoesNotExist
		}
		return "", err
	}

	isFound := false
	var foundFileName string
	for _, e := range dirEntry {
		if strings.Contains(e.Name(), pattern) {
			isFound = true
			foundFileName = e.Name()
			break
		}
	}
	if isFound {
		return foundFileName, nil
	}
	return "", ErrNoMatchingFile
}

func GenerateImageName() string {
	return fmt.Sprintf("%d", time.Now().Nanosecond())
}
