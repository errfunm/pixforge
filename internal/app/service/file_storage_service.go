package appsvc

import (
	"errors"
	"example.com/imageProc/internal/domain"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrNoMatchingFile = errors.New("no file found with in the directory with the given pattern")
	ErrInternal       = errors.New("internal error")
)

type ImageStorageServiceInterface interface {
	StoreParentImage(image []byte, format domain.ImageType, tenantOpts domain.TenantOpts) (string, error)
	StoreChildImage(image []byte, name string, spec domain.ImageSpec, tenantOpts domain.TenantOpts) error
	GetParentImage(name string, tenantOpts domain.TenantOpts) ([]byte, error)
	GetChildImage(name string, format domain.ImageType, width, height int, tenantOpts domain.TenantOpts) ([]byte, error)
}

type localImageStorageService struct {
	baseDir string
}

func (l localImageStorageService) StoreParentImage(image []byte, format domain.ImageType, tenantOpts domain.TenantOpts) (string, error) {
	fName := GenerateImageName()

	path := parentImageDir(l.baseDir, tenantOpts, fName)

	if err := os.MkdirAll(path, 0750); err != nil {
		return "", fmt.Errorf("error while making directory %s", err.Error())
	}

	fDir := filepath.Join(path, fName+"."+format.String())
	if err := os.WriteFile(fDir, image, 0666); err != nil {
		return "", fmt.Errorf("error while writing file %s", err.Error())
	}
	return fName, nil
}

func (l localImageStorageService) StoreChildImage(image []byte, name string, spec domain.ImageSpec, tenantOpts domain.TenantOpts) error {
	path := childImageDir(
		parentImageDir(l.baseDir, tenantOpts, name),
		spec.Format, spec.Width, spec.Height)

	if err := os.MkdirAll(path, 0750); err != nil {
		return fmt.Errorf("error while making directory %s", err.Error())
	}

	fDir := filepath.Join(path, name+"."+spec.Format.String())

	if err := os.WriteFile(fDir, image, 0666); err != nil {
		return fmt.Errorf("error while writing file %s", err.Error())
	}
	return nil
}

func (l localImageStorageService) GetParentImage(name string, tenantOpts domain.TenantOpts) ([]byte, error) {
	path := parentImageDir(l.baseDir, tenantOpts, name)

	dirEntry, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoMatchingFile
		}
		return nil, err
	}

	found := false
	var parentImageFileName string
	for _, e := range dirEntry {
		if strings.Contains(e.Name(), name+".") {
			found = true
			parentImageFileName = e.Name()
			break
		}
	}
	if !found {
		return nil, ErrInternal
	}

	fDir := filepath.Join(path, parentImageFileName)

	image, err := os.ReadFile(fDir)
	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return image, nil
}

func (l localImageStorageService) GetChildImage(name string, format domain.ImageType, width, height int, tenantOpts domain.TenantOpts) ([]byte, error) {
	path := childImageDir(
		parentImageDir(l.baseDir, tenantOpts, name),
		format, width, height,
	)

	fDir := filepath.Join(path, name+"."+format.String())

	image, err := os.ReadFile(fDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoMatchingFile
		}
		return nil, fmt.Errorf("internal error: %v", err)
	}
	return image, nil
}

func NewLocalImageStorageService(baseDir string) ImageStorageServiceInterface {
	return localImageStorageService{
		baseDir: baseDir,
	}
}

func parentImageDir(baseUrl string, tenantOpts domain.TenantOpts, name string) string {
	return fmt.Sprintf("%s/%s-%s/%s", baseUrl, tenantOpts.TenantCode, tenantOpts.OrgCode, name)
}

func childImageDir(parentDir string, format domain.ImageType, width, height int) string {
	return fmt.Sprintf("%s/%s/%d/%d", parentDir, format, width, height)
}

func GenerateImageName() string {
	return fmt.Sprintf("%d", time.Now().Nanosecond())
}
